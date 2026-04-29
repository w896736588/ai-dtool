# 显示当前分支改动的文件路径列表（类似 GitLab MR 文件列表）
# 用法: .\show-branch-diff.ps1 [基分支名]

param(
    [string]$BaseBranch = ""
)

# 自动推断默认基分支 / Detect the default base branch automatically
function Get-DefaultBaseBranch {
    $branches = git branch -r | ForEach-Object { $_.Trim() }
    if ($branches -contains "origin/main") { return "origin/main" }
    if ($branches -contains "origin/master") { return "origin/master" }
    if ($branches -contains "main") { return "main" }
    if ($branches -contains "master") { return "master" }
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
    $BaseBranch = Get-DefaultBaseBranch
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
