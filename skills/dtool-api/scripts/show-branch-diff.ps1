# 显示当前分支改动的文件路径列表（类似 GitLab MR 文件列表）
# 用法: .\show-branch-diff.ps1 [基分支名]

param(
    [string]$BaseBranch = ""
)

# 自动检测当前分支的真实基分支（merge-base 最近原则）
# 遍历所有本地和远程分支，计算每个分支与 HEAD 的 merge-base，
# 选出 merge-base 距离 HEAD 最近（独占提交数最少）的分支作为基分支
function Detect-BaseBranch {
    $currentBranch = git rev-parse --abbrev-ref HEAD 2>$null
    if ([string]::IsNullOrWhiteSpace($currentBranch) -or $currentBranch -eq "HEAD") {
        return $null
    }

    $refs = git for-each-ref --format="%(refname)" refs/heads/ refs/remotes/ 2>$null
    if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($refs)) {
        return $null
    }

    $bestBranch = $null
    $bestCommits = -1

    foreach ($ref in ($refs -split "`n" | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })) {
        $ref = $ref.Trim()
        $branch = $ref -replace '^refs/heads/', '' -replace '^refs/remotes/', ''
        if ($branch -eq $currentBranch -or $branch -eq "HEAD") {
            continue
        }

        $mb = git merge-base $branch HEAD 2>$null
        if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($mb)) {
            continue
        }
        $mb = $mb.Trim()

        $commits = git rev-list --count "$mb..HEAD" 2>$null
        if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($commits)) {
            continue
        }
        $commits = [int]$commits.Trim()

        if ($commits -eq 0) {
            continue
        }

        if ($bestCommits -eq -1 -or $commits -lt $bestCommits) {
            $bestCommits = $commits
            $bestBranch = $branch
        }
    }

    if (-not [string]::IsNullOrWhiteSpace($bestBranch)) {
        Write-Host "自动检测到基分支: $bestBranch" -ForegroundColor Cyan
        return $bestBranch
    }

    return $null
}

# 获取 merge-base，确保比较语义与 GitLab MR 一致 / Resolve merge-base to match GitLab MR style diff semantics
function Get-MergeBaseCommit {
    param([string]$Base)

    $mergeBase = git merge-base $Base HEAD 2>$null
    if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($mergeBase)) {
        return $null
    }

    return $mergeBase.Trim()
}

# Git 路径过滤规则，排除 Vue dist 构建产物 / Git pathspec filters to exclude Vue dist build artifacts
function Get-DiffPathspecArgs {
    return @("--", ".", ":(exclude)**/dist/**")
}

# 获取改动文件列表 / List changed files from merge-base to HEAD
function Get-ChangedFiles {
    param([string]$MergeBase)

    $diffArgs = @("diff", "--name-only", $MergeBase, "HEAD") + (Get-DiffPathspecArgs)
    $changedFiles = git @diffArgs 2>$null
    if ($LASTEXITCODE -ne 0) {
        return $null
    }

    return @($changedFiles | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
}

# 检查是否在 git 仓库中 / Ensure current directory is a git repository
$gitRoot = git rev-parse --show-toplevel 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Error "当前目录不是 git 仓库 / Current directory is not a git repository"
    exit 1
}

# 确定基分支 / Resolve base branch
if ([string]::IsNullOrWhiteSpace($BaseBranch)) {
    $BaseBranch = Detect-BaseBranch
    if ([string]::IsNullOrWhiteSpace($BaseBranch)) {
        Write-Error "无法自动检测基分支，请手动指定: .\show-branch-diff.ps1 <base-branch> / Failed to detect base branch automatically"
        exit 1
    }
}

# 验证基分支存在 / Verify base branch exists
$baseExists = git rev-parse --verify $BaseBranch 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Error "基分支 '$BaseBranch' 不存在 / Base branch '$BaseBranch' does not exist"
    exit 1
}

# 通过 merge-base 计算 MR 语义的比较起点 / Use merge-base so repeated merges from master stay MR-like
$mergeBase = Get-MergeBaseCommit -Base $BaseBranch
if ([string]::IsNullOrWhiteSpace($mergeBase)) {
    Write-Error "无法计算 '$BaseBranch' 与当前分支的 merge-base / Failed to resolve merge-base for '$BaseBranch'"
    exit 1
}

$changedFiles = Get-ChangedFiles -MergeBase $mergeBase
if ($null -eq $changedFiles) {
    Write-Error "获取改动文件列表失败 / Failed to load changed file list"
    exit 1
}

if ($changedFiles.Count -eq 0) {
    exit 0
}

foreach ($file in $changedFiles) {
    Write-Output $file
}
