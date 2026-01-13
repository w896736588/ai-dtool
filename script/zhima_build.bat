@echo on

set "targetDirectory=D:\go\release\dtool\devtool"
if not exist "%targetDirectory%" (
    echo dir：%targetDirectory% not exist,creating...
    mkdir "%targetDirectory%"
)

set "targetDirectory1=D:\go\release\dtool\goservice\build"
if not exist "%targetDirectory1%" (
    echo dir：%targetDirectory1% not exist,creating...
    mkdir "%targetDirectory1%"
)

set "targetDirectory2=D:\go\release\dtoolPub\goservice\build"
if not exist "%targetDirectory2%" (
    echo dir：%targetDirectory2% not exist,creating...
    mkdir "%targetDirectory2%"
)

set "targetDirectory3=D:\go\release"
if not exist "%targetDirectory3%" (
    echo dir：%targetDirectory3% not exist,creating...
    mkdir "%targetDirectory3%"
)

copy D:\go\cache_manager_api\build\dtool.exe D:\go\release\dtool\goservice\build\dtool.exe /Y
xcopy D:\go\cache_manager_api\script\start.bat D:\go\release\dtool\goservice\ /y
xcopy D:\go\cache_manager_api\go.mod D:\go\release\dtool\goservice /y
xcopy D:\go\cache_manager_api\internal\pkg\p_js D:\go\release\dtool\goservice\internal\pkg\p_js /E /Y /I
xcopy D:\go\cache_manager_api\internal\pkg\p_node D:\go\release\dtool\goservice\internal\pkg\p_node /E /Y /I
xcopy D:\go\cache_manager_api\internal\app\default\database D:\go\release\dtool\goservice\app\default\database /E /Y /I
xcopy D:\go\devtool\public\favicon.ico D:\go\release\dtool\devtool /y
xcopy D:\go\devtool\dist D:\go\release\dtool\devtool\dist /E /Y /I

copy D:\go\cache_manager_api\build\dtool.exe D:\go\release\dtoolPub\goservice\build\dtool.exe /Y
xcopy D:\go\cache_manager_api\internal\pkg\p_js D:\go\release\dtoolPub\goservice\internal\pkg\p_js /E /Y /I
xcopy D:\go\cache_manager_api\internal\pkg\p_node D:\go\release\dtoolPub\goservice\internal\pkg\p_node /E /Y /I
xcopy D:\go\cache_manager_api\internal\app\default\database D:\go\release\dtoolPub\goservice\app\default\database /E /Y /I
xcopy D:\go\cache_manager_api\script\start.bat D:\go\release\dtoolPub\goservice\ /y
xcopy D:\go\cache_manager_api\go.mod D:\go\release\dtoolPub\goservice\ /y
xcopy D:\go\devtool\public\favicon.ico D:\go\release\dtoolPub\devtool\ /y
xcopy D:\go\devtool\dist D:\go\release\dtoolPub\devtool\dist\ /E /Y /I
if exist "D:\go\release\dtoolPub\goservice\playwright.RunLock" (
    del /f /q "D:\go\release\dtoolPub\goservice\playwright.RunLock"
)
if exist "D:\go\release\dtoolPub.zip" (
    del /f /q "D:\go\release\dtoolPub.zip"
)

"C:\Program Files\WinRAR\winrar.exe" a -afzip -r -ep1 D:\go\release\dtoolPub.zip D D:\go\release\dtoolPub