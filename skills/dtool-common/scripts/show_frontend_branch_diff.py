#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
显示当前分支相对指定基分支的“前端范围”改动文件及完整 diff 内容。

脚本不假设项目目录结构；默认按常见前端文件类型筛选改动，
并自动排除任意 dist 目录。也可通过参数追加更多排除路径。

用法:
python show_frontend_branch_diff.py <基分支> [额外排除路径1] [额外排除路径2] ...
"""

import subprocess
import sys

# Windows 中文环境下 stdout 默认编码为 GBK，导致 UTF-8 输出乱码。
sys.stdout.reconfigure(encoding='utf-8', errors='replace')
sys.stderr.reconfigure(encoding='utf-8', errors='replace')


def git_run(args: list[str]) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["git"] + args,
        capture_output=True,
        text=True,
        encoding="utf-8",
        errors="replace",
    )


def run_git(*args: str) -> str:
    result = git_run(list(args))
    if result.returncode != 0:
        msg = result.stderr.strip()
        print(f"ERROR: {msg}", file=sys.stderr)
        sys.exit(1)
    return result.stdout.strip()


def normalize_path(value: str) -> str:
    return value.strip().replace("\\", "/").strip("/")


def build_pathspecs(extra_excludes: list[str]) -> list[str]:
    pathspecs = [
        "*.js",
        "*.jsx",
        "*.ts",
        "*.tsx",
        "*.vue",
        "*.css",
        "*.scss",
        "*.sass",
        "*.less",
        "*.styl",
        "*.html",
        "*.htm",
        "*.json",
        "*.mjs",
        "*.cjs",
        "*.svg",
        "*.png",
        "*.jpg",
        "*.jpeg",
        "*.gif",
        "*.webp",
        "*.ico",
        "*.map",
        "package.json",
        "package-lock.json",
        "pnpm-lock.yaml",
        "yarn.lock",
        "npm-shrinkwrap.json",
        "tsconfig.json",
        "jsconfig.json",
        "vite.config.*",
        "webpack.config.*",
        "vue.config.*",
        "postcss.config.*",
        "tailwind.config.*",
        "babel.config.*",
        ".browserslistrc",
        ".npmrc",
        ".nvmrc",
        ":(exclude)**/dist/**",
    ]
    for path in extra_excludes:
        normalized = normalize_path(path)
        if normalized:
            pathspecs.append(f":(exclude){normalized}/**")
            pathspecs.append(f":(exclude){normalized}")
    return pathspecs


def main() -> int:
    if len(sys.argv) < 2:
        print(
            "用法: python show_frontend_branch_diff.py <基分支> [额外排除路径...]",
            file=sys.stderr,
        )
        sys.exit(1)

    base_branch = sys.argv[1].strip()
    extra_excludes = sys.argv[2:]

    if not base_branch:
        print("基分支不能为空", file=sys.stderr)
        sys.exit(1)

    try:
        run_git("rev-parse", "--show-toplevel")
    except SystemExit:
        print("当前目录不是 git 仓库", file=sys.stderr)
        sys.exit(1)

    try:
        run_git("rev-parse", "--verify", base_branch)
    except SystemExit:
        print(f"基分支 '{base_branch}' 不存在", file=sys.stderr)
        sys.exit(1)

    merge_base = run_git("merge-base", base_branch, "HEAD")
    if not merge_base:
        print(f"无法计算 '{base_branch}' 与当前分支的 merge-base", file=sys.stderr)
        sys.exit(1)

    pathspecs = build_pathspecs(extra_excludes)

    # 收集改动文件：已提交 + 暂存区 + 工作区
    files = set()

    committed_names = git_run(["diff", "--name-only", merge_base, "HEAD", "--", *pathspecs])
    if committed_names.returncode == 0:
        files.update(f for f in committed_names.stdout.strip().splitlines() if f.strip())

    cached_names = git_run(["diff", "--name-only", "--cached", "--", *pathspecs])
    if cached_names.returncode == 0:
        files.update(f for f in cached_names.stdout.strip().splitlines() if f.strip())

    wt_names = git_run(["diff", "--name-only", "--", *pathspecs])
    if wt_names.returncode == 0:
        files.update(f for f in wt_names.stdout.strip().splitlines() if f.strip())

    if not files:
        print("当前分支没有匹配范围内的改动")
        return 0

    # 逐文件获取 diff（兼容未提交改动）
    diff_parts = []
    for f in sorted(files):
        # 优先已提交的 diff
        r = git_run(["diff", merge_base, "HEAD", "--", f])
        if r.returncode == 0 and r.stdout.strip():
            diff_parts.append(r.stdout)
            continue
        # 暂存区 diff
        r = git_run(["diff", "--cached", "--", f])
        if r.returncode == 0 and r.stdout.strip():
            diff_parts.append(r.stdout)
            continue
        # 工作区 diff
        r = git_run(["diff", "--", f])
        if r.returncode == 0 and r.stdout.strip():
            diff_parts.append(r.stdout)

    combined = "\n".join(diff_parts)
    sys.stdout.buffer.write(combined.encode("utf-8", errors="replace"))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
