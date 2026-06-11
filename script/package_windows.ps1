[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

$ErrorActionPreference = "Stop"

function Write-Step {
    param([string]$Message)
    Write-Host $Message
}

function Assert-PathExists {
    param(
        [string]$Path,
        [string]$Description
    )

    # 打包前先校验构建产物是否存在，避免复制阶段才出现不明确错误。
    if (-not (Test-Path $Path)) {
        throw "缺少$Description：$Path"
    }
}

function Remove-IfExists {
    param([string]$Path)

    if (Test-Path $Path) {
        Remove-Item $Path -Recurse -Force
    }
}

$PackageName = if ($args.Count -gt 0) { $args[0] } else { "" }
if ($PackageName -match '[\\/:*?"<>|]') {
    throw "包名包含非法字符：$PackageName"
}

$RootDir = Split-Path -Parent $PSScriptRoot
$BuildDir = Join-Path $RootDir "build"
$StageDir = Join-Path $BuildDir "release_windows"
$PackageDir = if ($PackageName -ne "") {
    Join-Path $StageDir "package_$PackageName"
} else {
    Join-Path $StageDir "package"
}
$TempPackageDir = if ($PackageName -ne "") {
    Join-Path $StageDir "__package_build_$PackageName"
} else {
    Join-Path $StageDir "__package_build"
}
$ZipFile = Join-Path $BuildDir "dtool_release_windows.zip"
$FrontendDistDir = Join-Path $RootDir "web/dist"
$WebExe = Join-Path $BuildDir "dtool.exe"

Remove-IfExists $TempPackageDir
New-Item -ItemType Directory -Force -Path $TempPackageDir | Out-Null

Write-Step "[1/4] 校验构建产物"
Assert-PathExists -Path $FrontendDistDir -Description "前端构建产物 web/dist"
Assert-PathExists -Path $WebExe -Description "Windows Web 模式后端 build/dtool.exe"

Write-Step "[2/4] 复制运行资源"
Copy-Item $WebExe (Join-Path $TempPackageDir "dtool.exe") -Force
Copy-Item (Join-Path $RootDir "go.mod") (Join-Path $TempPackageDir "go.mod") -Force
New-Item -ItemType Directory -Force -Path (Join-Path $TempPackageDir "config/dtool") | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $TempPackageDir "web") | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $TempPackageDir "web/dist") | Out-Null
Copy-Item (Join-Path $FrontendDistDir "*") (Join-Path $TempPackageDir "web/dist") -Recurse -Force
Copy-Item (Join-Path $RootDir "internal/pkg/p_js") (Join-Path $TempPackageDir "internal/pkg/p_js") -Recurse -Force
Copy-Item (Join-Path $RootDir "internal/app/dtool/database") (Join-Path $TempPackageDir "internal/app/dtool/database") -Recurse -Force
Copy-Item (Join-Path $RootDir "internal/app/dtool/database_log") (Join-Path $TempPackageDir "internal/app/dtool/database_log") -Recurse -Force
Copy-Item (Join-Path $RootDir "skills") (Join-Path $TempPackageDir "skills") -Recurse -Force

Write-Step "[3/4] 生成启动脚本和说明文件"
# 显式拼接多行文本，避免 here-string 在部分 PowerShell 环境下解析异常。
$WebLauncher = @(
    '@echo off'
    'chcp 65001 >nul'
    'start "dtool-web" /D "%~dp0" "%~dp0dtool.exe" --ConfigFile=config'
    'timeout /t 2 /nobreak >nul'
    'start "" "http://localhost:17170/"'
) -join "`r`n"

$ReleaseNote = @(
    'dtool release package (windows)'
    ''
    'Run web mode:'
    '  Double-click 网页版.bat'
    ''
    'Notes:'
    '1. ConfigFile matches config\dtool\*.ini filename without extension'
    '2. Check webPath/dbPath and other ini settings before first run'
) -join "`r`n"
Set-Content -Path (Join-Path $TempPackageDir "网页版.bat") -Value $WebLauncher -Encoding UTF8
Set-Content -Path (Join-Path $TempPackageDir "README_RELEASE.txt") -Value $ReleaseNote -Encoding UTF8

Write-Step "[4/4] 压缩 zip"
if (Test-Path $ZipFile) {
    Remove-Item $ZipFile -Force
}
$ArchiveEntries = @(
    (Join-Path $TempPackageDir "config"),
    (Join-Path $TempPackageDir "internal"),
    (Join-Path $TempPackageDir "skills"),
    (Join-Path $TempPackageDir "web"),
    (Join-Path $TempPackageDir "dtool.exe"),
    (Join-Path $TempPackageDir "go.mod"),
    (Join-Path $TempPackageDir "README_RELEASE.txt"),
    (Join-Path $TempPackageDir "网页版.bat")
)
Compress-Archive -Path $ArchiveEntries -DestinationPath $ZipFile -Force

try {
    New-Item -ItemType Directory -Force -Path $PackageDir | Out-Null

    $PackageEntries = @(
        "config",
        "internal",
        "skills",
        "web",
        "dtool.exe",
        "go.mod",
        "README_RELEASE.txt",
        "网页版.bat"
    )

    foreach ($Entry in $PackageEntries) {
        $SourcePath = Join-Path $TempPackageDir $Entry
        if (Test-Path $SourcePath) {
            $DestPath = Join-Path $PackageDir $Entry
            if ((Test-Path $SourcePath -PathType Container) -and (Test-Path $DestPath)) {
                # 目录已存在：复制内容而非目录本身，避免嵌套
                Copy-Item (Join-Path $SourcePath "*") $DestPath -Recurse -Force
            } else {
                Copy-Item $SourcePath $DestPath -Recurse -Force
            }
        }
    }

    Remove-IfExists $TempPackageDir
} catch {
    Write-Warning "无法更新输出目录 $PackageDir，已保留本次打包目录：$TempPackageDir。原因：$($_.Exception.Message)"
}

Write-Host ""
Write-Host "[OK] Package created: $ZipFile"
