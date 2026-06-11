#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
显示当前分支相对指定基分支的改动文件路径列表（类似 GitLab MR 文件列表）

用法: python show_branch_diff.py <基分支>
"""

import subprocess
import sys

# Windows 中文环境下 stdout 默认编码为 GBK，导致 UTF-8 输出乱码。
sys.stdout.reconfigure(encoding='utf-8', errors='replace')
sys.stderr.reconfigure(encoding='utf-8', errors='replace')


def run_git(*args: str) -> str:
    result = subprocess.run(
        ["git"] + list(args), capture_output=True, text=True,
        encoding="utf-8", errors="replace",
    )
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

    exclude = ["--", ".", ":(exclude)**/dist/**"]

    # 分别收集各状态文件
    committed = set()
    staged = set()
    modified = set()
    untracked = set()

    # 已提交：merge_base vs HEAD
    result = subprocess.run(
        ["git", "diff", "--name-only", merge_base, "HEAD"] + exclude,
        capture_output=True, text=True, encoding="utf-8", errors="replace",
    )
    if result.returncode != 0:
        print(f"获取改动文件列表失败: {result.stderr.strip()}", file=sys.stderr)
        sys.exit(1)
    committed = set(f for f in result.stdout.strip().splitlines() if f.strip())

    # 暂存区：已 git add 未 commit
    result_cached = subprocess.run(
        ["git", "diff", "--name-only", "--cached"] + exclude,
        capture_output=True, text=True, encoding="utf-8", errors="replace",
    )
    if result_cached.returncode == 0:
        staged = set(f for f in result_cached.stdout.strip().splitlines() if f.strip())

    # 工作区：未 git add 的改动
    result_wt = subprocess.run(
        ["git", "diff", "--name-only"] + exclude,
        capture_output=True, text=True, encoding="utf-8", errors="replace",
    )
    if result_wt.returncode == 0:
        modified = set(f for f in result_wt.stdout.strip().splitlines() if f.strip())

    # 未跟踪文件
    result_untracked = subprocess.run(
        ["git", "ls-files", "--others", "--exclude-standard", "."],
        capture_output=True, text=True, encoding="utf-8", errors="replace",
    )
    if result_untracked.returncode == 0:
        untracked = set(f for f in result_untracked.stdout.strip().splitlines() if f.strip())

    # 合并所有文件，标注状态
    all_files = committed | staged | modified | untracked
    for f in sorted(all_files):
        statuses = []
        if f in committed:
            statuses.append("Committed")
        if f in staged:
            statuses.append("Staged")
        if f in modified:
            statuses.append("Modified")
        if f in untracked:
            statuses.append("Untracked")
        print(f"{f}\t[{','.join(statuses)}]")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
