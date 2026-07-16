# dtool-git reference

## Local actions

Both actions compare the current `HEAD` with the branch explicitly supplied in `compare_branch`.

### `changes`

Return a structured summary and file list.

Required in branch mode:

- `--local-dir <repo>`
- `--compare-branch <branch>`

Options:

- `--compare-strategy merge_base|tree`: default `merge_base`; use `tree` only for a literal snapshot comparison.
- `--scope all|frontend|backend`: `frontend` means `web/**`; `backend` excludes `web/**`.
- `--include-workspace`: include staged and unstaged final state in a branch comparison.
- `--[no-]include-untracked`: include untracked files when workspace content is part of the comparison.
- `--file-path <path>`: limit the result to one repository-relative file.
- `--exclude <path>`: add a repository-relative excluded path; repeat as needed.
- `--max-files <n>`: cap the returned file array; summary totals still cover the full comparison.

### `diff`

Accept all `changes` options and return `diff`, `diff_bytes`, `original_diff_bytes`, and `truncated`.

- `--context-lines <n>`: default 3.
- `--max-diff-bytes <n>`: default 200000.
- `--include-content`: for a single file, add old/new text, binary sizes, or image data. Avoid this for ordinary AI review.

### Comparison semantics

- Branch default: `merge-base(compare_branch, HEAD) -> HEAD`.
- Branch with `--compare-strategy tree`: `compare_branch -> HEAD`.
- Workspace: `HEAD -> working tree`, plus untracked files.
- Branch with `--include-workspace`: `merge-base/tree start -> working tree`, plus untracked files.

Use `--compare-mode workspace` without `--compare-branch` for workspace-only changes.

## Remote actions

Run `python scripts/git_tool.py remote <action> [options]`.

| Action | Required parameters | Purpose |
| --- | --- | --- |
| `config_list` | base URL | List Git groups and repositories |
| `current_branch` | `git_id` | Query local and tracked remote branch |
| `status` | `git_id` | Query repository status |
| `commit_log` | `git_id` | Query recent commits |
| `pull` | `git_id` | Pull the current branch |
| `change_branch` | `git_id`, `branch_name` | Switch a local branch |
| `change_remote_branch` | `git_id`, `branch_name` | Associate/switch a remote branch |
| `remote_branches` | `git_id` | List remote branches |
| `group_branches` | `git_group_id` | Query branches for all repositories in a group |
| `quick_create_branch` | `git_id`, `base_branch`, `branch_type`, `business_en` | Create and push a feature/hotfix branch |
| `set_safe` | `git_id` | Configure `safe.directory` |
| `save_credentials` | `git_id` | Configure the repository credential helper |
| `upload_file` | `git_id`, `local_file_paths` JSON | Upload files to repository-relative destinations |
| `settings_list` | none | List editable Git settings |
| `config_add` / `config_delete` | `payload_json` | Create, update, or delete a Git configuration |
| `group_list` | none | List editable Git groups |
| `group_add` / `group_delete` | `payload_json` | Create, update, or delete a Git group |
| `quick_list` | `payload_json` | Query quick Git directory configuration |

Pass `--discard-local-changes` only after explicit authorization. Without it, mutating branch operations fail when the remote worktree is dirty.

## Errors

The tool exits nonzero and prints:

```json
{
  "ok": false,
  "error": {
    "code": "COMPARE_BRANCH_NOT_FOUND",
    "message": "对比分支不存在: master"
  }
}
```

Common codes include `COMPARE_BRANCH_REQUIRED`, `COMPARE_BRANCH_NOT_FOUND`, `FILE_OUTSIDE_REPOSITORY`, `INVALID_SCOPE`, `GIT_ID_REQUIRED`, and `REMOTE_OPERATION_FAILED`.
