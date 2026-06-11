#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
显示指定文件在当前分支中的改动内容

用法: python show_file_diff.py <基分支> <文件路径>

输出 JSON: {"diff": "...", "old_content": "...", "new_content": "..."}
"""

import json
import os
import re
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


def run_git_safe(*args: str) -> str:
    """运行 git 命令，出错时返回空字符串而不退出。"""
    result = subprocess.run(
        ["git"] + list(args), capture_output=True, text=True,
        encoding="utf-8", errors="replace",
    )
    if result.returncode != 0:
        return ""
    return result.stdout


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

    normalized_path = file_path.replace("\\", "/")

    # 依次尝试获取 diff 文本：已提交的改动 -> 暂存区改动 -> 工作区改动
    diff_content = ""

    # 1) 已提交的改动（merge_base vs HEAD）
    result = subprocess.run(
        ["git", "diff", merge_base, "HEAD", "--", normalized_path],
        capture_output=True, text=True, encoding="utf-8", errors="replace",
    )
    if result.returncode == 0 and result.stdout and result.stdout.strip():
        diff_content = result.stdout

    # 2) 暂存区改动（已 git add 未 commit）
    if not diff_content:
        result = subprocess.run(
            ["git", "diff", "--cached", "--", normalized_path],
            capture_output=True, text=True, encoding="utf-8", errors="replace",
        )
        if result.returncode == 0 and result.stdout and result.stdout.strip():
            diff_content = result.stdout

    # 3) 工作区改动（未 git add）
    if not diff_content:
        result = subprocess.run(
            ["git", "diff", "--", normalized_path],
            capture_output=True, text=True, encoding="utf-8", errors="replace",
        )
        if result.returncode == 0 and result.stdout and result.stdout.strip():
            diff_content = result.stdout

    # 获取基分支文件内容（old_content）
    old_content = run_git_safe("show", f"{merge_base}:{normalized_path}")

    # 获取当前工作区文件内容（new_content）
    new_content = ""
    abs_path = os.path.abspath(file_path)
    if os.path.isfile(abs_path):
        try:
            with open(abs_path, 'r', encoding='utf-8', errors='replace') as f:
                new_content = f.read()
        except Exception:
            new_content = ""

    output = {
        "diff": diff_content,
        "old_content": old_content,
        "new_content": new_content,
    }

    print(json.dumps(output, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
