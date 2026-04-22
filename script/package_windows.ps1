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

$RootDir = Split-Path -Parent $PSScriptRoot
$BuildDir = Join-Path $RootDir "build"
$StageDir = Join-Path $BuildDir "release_windows"
$PackageDir = Join-Path $StageDir "package"
$ZipFile = Join-Path $BuildDir "dtool_release_windows.zip"
$FrontendDistDir = Join-Path $RootDir "web/dist"
$WebExe = Join-Path $BuildDir "dtool.exe"

New-Item -ItemType Directory -Force -Path $PackageDir | Out-Null

Write-Step "[1/4] 校验构建产物"
Assert-PathExists -Path $FrontendDistDir -Description "前端构建产物 web/dist"
Assert-PathExists -Path $WebExe -Description "Windows Web 模式后端 build/dtool.exe"

Write-Step "[2/4] 复制运行资源"
Copy-Item $WebExe (Join-Path $PackageDir "dtool.exe") -Force
Copy-Item (Join-Path $RootDir "go.mod") (Join-Path $PackageDir "go.mod") -Force
New-Item -ItemType Directory -Force -Path (Join-Path $PackageDir "config/dtool") | Out-Null
Copy-Item (Join-Path $RootDir "config/dtool/company.ini") (Join-Path $PackageDir "config/dtool/config.ini") -Force
Copy-Item (Join-Path $RootDir "config/dtool/frog.db") (Join-Path $PackageDir "config/dtool/frog.db") -Force
Copy-Item $FrontendDistDir (Join-Path $PackageDir "web/dist") -Recurse -Force
Copy-Item (Join-Path $RootDir "internal/pkg/p_js") (Join-Path $PackageDir "internal/pkg/p_js") -Recurse -Force
Copy-Item (Join-Path $RootDir "internal/app/dtool/database") (Join-Path $PackageDir "internal/app/dtool/database") -Recurse -Force
Copy-Item (Join-Path $RootDir "internal/app/dtool/database_memory") (Join-Path $PackageDir "internal/app/dtool/database_memory") -Recurse -Force

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
Set-Content -Path (Join-Path $PackageDir "网页版.bat") -Value $WebLauncher -Encoding UTF8
Set-Content -Path (Join-Path $PackageDir "README_RELEASE.txt") -Value $ReleaseNote -Encoding UTF8

Write-Step "[4/4] 压缩 zip"
if (Test-Path $ZipFile) {
    Remove-Item $ZipFile -Force
}
Compress-Archive -Path (Join-Path $PackageDir "*") -DestinationPath $ZipFile -Force

Write-Host ""
Write-Host "[OK] Package created: $ZipFile"
