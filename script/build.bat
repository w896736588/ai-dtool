@echo off
setlocal EnableExtensions EnableDelayedExpansion
chcp 65001 >nul

REM dtool one-click build package script
REM Build: web/dist + dtool.exe + dtool_wails.exe + zip

cd /d "%~dp0.."
set "ROOT_DIR=%cd%"
set "BUILD_DIR=%ROOT_DIR%\build"

for /f %%i in ('powershell -NoProfile -Command "Get-Date -Format yyyyMMdd_HHmmss"') do set "TS=%%i"
set "STAGE_DIR=%BUILD_DIR%\release_%TS%"
set "PKG_DIR=%STAGE_DIR%\package"
set "ZIP_FILE=%BUILD_DIR%\dtool_release_%TS%.zip"

where go >nul 2>nul
if errorlevel 1 (
  echo [ERROR] go not found in PATH
  exit /b 1
)

where npm >nul 2>nul
if errorlevel 1 (
  echo [ERROR] npm not found in PATH
  exit /b 1
)

if exist "%STAGE_DIR%" rmdir /s /q "%STAGE_DIR%"
mkdir "%PKG_DIR%" || goto :error

echo [1/6] Build frontend web/dist
pushd "%ROOT_DIR%\web" || goto :error
if exist node_modules\.cache (
  rmdir /s /q node_modules\.cache
)
if exist package-lock.json (
  call npm ci
  if errorlevel 1 (
    echo [WARN] npm ci failed, clean cache and retry once
    if exist node_modules\.cache (
      rmdir /s /q node_modules\.cache
    )
    call npm cache verify
    call npm ci --no-audit --no-fund || goto :error
  )
) else (
  call npm install --no-audit --no-fund || goto :error
)
call npm run prod || goto :error
popd

echo [2/6] Build web backend exe
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-s -w" -o "%PKG_DIR%\dtool.exe" ./cmd/dtool || goto :error

echo [3/6] Build desktop exe
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64
go build -tags production -ldflags "-s -w" -o "%PKG_DIR%\dtool_wails.exe" ./cmd/dtool_wails || goto :error

echo [4/6] Copy runtime resources
xcopy "%ROOT_DIR%\config\dtool" "%PKG_DIR%\config\dtool" /E /I /Y >nul || goto :error
xcopy "%ROOT_DIR%\web\dist" "%PKG_DIR%\web\dist" /E /I /Y >nul || goto :error
xcopy "%ROOT_DIR%\internal\pkg\p_js" "%PKG_DIR%\internal\pkg\p_js" /E /I /Y >nul || goto :error
xcopy "%ROOT_DIR%\internal\pkg\p_node" "%PKG_DIR%\internal\pkg\p_node" /E /I /Y >nul || goto :error
xcopy "%ROOT_DIR%\internal\app\dtool\database" "%PKG_DIR%\internal\app\dtool\database" /E /I /Y >nul || goto :error

echo [5/6] Generate release note
(
  echo dtool release package
  echo.
  echo Run web mode:
  echo   dtool.exe --ConfigFile=company
  echo.
  echo Run desktop mode:
  echo   dtool_wails.exe --ConfigFile=company
  echo.
  echo Notes:
  echo 1. ConfigFile matches config\dtool\*.ini filename without extension
  echo 2. Check webPath/dbPath and other ini settings before first run
) > "%PKG_DIR%\README_RELEASE.txt"

echo [6/6] Compress zip
if exist "%ZIP_FILE%" del /f /q "%ZIP_FILE%"
powershell -NoProfile -Command "Compress-Archive -Path '%PKG_DIR%\*' -DestinationPath '%ZIP_FILE%' -Force" || goto :error

echo.
echo [OK] Package created: %ZIP_FILE%
exit /b 0

:error
echo.
echo [ERROR] Build or packaging failed
exit /b 1
