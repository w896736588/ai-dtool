#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
显示当前分支相对指定基分支的改动文件路径列表（类似 GitLab MR 文件列表）

用法: python show_branch_diff.py <基分支>
"""

import subprocess
import sys


def run_git(*args: str) -> str:
    result = subprocess.run(["git"] + list(args), capture_output=True, text=True)
    if result.returncode != 0:
        msg = result.stderr.strip()
        print(f"ERROR: {msg}", file=sys.stderr)
        sys.exit(1)
    return result.stdout.strip()


def main() -> int:
    if len(sys.argv) < 2:
        print("用法: python show_branch_diff.py <基分支>", file=sys.stderr)
        sys.exit(1)

    base_branch = sys.argv[1].strip()
    if not base_branch:
        print("基分支不能为空", file=sys.stderr)
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

    # 获取 merge-base
    merge_base = run_git("merge-base", base_branch, "HEAD")
    if not merge_base:
        print(f"无法计算 '{base_branch}' 与当前分支的 merge-base", file=sys.stderr)
        sys.exit(1)

    # 获取改动文件列表，排除 dist 目录
    result = subprocess.run(
        ["git", "diff", "--name-only", merge_base, "HEAD", "--", ".", ":(exclude)**/dist/**"],
        capture_output=True, text=True,
    )
    if result.returncode != 0:
        print(f"获取改动文件列表失败: {result.stderr.strip()}", file=sys.stderr)
        sys.exit(1)

    files = [f for f in result.stdout.strip().splitlines() if f.strip()]
    for f in files:
        print(f)

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
