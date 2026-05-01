#!/usr/bin/env bash

set -euo pipefail

FILE_PATH="${1:-}"
BASE_BRANCH="${2:-}"

# 关键判断：文件路径不能为空 / Critical guard: file path is required
if [[ -z "${FILE_PATH}" ]]; then
  echo "用法: ./show-file-diff.sh <文件路径> [基分支名] / Usage: ./show-file-diff.sh <file-path> [base-branch]" >&2
  exit 1
fi

# 自动检测当前分支的真实基分支（merge-base 最近原则）
# 遍历所有本地和远程分支，计算每个分支与 HEAD 的 merge-base，
# 选出 merge-base 距离 HEAD 最近（独占提交数最少）的分支作为基分支
detect_base_branch() {
  local current_branch
  current_branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null)
  if [[ -z "${current_branch}" || "${current_branch}" == "HEAD" ]]; then
    return 1
  fi

  local best_branch=""
  local best_commits=-1

  while IFS= read -r ref; do
    [[ -z "${ref}" ]] && continue
    local branch="${ref#refs/heads/}"
    branch="${branch#refs/remotes/}"
    [[ "${branch}" == "${current_branch}" ]] && continue
    [[ "${branch}" == "HEAD" ]] && continue

    local mb
    mb=$(git merge-base "${branch}" HEAD 2>/dev/null) || continue
    [[ -z "${mb}" ]] && continue

    local commits
    commits=$(git rev-list --count "${mb}..HEAD" 2>/dev/null) || continue
    [[ "${commits}" -eq 0 ]] && continue

    if [[ ${best_commits} -eq -1 || ${commits} -lt ${best_commits} ]]; then
      best_commits="${commits}"
      best_branch="${branch}"
    fi
  done < <(git for-each-ref --format='%(refname)' refs/heads/ refs/remotes/ 2>/dev/null)

  if [[ -n "${best_branch}" ]]; then
    echo "自动检测到基分支: ${best_branch}" >&2
    echo "${best_branch}"
    return 0
  fi
  return 1
}

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

if [[ -z "${BASE_BRANCH}" ]]; then
  if ! BASE_BRANCH="$(detect_base_branch)"; then
    echo "无法自动检测基分支，请手动指定: ./show-file-diff.sh <文件路径> <base-branch> / Failed to detect base branch automatically" >&2
    exit 1
  fi
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
