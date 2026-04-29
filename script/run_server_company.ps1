$ErrorActionPreference = 'Continue'

function Test-IsWindowsHost {
    return $env:OS -eq 'Windows_NT'
}

function Reset-GoCrossCompileEnv {
    # Avoid inheriting stale cross-compile settings that make air build a Linux binary on Windows.
    foreach ($envName in @('GOOS', 'GOARCH', 'CGO_ENABLED')) {
        if (Test-Path "Env:$envName") {
            Remove-Item "Env:$envName"
        }
    }
}

function Get-AirConfigFile {
    if (Test-IsWindowsHost) {
        return '.air.windows.toml'
    }

    return '.air.toml'
}

if (Test-IsWindowsHost) {
    Reset-GoCrossCompileEnv
}

$airConfigFile = Get-AirConfigFile
$airCmd = Get-Command air -ErrorAction SilentlyContinue
if ($airCmd) {
    & $airCmd.Source -c $airConfigFile
    exit $LASTEXITCODE
}

$gopath = (go env GOPATH).Trim()
$airExe = Join-Path $gopath 'bin\air.exe'
if (Test-Path $airExe) {
    & $airExe -c $airConfigFile
    exit $LASTEXITCODE
}

Write-Error 'air not found. Install it with: go install github.com/air-verse/air@latest'
