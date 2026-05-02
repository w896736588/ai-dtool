#!/usr/bin/env bash

set -euo pipefail

# 用法: ./show-file-diff.sh <基分支> <文件路径>
# 显示指定文件在当前分支中的改动内容（类似 GitLab MR 单文件 diff）

BASE_BRANCH="${1:-}"
FILE_PATH="${2:-}"

if [[ -z "${BASE_BRANCH}" ]]; then
  echo "用法: ./show-file-diff.sh <基分支> <文件路径> / Usage: ./show-file-diff.sh <base-branch> <file-path>" >&2
  exit 1
fi

if [[ -z "${FILE_PATH}" ]]; then
  echo "用法: ./show-file-diff.sh <基分支> <文件路径> / Usage: ./show-file-diff.sh <base-branch> <file-path>" >&2
  exit 1
fi

# 获取 merge-base，确保比较语义与 GitLab MR 一致 / Resolve merge-base to match GitLab MR style diff semantics
get_merge_base_commit() {
  local base_branch="$1"
  git merge-base "${base_branch}" HEAD 2>/dev/null
}

# 关键判断：排除 Vue dist 目录 / Critical guard: exclude Vue dist artifacts
is_excluded_file() {
  local normalized_path="$1"
  [[ "${normalized_path}" =~ (^|/)dist/ ]]
}

# 检查是否在 git 仓库中 / Ensure current directory is a git repository
if ! git rev-parse --show-toplevel >/dev/null 2>&1; then
  echo "当前目录不是 git 仓库 / Current directory is not a git repository" >&2
  exit 1
fi

if ! git rev-parse --verify "${BASE_BRANCH}" >/dev/null 2>&1; then
  echo "基分支 '${BASE_BRANCH}' 不存在 / Base branch '${BASE_BRANCH}' does not exist" >&2
  exit 1
fi

NORMALIZED_FILE_PATH="${FILE_PATH//\\//}"
if is_excluded_file "${NORMALIZED_FILE_PATH}"; then
  echo "文件 '${FILE_PATH}' 位于 dist 目录下，已按规则过滤 / File '${FILE_PATH}' is filtered because it is under dist" >&2
  exit 1
fi

MERGE_BASE="$(get_merge_base_commit "${BASE_BRANCH}")"
if [[ -z "${MERGE_BASE}" ]]; then
  echo "无法计算 '${BASE_BRANCH}' 与当前分支的 merge-base / Failed to resolve merge-base for '${BASE_BRANCH}'" >&2
  exit 1
fi

if ! git diff --name-only "${MERGE_BASE}" HEAD -- "${NORMALIZED_FILE_PATH}" | grep -q .; then
  echo "文件 '${FILE_PATH}' 在当前分支中没有改动 / File '${FILE_PATH}' has no changes in current branch" >&2
  exit 1
fi

git diff "${MERGE_BASE}" HEAD -- "${NORMALIZED_FILE_PATH}"
