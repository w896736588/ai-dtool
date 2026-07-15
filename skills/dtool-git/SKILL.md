---
name: dtool-git
description: "Use for dtool Git work: compare the current branch with a user-selected branch, inspect structured changes or full/single-file diffs, inspect workspace changes, upload files, query status/logs/branches, pull or switch branches, create branches, and manage Git page operations through the dtool API."
---

# dtool-git

Use `scripts/git_tool.py` as the only executable entry point. Read [references/detail.md](references/detail.md) for the complete action and parameter reference.

## Local workflow

1. Ask the user to explicitly select `compare_branch` for branch code changes. Never infer it from branch ancestry or assume a default branch.
2. Run `changes` first to obtain the structured file list.
3. Run `diff` only for the full requested scope or selected files. Prefer single-file calls when the complete patch may be large.
4. Keep committed branch changes separate from workspace changes unless the user explicitly requests both.

```powershell
python scripts/git_tool.py changes --local-dir <repo> --compare-branch <branch> --scope all
python scripts/git_tool.py diff --local-dir <repo> --compare-branch <branch> --scope frontend
python scripts/git_tool.py diff --local-dir <repo> --compare-mode workspace --file-path <path>
```

Use `scope=frontend` for `web/**`, `scope=backend` for everything outside `web/**`, and `scope=all` for the repository.

## Remote workflow

Confirm `base_url`, Token, and the required repository/group identifiers before calling remote actions. Pass credentials through `DTOOL_TOKEN` or `--token`; never write a Token into a source file.

```powershell
$env:DTOOL_TOKEN='<token>'
python scripts/git_tool.py remote config_list --base-url <url>
python scripts/git_tool.py remote status --base-url <url> --git-id <id>
```

Treat `pull`, `change_branch`, `change_remote_branch`, and `quick_create_branch` as mutating operations. They refuse a dirty remote working tree by default. Set `--discard-local-changes` only after the user explicitly authorizes deleting all uncommitted and untracked files.

## Output contract

Expect UTF-8 JSON on stdout. Check `ok` before using any other field. On failure, report `error.code` and `error.message`; do not silently retry with destructive options.
