@echo off
setlocal EnableExtensions EnableDelayedExpansion
chcp 65001 >nul

REM dtool one-click build package script
REM Build: web/dist + dtool.exe + zip

cd /d "%~dp0.."
set "ROOT_DIR=%cd%"
set "BUILD_DIR=%ROOT_DIR%\build"

set "STAGE_DIR=%BUILD_DIR%\release"
set "PKG_DIR=%STAGE_DIR%\package"
set "ZIP_FILE=%BUILD_DIR%\dtool_release.zip"

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

echo [1/5] Build frontend web/dist
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

echo [2/5] Build web backend exe
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-s -w" -o "%PKG_DIR%\dtool.exe" ./cmd/dtool || goto :error

echo [3/5] Copy runtime resources
copy /Y "%ROOT_DIR%\go.mod" "%PKG_DIR%\go.mod" >nul || goto :error
if not exist "%PKG_DIR%\config\dtool" mkdir "%PKG_DIR%\config\dtool" || goto :error
REM Copy only company.ini and rename it to config.ini
copy /Y "%ROOT_DIR%\config\dtool\company.ini" "%PKG_DIR%\config\dtool\config.ini" >nul || goto :error
copy /Y "%ROOT_DIR%\config\dtool\frog.db" "%PKG_DIR%\config\dtool\frog.db" >nul || goto :error
xcopy "%ROOT_DIR%\web\dist" "%PKG_DIR%\web\dist" /E /I /Y >nul || goto :error
xcopy "%ROOT_DIR%\internal\pkg\p_js" "%PKG_DIR%\internal\pkg\p_js" /E /I /Y >nul || goto :error
xcopy "%ROOT_DIR%\internal\app\dtool\database" "%PKG_DIR%\internal\app\dtool\database" /E /I /Y >nul || goto :error
xcopy "%ROOT_DIR%\internal\app\dtool\database_memory" "%PKG_DIR%\internal\app\dtool\database_memory" /E /I /Y >nul || goto :error

echo [4/5] Generate launch scripts and release note
for /f %%i in ('powershell -NoProfile -Command "[string]([char]32593+[char]39029+[char]29256)+'.bat'"') do set "WEB_BAT=%%i"

(
  echo @echo off
  echo chcp 65001 ^>nul
  echo start "dtool-web" /D "%%~dp0" "%%~dp0dtool.exe" --ConfigFile=config
  echo timeout /t 2 /nobreak ^>nul
  echo start "" "http://localhost:17170/"
) > "%PKG_DIR%\%WEB_BAT%"

(
  echo dtool release package
  echo.
  echo Run web mode:
  echo   Double-click %WEB_BAT%
  echo.
  echo Notes:
  echo 1. ConfigFile matches config\dtool\*.ini filename without extension
  echo 2. Check webPath/dbPath and other ini settings before first run
) > "%PKG_DIR%\README_RELEASE.txt"

echo [5/5] Compress zip
if exist "%ZIP_FILE%" del /f /q "%ZIP_FILE%"
powershell -NoProfile -Command "Compress-Archive -Path '%PKG_DIR%\*' -DestinationPath '%ZIP_FILE%' -Force" || goto :error

echo.
echo [OK] Package created: %ZIP_FILE%
exit /b 0

:error
echo.
echo [ERROR] Build or packaging failed
exit /b 1
