# 显示指定文件在当前分支中的改动内容（类似 GitLab MR 单文件 diff）
# 用法: .\show-file-diff.ps1 <文件路径> [基分支名]

param(
    [Parameter(Mandatory = $true, Position = 0)]
    [string]$FilePath,

    [Parameter(Position = 1)]
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
    $BaseBranch = Get-DefaultBaseBranch
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
