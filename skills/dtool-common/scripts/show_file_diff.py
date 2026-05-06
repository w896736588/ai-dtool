#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
显示指定文件在当前分支中的改动内容（类似 GitLab MR 单文件 diff）

用法: python show_file_diff.py <基分支> <文件路径>
"""

import re
import subprocess
import sys


def run_git(*args: str) -> str:
    result = subprocess.run(["git"] + list(args), capture_output=True, text=True)
    if result.returncode != 0:
        msg = result.stderr.strip()
        print(f"ERROR: {msg}", file=sys.stderr)
        sys.exit(1)
    return result.stdout.strip()


def is_excluded_file(file_path: str) -> bool:
    normalized = file_path.replace("\\", "/")
    return bool(re.search(r"(^|/)dist/", normalized))


def main() -> int:
    if len(sys.argv) < 3:
        print("用法: python show_file_diff.py <基分支> <文件路径>", file=sys.stderr)
        sys.exit(1)

    base_branch = sys.argv[1].strip()
    file_path = sys.argv[2].strip()
    if not base_branch:
        print("基分支不能为空", file=sys.stderr)
        sys.exit(1)
    if not file_path:
        print("文件路径不能为空", file=sys.stderr)
        sys.exit(1)

    # 检查是否在 git 仓库中
    try:
        run_git("rev-parse", "--show-toplevel")
    except SystemExit:
        print("当前目录不是 git 仓库", file=sys.stderr)
        sys.exit(1)

    # 验证基分支存在
    try:
        run_git("rev-parse", "--verify", base_branch)
    except SystemExit:
        print(f"基分支 '{base_branch}' 不存在", file=sys.stderr)
        sys.exit(1)

    # 排除 dist 目录
    if is_excluded_file(file_path):
        print(f"文件 '{file_path}' 位于 dist 目录下，已按规则过滤", file=sys.stderr)
        sys.exit(1)

    # 获取 merge-base
    merge_base = run_git("merge-base", base_branch, "HEAD")
    if not merge_base:
        print(f"无法计算 '{base_branch}' 与当前分支的 merge-base", file=sys.stderr)
        sys.exit(1)

    # 检查文件是否有改动
    normalized_path = file_path.replace("\\", "/")
    result = subprocess.run(
        ["git", "diff", "--name-only", merge_base, "HEAD", "--", normalized_path],
        capture_output=True, text=True,
    )
    if result.returncode != 0 or not result.stdout.strip():
        print(f"文件 '{file_path}' 在当前分支中没有改动", file=sys.stderr)
        sys.exit(1)

    # 输出 diff 内容
    result = subprocess.run(
        ["git", "diff", merge_base, "HEAD", "--", normalized_path],
        capture_output=True, text=True,
    )
    if result.returncode != 0:
        print(f"获取文件 '{file_path}' 的 diff 内容失败", file=sys.stderr)
        sys.exit(1)

    print(result.stdout, end="")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
