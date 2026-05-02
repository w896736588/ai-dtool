#!/usr/bin/env bash

set -euo pipefail

# 用法: ./show-branch-diff.sh <基分支>
# 显示当前分支相对指定基分支的改动文件路径列表（类似 GitLab MR 文件列表）

BASE_BRANCH="${1:-}"

if [[ -z "${BASE_BRANCH}" ]]; then
  echo "用法: ./show-branch-diff.sh <基分支> / Usage: ./show-branch-diff.sh <base-branch>" >&2
  exit 1
fi

# Git 路径过滤规则，排除 Vue dist 构建产物 / Git pathspec filters to exclude Vue dist build artifacts
DIFF_PATHS=(-- . ":(exclude)**/dist/**")

# 获取 merge-base，确保比较语义与 GitLab MR 一致 / Resolve merge-base to match GitLab MR style diff semantics
get_merge_base_commit() {
  local base_branch="$1"
  git merge-base "${base_branch}" HEAD 2>/dev/null
}

# 获取改动文件列表 / List changed files from merge-base to HEAD
get_changed_files() {
  local merge_base="$1"
  git diff --name-only "${merge_base}" HEAD "${DIFF_PATHS[@]}" 2>/dev/null
}

# 关键校验：必须在 git 仓库中运行 / Critical guard: script must run inside a git repository
if ! git rev-parse --show-toplevel >/dev/null 2>&1; then
  echo "当前目录不是 git 仓库 / Current directory is not a git repository" >&2
  exit 1
fi

if ! git rev-parse --verify "${BASE_BRANCH}" >/dev/null 2>&1; then
  echo "基分支 '${BASE_BRANCH}' 不存在 / Base branch '${BASE_BRANCH}' does not exist" >&2
  exit 1
fi

MERGE_BASE="$(get_merge_base_commit "${BASE_BRANCH}")"
if [[ -z "${MERGE_BASE}" ]]; then
  echo "无法计算 '${BASE_BRANCH}' 与当前分支的 merge-base / Failed to resolve merge-base for '${BASE_BRANCH}'" >&2
  exit 1
fi

get_changed_files "${MERGE_BASE}"
