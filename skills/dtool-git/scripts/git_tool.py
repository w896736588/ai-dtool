#!/usr/bin/env python3
"""Structured Git operations for AI agents."""

from __future__ import annotations

import argparse
import base64
import difflib
import json
import os
from pathlib import Path
import subprocess
import sys
from typing import Any
from urllib import error, request


DEFAULT_MAX_DIFF_BYTES = 200_000
DEFAULT_MAX_FILES = 500
DEFAULT_CONTEXT_LINES = 3


class ToolError(RuntimeError):
    def __init__(self, code: str, message: str):
        super().__init__(message)
        self.code = code


def emit(data: dict[str, Any], exit_code: int = 0) -> None:
    json.dump(data, sys.stdout, ensure_ascii=False, indent=2)
    sys.stdout.write("\n")
    raise SystemExit(exit_code)


def run_git(repo: Path, args: list[str], *, check: bool = True) -> bytes:
    result = subprocess.run(
        ["git", "-C", str(repo), *args],
        capture_output=True,
    )
    if check and result.returncode != 0:
        message = result.stderr.decode("utf-8", errors="replace").strip()
        raise ToolError("GIT_COMMAND_FAILED", message or f"git {' '.join(args)} failed")
    return result.stdout


def text(value: bytes) -> str:
    return value.decode("utf-8", errors="replace")


def repo_root(local_dir: str) -> Path:
    candidate = Path(local_dir or os.getcwd()).resolve()
    if not candidate.is_dir():
        raise ToolError("LOCAL_DIR_NOT_FOUND", f"目录不存在: {candidate}")
    result = subprocess.run(
        ["git", "-C", str(candidate), "rev-parse", "--show-toplevel"],
        capture_output=True,
    )
    if result.returncode != 0:
        raise ToolError("NOT_A_GIT_REPOSITORY", f"目录不是 Git 仓库: {candidate}")
    return Path(text(result.stdout).strip()).resolve()


def current_branch(repo: Path) -> str:
    name = text(run_git(repo, ["symbolic-ref", "--short", "-q", "HEAD"], check=False)).strip()
    return name or "HEAD"


def verify_compare_branch(repo: Path, compare_branch: str) -> str:
    branch = compare_branch.strip()
    if not branch:
        raise ToolError("COMPARE_BRANCH_REQUIRED", "branch 模式必须指定 compare_branch")
    result = run_git(
        repo,
        ["rev-parse", "--verify", "--end-of-options", f"{branch}^{{commit}}"],
        check=False,
    )
    commit = text(result).strip()
    if not commit:
        raise ToolError("COMPARE_BRANCH_NOT_FOUND", f"对比分支不存在: {branch}")
    return commit


def normalize_repo_path(repo: Path, file_path: str) -> str:
    raw = file_path.strip().replace("\\", "/")
    if not raw:
        raise ToolError("FILE_PATH_REQUIRED", "file_path 不能为空")
    candidate = Path(raw)
    resolved = candidate.resolve() if candidate.is_absolute() else (repo / candidate).resolve()
    try:
        relative = resolved.relative_to(repo)
    except ValueError as exc:
        raise ToolError("FILE_OUTSIDE_REPOSITORY", f"文件不在仓库内: {file_path}") from exc
    return relative.as_posix()


def normalize_excludes(values: list[str]) -> list[str]:
    result: list[str] = []
    for value in values:
        normalized = value.strip().replace("\\", "/").strip("/")
        if not normalized:
            continue
        if normalized == ".." or normalized.startswith("../") or ":" in normalized:
            raise ToolError("INVALID_EXCLUDE", f"排除路径必须是仓库相对路径: {value}")
        result.extend([f":(exclude){normalized}", f":(exclude){normalized}/**"])
    return result


def scope_pathspecs(scope: str, file_path: str = "", extra_excludes: list[str] | None = None) -> list[str]:
    if file_path:
        if scope == "frontend" and not (file_path == "web" or file_path.startswith("web/")):
            raise ToolError("FILE_OUTSIDE_SCOPE", f"文件不属于 frontend 范围: {file_path}")
        if scope == "backend" and (file_path == "web" or file_path.startswith("web/")):
            raise ToolError("FILE_OUTSIDE_SCOPE", f"文件不属于 backend 范围: {file_path}")
        pathspecs = [f":(literal){file_path}"]
    elif scope == "frontend":
        pathspecs = ["web/**", ":(exclude)web/dist/**", ":(exclude)web/node_modules/**"]
    elif scope == "backend":
        pathspecs = [".", ":(exclude)web/**", ":(exclude)**/dist/**"]
    elif scope == "all":
        pathspecs = [".", ":(exclude)**/dist/**", ":(exclude)web/node_modules/**"]
    else:
        raise ToolError("INVALID_SCOPE", f"不支持的 scope: {scope}")
    return [*pathspecs, *normalize_excludes(extra_excludes or [])]


def comparison(repo: Path, args: argparse.Namespace) -> dict[str, Any]:
    mode = args.compare_mode
    if mode == "workspace":
        if args.compare_branch:
            raise ToolError("COMPARE_BRANCH_NOT_ALLOWED", "workspace 模式不能指定 compare_branch")
        run_git(repo, ["rev-parse", "--verify", "HEAD"])
        return {
            "mode": mode,
            "compare_branch": None,
            "compare_strategy": None,
            "merge_base": None,
            "diff_range": ["HEAD"],
            "include_untracked": True,
        }

    compare_commit = verify_compare_branch(repo, args.compare_branch)
    if args.compare_strategy == "tree":
        start = compare_commit
        merge_base = None
    else:
        start = text(run_git(repo, ["merge-base", compare_commit, "HEAD"])).strip()
        if not start:
            raise ToolError("MERGE_BASE_NOT_FOUND", f"无法计算 {args.compare_branch} 与 HEAD 的 merge-base")
        merge_base = start
    diff_range = [start] if args.include_workspace else [start, "HEAD"]
    return {
        "mode": mode,
        "compare_branch": args.compare_branch,
        "compare_strategy": args.compare_strategy,
        "merge_base": merge_base,
        "diff_range": diff_range,
        "include_untracked": bool(args.include_workspace),
    }


def parse_name_status(raw: bytes) -> list[dict[str, Any]]:
    tokens = raw.split(b"\0")
    files: list[dict[str, Any]] = []
    index = 0
    while index < len(tokens) and tokens[index]:
        status_token = text(tokens[index])
        index += 1
        status = status_token[:1]
        if status in {"R", "C"}:
            if index + 1 >= len(tokens):
                break
            old_path = text(tokens[index])
            path = text(tokens[index + 1])
            index += 2
        else:
            if index >= len(tokens):
                break
            old_path = None
            path = text(tokens[index])
            index += 1
        files.append({
            "path": path,
            "old_path": old_path,
            "status": status,
            "similarity": status_token[1:] or None,
        })
    return files


def parse_numstat(raw: bytes) -> dict[str, tuple[int | None, int | None]]:
    tokens = raw.split(b"\0")
    result: dict[str, tuple[int | None, int | None]] = {}
    index = 0
    while index < len(tokens) and tokens[index]:
        parts = tokens[index].split(b"\t", 2)
        index += 1
        if len(parts) < 3:
            continue
        if parts[2]:
            path = text(parts[2])
        else:
            # With -z, rename/copy records store old and new paths in the next two tokens.
            if index + 1 >= len(tokens):
                break
            path = text(tokens[index + 1])
            index += 2
        if parts[0] == b"-" or parts[1] == b"-":
            result[path] = (None, None)
            continue
        try:
            result[path] = (int(parts[0]), int(parts[1]))
        except ValueError:
            result[path] = (0, 0)
    return result


def is_binary_file(path: Path) -> bool:
    try:
        return b"\0" in path.read_bytes()[:8192]
    except OSError:
        return False


def untracked_files(repo: Path, pathspecs: list[str]) -> list[dict[str, Any]]:
    raw = run_git(repo, ["ls-files", "-z", "--others", "--exclude-standard", "--", *pathspecs])
    result: list[dict[str, Any]] = []
    for item in raw.split(b"\0"):
        if not item:
            continue
        path = text(item)
        absolute = repo / path
        try:
            size = absolute.stat().st_size
        except OSError:
            size = 0
        additions: int | None = None
        if not is_binary_file(absolute):
            try:
                additions = len(absolute.read_text(encoding="utf-8", errors="replace").splitlines())
            except OSError:
                additions = 0
        result.append({
            "path": path,
            "old_path": None,
            "status": "?",
            "similarity": None,
            "additions": additions,
            "deletions": 0 if additions is not None else None,
            "binary": additions is None,
            "size": size,
            "source": "workspace",
            "layers": ["untracked"],
        })
    return result


def workspace_layers(repo: Path) -> dict[str, list[str]]:
    raw = run_git(repo, ["status", "--porcelain=v1", "-z", "--untracked-files=all"])
    tokens = raw.split(b"\0")
    result: dict[str, list[str]] = {}
    index = 0
    while index < len(tokens) and tokens[index]:
        record = tokens[index]
        index += 1
        if len(record) < 4:
            continue
        code = text(record[:2])
        path = text(record[3:])
        if code[:1] in {"R", "C"} and index < len(tokens):
            index += 1  # porcelain -z stores the old path after the destination path
        layers: list[str] = []
        if code == "??":
            layers.append("untracked")
        else:
            if code[0] != " ":
                layers.append("staged")
            if code[1] != " ":
                layers.append("modified")
        result[path] = layers
    return result


def collect_changes(repo: Path, args: argparse.Namespace) -> tuple[dict[str, Any], list[dict[str, Any]]]:
    comp = comparison(repo, args)
    file_path = normalize_repo_path(repo, args.file_path) if args.file_path else ""
    pathspecs = scope_pathspecs(args.scope, file_path, args.exclude)
    raw = run_git(
        repo,
        ["diff", "--name-status", "-z", "--find-renames", *comp["diff_range"], "--", *pathspecs],
    )
    files = parse_name_status(raw)
    stats = parse_numstat(run_git(
        repo,
        ["diff", "--numstat", "-z", "--find-renames", *comp["diff_range"], "--", *pathspecs],
    ))
    layer_map = workspace_layers(repo) if comp["include_untracked"] else {}
    for item in files:
        additions, deletions = stats.get(item["path"], (0, 0))
        item.update({
            "additions": additions,
            "deletions": deletions,
            "binary": additions is None,
            "source": comp["mode"] if comp["mode"] == "workspace" else "branch",
            "layers": layer_map.get(item["path"], []),
        })
    if comp["include_untracked"] and args.include_untracked:
        tracked_paths = {item["path"] for item in files}
        files.extend(item for item in untracked_files(repo, pathspecs) if item["path"] not in tracked_paths)
    files.sort(key=lambda item: item["path"])
    return comp, files


def summary(files: list[dict[str, Any]]) -> dict[str, Any]:
    additions = sum(item["additions"] or 0 for item in files)
    deletions = sum(item["deletions"] or 0 for item in files)
    return {
        "total": len(files),
        "added": sum(item["status"] in {"A", "?"} for item in files),
        "modified": sum(item["status"] == "M" for item in files),
        "deleted": sum(item["status"] == "D" for item in files),
        "renamed": sum(item["status"] == "R" for item in files),
        "copied": sum(item["status"] == "C" for item in files),
        "binary": sum(bool(item["binary"]) for item in files),
        "additions": additions,
        "deletions": deletions,
    }


def base_result(repo: Path, args: argparse.Namespace, comp: dict[str, Any]) -> dict[str, Any]:
    return {
        "ok": True,
        "action": args.action,
        "local_dir": str(repo),
        "current_branch": current_branch(repo),
        "compare_mode": comp["mode"],
        "compare_branch": comp["compare_branch"],
        "compare_strategy": comp["compare_strategy"],
        "merge_base": comp["merge_base"],
        "include_workspace": comp["mode"] == "workspace" or bool(args.include_workspace),
        "scope": args.scope,
    }


def action_changes(args: argparse.Namespace) -> dict[str, Any]:
    repo = repo_root(args.local_dir)
    comp, files = collect_changes(repo, args)
    result = base_result(repo, args, comp)
    visible_files = files[:args.max_files]
    result.update({
        "summary": summary(files),
        "files": visible_files,
        "files_truncated": len(visible_files) < len(files),
        "omitted_files": max(0, len(files) - len(visible_files)),
    })
    return result


def untracked_patch(repo: Path, item: dict[str, Any], context_lines: int) -> str:
    path = item["path"]
    absolute = repo / path
    if item["binary"]:
        return f"diff --git a/{path} b/{path}\nnew file mode 100644\nBinary file b/{path} added\n"
    try:
        new_lines = absolute.read_text(encoding="utf-8", errors="replace").splitlines(keepends=True)
    except OSError:
        return ""
    return "".join(difflib.unified_diff(
        [], new_lines, fromfile="/dev/null", tofile=f"b/{path}", n=context_lines,
    ))


def truncate_utf8(value: str, max_bytes: int) -> tuple[str, bool, int]:
    encoded = value.encode("utf-8")
    if len(encoded) <= max_bytes:
        return value, False, len(encoded)
    return encoded[:max_bytes].decode("utf-8", errors="ignore"), True, len(encoded)


def action_diff(args: argparse.Namespace) -> dict[str, Any]:
    repo = repo_root(args.local_dir)
    comp, files = collect_changes(repo, args)
    file_path = normalize_repo_path(repo, args.file_path) if args.file_path else ""
    pathspecs = scope_pathspecs(args.scope, file_path, args.exclude)
    patch = text(run_git(
        repo,
        ["diff", "--no-ext-diff", f"--unified={args.context_lines}", *comp["diff_range"], "--", *pathspecs],
    ))
    if comp["include_untracked"] and args.include_untracked:
        patch += "".join(untracked_patch(repo, item, args.context_lines) for item in files if item["status"] == "?")
    patch, truncated, original_bytes = truncate_utf8(patch, args.max_diff_bytes)
    result = base_result(repo, args, comp)
    result.update({
        "file_path": file_path or None,
        "summary": summary(files),
        "files": files[:args.max_files],
        "files_truncated": len(files) > args.max_files,
        "omitted_files": max(0, len(files) - args.max_files),
        "diff": patch,
        "truncated": truncated,
        "diff_bytes": len(patch.encode("utf-8")),
        "original_diff_bytes": original_bytes,
    })
    if file_path and args.include_content:
        old_ref = comp["diff_range"][0]
        old_path = next((item.get("old_path") for item in files if item["path"] == file_path and item.get("old_path")), file_path)
        old_bytes = run_git(repo, ["show", f"{old_ref}:{old_path}"], check=False)
        if comp["mode"] == "workspace" or args.include_workspace:
            try:
                new_bytes = (repo / file_path).read_bytes()
            except OSError:
                new_bytes = b""
        else:
            new_bytes = run_git(repo, ["show", f"HEAD:{file_path}"], check=False)
        binary = b"\0" in old_bytes[:8192] or b"\0" in new_bytes[:8192]
        image_type = Path(file_path).suffix.lower().lstrip(".")
        if image_type in {"png", "jpg", "jpeg", "gif", "webp", "bmp", "tiff", "tif", "ico"}:
            result.update({
                "is_image": True,
                "image_type": image_type,
                "old_image": base64.b64encode(old_bytes).decode("ascii"),
                "new_image": base64.b64encode(new_bytes).decode("ascii"),
            })
        elif binary:
            result.update({
                "is_binary": True,
                "file_type": Path(file_path).suffix.lower(),
                "old_size": len(old_bytes),
                "new_size": len(new_bytes),
            })
        else:
            result.update({
                "old_content": old_bytes.decode("utf-8", errors="replace"),
                "new_content": new_bytes.decode("utf-8", errors="replace"),
            })
    return result


def call_api(base_url: str, token: str, path: str, payload: dict[str, Any]) -> dict[str, Any]:
    url = base_url.rstrip("/") + path
    body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
    req = request.Request(
        url=url,
        data=body,
        headers={"Content-Type": "application/json; charset=utf-8", "Token": token},
        method="POST",
    )
    try:
        with request.urlopen(req, timeout=120) as response:
            result = json.loads(response.read().decode("utf-8"))
    except error.HTTPError as exc:
        message = exc.read().decode("utf-8", errors="replace")
        raise ToolError("HTTP_ERROR", f"HTTP {exc.code}: {message}") from exc
    except Exception as exc:
        raise ToolError("API_REQUEST_FAILED", str(exc)) from exc
    normalized = {
        "ok": result.get("ErrCode", result.get("code", -1)) == 0,
        "code": result.get("ErrCode", result.get("code")),
        "message": result.get("ErrMsg", result.get("msg", "")),
        "data": result.get("Data", result.get("data")),
    }
    if not normalized["ok"]:
        raise ToolError("REMOTE_OPERATION_FAILED", normalized["message"] or json.dumps(result, ensure_ascii=False))
    return normalized


REMOTE_PATHS = {
    "config_list": "/api/GitConfigList",
    "current_branch": "/api/GitCurrentBranch",
    "pull": "/api/GitPull",
    "change_branch": "/api/GitChangeBranchById",
    "status": "/api/GitQueryStatus",
    "commit_log": "/api/GitCommitLog",
    "change_remote_branch": "/api/GitChangeBranchRemote",
    "remote_branches": "/api/GitRemoteBranchList",
    "group_branches": "/api/GitGroupBranchList",
    "quick_create_branch": "/api/GitQuickCreateBranch",
    "set_safe": "/api/GitSetSafeLog",
    "save_credentials": "/api/GitSaveCredentials",
    "upload_file": "/api/GitUploadFile",
    "settings_list": "/api/Set/GitList",
    "config_add": "/api/Set/GitAdd",
    "config_delete": "/api/Set/GitDelete",
    "group_list": "/api/Set/GitGroupList",
    "group_add": "/api/Set/GitGroupAdd",
    "group_delete": "/api/Set/GitGroupDelete",
    "quick_list": "/api/Set/GitQuickList",
}

PASSTHROUGH_REMOTE_ACTIONS = {
    "settings_list", "config_add", "config_delete", "group_list", "group_add", "group_delete", "quick_list",
}


def repository_config(args: argparse.Namespace) -> dict[str, Any]:
    if not args.git_id:
        raise ToolError("GIT_ID_REQUIRED", f"{args.remote_action} 必须指定 git_id")
    response = call_api(args.base_url, args.token, REMOTE_PATHS["config_list"], {})
    repositories = (response.get("data") or {}).get("git_list") or []
    for item in repositories:
        if str(item.get("id")) == str(args.git_id):
            return dict(item)
    raise ToolError("GIT_CONFIG_NOT_FOUND", f"未找到 git_id={args.git_id} 的 Git 配置")


def action_remote(args: argparse.Namespace) -> dict[str, Any]:
    action = args.remote_action
    if action not in REMOTE_PATHS:
        raise ToolError("UNKNOWN_REMOTE_ACTION", f"不支持的远程 action: {action}")
    if action == "config_list":
        payload: dict[str, Any] = {}
    elif action in PASSTHROUGH_REMOTE_ACTIONS:
        payload = json.loads(args.payload_json)
        if not isinstance(payload, dict):
            raise ToolError("INVALID_PAYLOAD", "payload_json 必须是 JSON 对象")
    elif action == "group_branches":
        if not args.git_group_id:
            raise ToolError("GIT_GROUP_ID_REQUIRED", "group_branches 必须指定 git_group_id")
        payload = {"git_group_id": args.git_group_id}
    elif action == "upload_file":
        if not args.git_id:
            raise ToolError("GIT_ID_REQUIRED", "upload_file 必须指定 git_id")
        payload = {"git_id": args.git_id, "local_file_paths": json.loads(args.local_file_paths)}
    elif action in {"current_branch", "pull", "change_branch"}:
        if not args.git_id:
            raise ToolError("GIT_ID_REQUIRED", f"{action} 必须指定 git_id")
        payload = {"git_id": args.git_id}
        if action == "change_branch":
            if not args.branch_name:
                raise ToolError("BRANCH_NAME_REQUIRED", "change_branch 必须指定 branch_name")
            payload["branch_name"] = args.branch_name
        if action in {"pull", "change_branch"}:
            payload["discard_local_changes"] = bool(args.discard_local_changes)
    else:
        payload = repository_config(args)
        if action == "change_remote_branch":
            if not args.branch_name:
                raise ToolError("BRANCH_NAME_REQUIRED", "change_remote_branch 必须指定 branch_name")
            payload["BranchName"] = args.branch_name
            payload["discard_local_changes"] = bool(args.discard_local_changes)
        elif action == "quick_create_branch":
            if not args.base_branch or not args.branch_type or not args.business_en:
                raise ToolError("QUICK_BRANCH_PARAMS_REQUIRED", "快捷建分支需要 base_branch、branch_type、business_en")
            payload.update({
                "base_branch": args.base_branch,
                "branch_type": args.branch_type,
                "business_en": args.business_en,
                "discard_local_changes": bool(args.discard_local_changes),
            })
    response = call_api(args.base_url, args.token, REMOTE_PATHS[action], payload)
    return {"ok": True, "action": "remote", "remote_action": action, **response}


def parser() -> argparse.ArgumentParser:
    root = argparse.ArgumentParser(description="AI-oriented dtool Git operations")
    sub = root.add_subparsers(dest="action", required=True)

    for name in ("changes", "diff"):
        command = sub.add_parser(name)
        command.add_argument("--local-dir", default=os.getcwd())
        command.add_argument("--compare-mode", choices=("branch", "workspace"), default="branch")
        command.add_argument("--compare-branch", default="")
        command.add_argument("--compare-strategy", choices=("merge_base", "tree"), default="merge_base")
        command.add_argument("--include-workspace", action="store_true")
        command.add_argument("--include-untracked", action=argparse.BooleanOptionalAction, default=True)
        command.add_argument("--scope", choices=("all", "frontend", "backend"), default="all")
        command.add_argument("--file-path", default="")
        command.add_argument("--exclude", action="append", default=[])
        command.add_argument("--max-files", type=int, default=DEFAULT_MAX_FILES)
        if name == "diff":
            command.add_argument("--context-lines", type=int, default=DEFAULT_CONTEXT_LINES)
            command.add_argument("--max-diff-bytes", type=int, default=DEFAULT_MAX_DIFF_BYTES)
            command.add_argument("--include-content", action="store_true")

    remote = sub.add_parser("remote")
    remote.add_argument("remote_action", choices=tuple(REMOTE_PATHS))
    remote.add_argument("--base-url", default=os.environ.get("DTOOL_BASE_URL", "http://localhost:17170"))
    remote.add_argument("--token", default=os.environ.get("DTOOL_TOKEN", ""))
    remote.add_argument("--git-id", default="")
    remote.add_argument("--git-group-id", default="")
    remote.add_argument("--branch-name", default="")
    remote.add_argument("--base-branch", default="")
    remote.add_argument("--branch-type", choices=("feature", "hotfix"))
    remote.add_argument("--business-en", default="")
    remote.add_argument("--local-file-paths", default="[]")
    remote.add_argument("--payload-json", default="{}")
    remote.add_argument("--discard-local-changes", action="store_true")
    return root


def main() -> None:
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
    sys.stderr.reconfigure(encoding="utf-8", errors="replace")
    args = parser().parse_args()
    try:
        if args.action in {"changes", "diff"} and args.max_files < 1:
            raise ToolError("INVALID_LIMIT", "max_files 必须大于 0")
        if args.action == "diff" and (args.max_diff_bytes < 1 or args.context_lines < 0):
            raise ToolError("INVALID_LIMIT", "max_diff_bytes 必须大于 0，context_lines 不能小于 0")
        if args.action == "changes":
            result = action_changes(args)
        elif args.action == "diff":
            result = action_diff(args)
        else:
            result = action_remote(args)
        emit(result)
    except ToolError as exc:
        emit({"ok": False, "error": {"code": exc.code, "message": str(exc)}}, 1)
    except (json.JSONDecodeError, ValueError) as exc:
        emit({"ok": False, "error": {"code": "INVALID_ARGUMENT", "message": str(exc)}}, 1)


if __name__ == "__main__":
    main()
