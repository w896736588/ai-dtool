# 显示指定文件在当前分支中的改动内容（类似 GitLab MR 单文件 diff）
# 用法: .\show-file-diff.ps1 <文件路径> [基分支名]

param(
    [Parameter(Mandatory = $true, Position = 0)]
    [string]$FilePath,

    [Parameter(Position = 1)]
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

# 关键判断：排除 Vue dist 目录 / Critical guard: exclude Vue dist artifacts
function Test-IsExcludedFile {
    param([string]$Path)

    $normalizedPath = $Path.Replace("\", "/")
    return $normalizedPath -match '(^|/)dist/'
}

# 检查是否在 git 仓库中 / Ensure current directory is a git repository
$gitRoot = git rev-parse --show-toplevel 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Error "当前目录不是 git 仓库 / Current directory is not a git repository"
    exit 1
}

if ([string]::IsNullOrWhiteSpace($BaseBranch)) {
    $BaseBranch = Detect-BaseBranch
    if ([string]::IsNullOrWhiteSpace($BaseBranch)) {
        Write-Error "无法自动检测基分支，请手动指定: .\show-file-diff.ps1 <文件路径> <base-branch> / Failed to detect base branch automatically"
        exit 1
    }
}

$baseExists = git rev-parse --verify $BaseBranch 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Error "基分支 '$BaseBranch' 不存在 / Base branch '$BaseBranch' does not exist"
    exit 1
}

if (Test-IsExcludedFile -Path $FilePath) {
    Write-Error "文件 '$FilePath' 位于 dist 目录下，已按规则过滤 / File '$FilePath' is filtered because it is under dist"
    exit 1
}

$mergeBase = Get-MergeBaseCommit -Base $BaseBranch
if ([string]::IsNullOrWhiteSpace($mergeBase)) {
    Write-Error "无法计算 '$BaseBranch' 与当前分支的 merge-base / Failed to resolve merge-base for '$BaseBranch'"
    exit 1
}

$normalizedFilePath = $FilePath.Replace("\", "/")
$nameOnlyArgs = @("diff", "--name-only", $mergeBase, "HEAD", "--", $normalizedFilePath)
$fileChanged = @(git @nameOnlyArgs 2>$null)
if ($LASTEXITCODE -ne 0) {
    Write-Error "无法检查文件 '$FilePath' 的改动状态 / Failed to inspect change state for '$FilePath'"
    exit 1
}

if ($fileChanged.Count -eq 0) {
    Write-Error "文件 '$FilePath' 在当前分支中没有改动 / File '$FilePath' has no changes in current branch"
    exit 1
}

$diffArgs = @("diff", $mergeBase, "HEAD", "--", $normalizedFilePath)
$diffLines = @(git @diffArgs 2>$null)
if ($LASTEXITCODE -ne 0) {
    Write-Error "无法获取文件 '$FilePath' 的 diff 内容 / Failed to load diff content for '$FilePath'"
    exit 1
}

foreach ($line in $diffLines) {
    Write-Output $line
}
