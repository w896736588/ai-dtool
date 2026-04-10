$ErrorActionPreference = 'Stop'

$airCmd = Get-Command air -ErrorAction SilentlyContinue
if ($airCmd) {
    & $airCmd.Source -c .air.toml
    exit $LASTEXITCODE
}

$gopath = (go env GOPATH).Trim()
$airExe = Join-Path $gopath 'bin\air.exe'
if (Test-Path $airExe) {
    & $airExe -c .air.toml
    exit $LASTEXITCODE
}

Write-Error 'air not found. Install it with: go install github.com/air-verse/air@latest'
