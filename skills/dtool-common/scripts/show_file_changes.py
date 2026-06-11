#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
显示本地工作目录的文件变更详情（JSON 格式输出）

用法:
  # 获取文件变更汇总（分类统计 + 文件列表）
  python show_file_changes.py <local_dir> [parent_branch]

  # 获取变更汇总 + 所有文件的 diff（慢，文件多时不推荐）
  python show_file_changes.py <local_dir> <parent_branch> --with-diffs

  # 获取单个文件的 diff
  python show_file_changes.py <local_dir> <parent_branch> --file <file_path>

输出均为 JSON 格式，格式说明：
  - summary: { added: N, modified: N, deleted: N, renamed: N, untracked: N, total: N }
  - files: [{ path: "xxx", type: "modified"/"added"/"deleted"/"renamed"/"untracked"/"other", status_code: " M" }]
  - diffs (--with-diffs 时): { "relative/path": "diff text..." }
  - diff (--file 时): "unified diff text"
"""

import json
import os
import subprocess
import sys
import traceback

# Windows 中文环境下 stdout 默认编码为 GBK，导致 UTF-8 输出乱码。
sys.stdout.reconfigure(encoding='utf-8', errors='replace')
sys.stderr.reconfigure(encoding='utf-8', errors='replace')


def run_git(local_dir: str, *args: str) -> str:
    """在指定目录执行 git 命令，返回 stdout（去除末尾换行）。失败时抛异常。"""
    cmd = ["git", "-C", local_dir] + list(args)
    result = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", errors="replace")
    if result.returncode != 0:
        msg = result.stderr.strip() or f"git {' '.join(args)} failed with code {result.returncode}"
        raise RuntimeError(msg)
    return result.stdout.strip()


def is_git_repo(local_dir: str) -> bool:
    try:
        run_git(local_dir, "rev-parse", "--show-toplevel")
        return True
    except Exception:
        return False


def categorize_status(status_code: str) -> str:
    """
    根据 git status --short 的状态码分类：
      ??  -> untracked
      A   -> added
      M   -> modified
      D   -> deleted
      R   -> renamed
      C   -> copied
      其他 -> other
    """
    if not status_code:
        return "other"
    code = status_code.strip()
    if code == "??":
        return "untracked"
    # 索引区（staged）的状态码
    idx = code[:1] if len(code) >= 2 else code
    wt = code[1:2] if len(code) >= 2 else ""
    if idx == "A" or wt == "A":
        return "added"
    if idx == "D" or wt == "D":
        return "deleted"
    if idx == "R" or wt == "R":
        return "renamed"
    if idx == "C" or wt == "C":
        return "copied"
    if idx == "M" or wt == "M":
        return "modified"
    return "other"


def extract_file_path(status_line: str) -> str:
    """
    从 git status --short 的行中提取文件路径。
    格式: "XY path" 或 "XY orig -> new"（重命名）
    """
    line = status_line.rstrip("\r\n")
    if len(line) < 3:
        return line.strip()
    code_and_space = line[:3]  # e.g. " M " or "?? "
    rest = line[3:].strip()
    # 处理重命名 "R  old -> new"
    if " -> " in rest:
        return rest.split(" -> ")[-1].strip()
    return rest


def get_git_diff(local_dir: str, merge_base: str, file_path: str = None) -> str:
    """获取 git diff（相对于 merge_base），可选限定文件路径。包含暂存区和工作区改动。"""
    path_args = ["--", file_path] if file_path else ["--", ".", ":(exclude)**/dist/**"]

    # 1) 已提交的 diff
    args = ["diff", merge_base, "HEAD"] + path_args
    result = subprocess.run(
        ["git", "-C", local_dir] + args,
        capture_output=True, text=True, encoding="utf-8", errors="replace",
    )
    if result.returncode == 0 and result.stdout.strip():
        return result.stdout

    # 2) 暂存区 diff
    args = ["diff", "--cached"] + path_args
    result = subprocess.run(
        ["git", "-C", local_dir] + args,
        capture_output=True, text=True, encoding="utf-8", errors="replace",
    )
    if result.returncode == 0 and result.stdout.strip():
        return result.stdout

    # 3) 工作区 diff
    args = ["diff"] + path_args
    result = subprocess.run(
        ["git", "-C", local_dir] + args,
        capture_output=True, text=True, encoding="utf-8", errors="replace",
    )
    if result.returncode != 0:
        msg = result.stderr.strip() or f"git diff failed"
        raise RuntimeError(msg)
    return result.stdout


def get_all_files_diff(local_dir: str, merge_base: str) -> dict:
    """
    逐文件获取 diff，返回 { "relative/path": "diff text" }。
    包含已提交、暂存区和工作区的改动。
    """
    exclude = ["--", ".", ":(exclude)**/dist/**"]
    files = set()

    # 已提交
    committed = run_git(local_dir, "diff", "--name-only", merge_base, "HEAD", *exclude)
    files.update(f.strip() for f in committed.splitlines() if f.strip())

    # 暂存区
    cached = run_git(local_dir, "diff", "--name-only", "--cached", *exclude)
    files.update(f.strip() for f in cached.splitlines() if f.strip())

    # 工作区
    wt = run_git(local_dir, "diff", "--name-only", *exclude)
    files.update(f.strip() for f in wt.splitlines() if f.strip())

    diffs = {}
    for f in sorted(files):
        try:
            diffs[f] = get_git_diff(local_dir, merge_base, f)
        except Exception as e:
            diffs[f] = f"<!-- ERROR: {e} -->"
    return diffs


def get_file_diff(local_dir: str, merge_base: str, file_path: str) -> str:
    """获取单个文件的 diff。"""
    return get_git_diff(local_dir, merge_base, file_path)


def output_json(data):
    """安全输出 JSON 到 stdout，出错时输出错误 JSON。"""
    try:
        json.dump(data, sys.stdout, ensure_ascii=False, indent=2)
        sys.stdout.write("\n")
        sys.stdout.flush()
    except Exception as e:
        # fallback
        json.dump({"error": str(e)}, sys.stdout, ensure_ascii=False)
        sys.stdout.write("\n")
        sys.stdout.flush()


def output_error_json(msg: str):
    output_json({"error": msg})
    sys.exit(1)


def main() -> int:
    args = sys.argv[1:]

    # 解析参数
    local_dir = ""
    parent_branch = ""
    mode = "summary"  # summary / diffs / file-diff
    target_file = ""

    i = 0
    while i < len(args):
        arg = args[i]
        if arg == "--with-diffs":
            mode = "diffs"
        elif arg == "--file":
            mode = "file-diff"
            i += 1
            if i < len(args):
                target_file = args[i]
        elif not local_dir:
            local_dir = arg
        elif not parent_branch:
            parent_branch = arg
        i += 1

    if not local_dir:
        output_error_json("缺少参数: local_dir")

    local_dir = os.path.abspath(local_dir)

    if not os.path.isdir(local_dir):
        output_error_json(f"目录不存在: {local_dir}")

    if not is_git_repo(local_dir):
        output_error_json(f"目录不是 git 仓库: {local_dir}")

    try:
        # 获取 git status --short
        raw_status = run_git(local_dir, "status", "--short")
    except Exception as e:
        output_error_json(f"获取 git status 失败: {e}")

    # 解析 status 输出
    summary = {"added": 0, "modified": 0, "deleted": 0, "renamed": 0, "untracked": 0, "other": 0, "total": 0}
    files = []

    if raw_status:
        for line in raw_status.splitlines():
            line = line.rstrip("\r")
            if not line.strip():
                continue
            status_code = line[:2] if len(line) >= 2 else line
            cat = categorize_status(status_code)
            file_path = extract_file_path(line)
            summary[cat] = summary.get(cat, 0) + 1
            summary["total"] += 1
            files.append({
                "path": file_path,
                "type": cat,
                "status_code": status_code,
            })

    result = {
        "local_dir": local_dir,
        "summary": summary,
        "files": files,
    }

    # 如果需要获取 diff
    if mode in ("diffs", "file-diff") and parent_branch:
        try:
            merge_base = run_git(local_dir, "merge-base", parent_branch, "HEAD")
            if not merge_base:
                result["diff_error"] = f"无法计算 '{parent_branch}' 与 HEAD 的 merge-base"
            elif mode == "file-diff":
                if not target_file:
                    result["diff_error"] = "缺少 --file 参数"
                else:
                    result["diff"] = get_file_diff(local_dir, merge_base, target_file)
            else:
                result["diffs"] = get_all_files_diff(local_dir, merge_base)
        except Exception as e:
            result["diff_error"] = str(e)

    output_json(result)
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception:
        output_error_json(traceback.format_exc())
        sys.exit(1)
