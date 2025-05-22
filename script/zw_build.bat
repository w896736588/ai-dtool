@echo on

set "targetDirectory=D:\go\release\zw\devtool"
if not exist "%targetDirectory%" (
    echo dir：%targetDirectory% not exist,creating...
    mkdir "%targetDirectory%"
)

set "targetDirectory1=D:\go\release\zw\goservice\build"
if not exist "%targetDirectory1%" (
    echo dir：%targetDirectory1% not exist,creating...
    mkdir "%targetDirectory1%"
)

set "targetDirectory2=D:\go\release\zwPub\goservice\build"
if not exist "%targetDirectory2%" (
    echo dir：%targetDirectory2% not exist,creating...
    mkdir "%targetDirectory2%"
)

set "targetDirectory3=D:\go\release"
if not exist "%targetDirectory3%" (
    echo dir：%targetDirectory3% not exist,creating...
    mkdir "%targetDirectory3%"
)

copy D:\job\cache_manager_api\build\zw.exe D:\go\release\zw\goservice\build\zw.exe /Y
xcopy D:\job\cache_manager_api\config\zw D:\go\release\zw\goservice\config\zw /E /Y /I
xcopy D:\job\cache_manager_api\script\zw_start.bat D:\go\release\zw\ /y
xcopy D:\job\cache_manager_api\go.mod D:\go\release\zw\goservice /y
xcopy D:\job\cache_manager_api\internal\pkg\p_js D:\go\release\zw\goservice\internal\pkg\p_js /E /Y /I
xcopy D:\job\cache_manager_api\internal\pkg\p_node D:\go\release\zw\goservice\internal\pkg\p_node /E /Y /I
xcopy D:\job\cache_manager_web\public\favicon.ico D:\go\release\zw\devtool /y
xcopy D:\job\cache_manager_web\dist D:\go\release\zw\devtool\dist /E /Y /I

copy D:\job\cache_manager_api\build\zwPub.exe D:\go\release\zwPub\goservice\build\zw.exe /Y
xcopy D:\job\cache_manager_api\config D:\go\release\zwPub\goservice\config\ /E /Y /I
xcopy D:\job\cache_manager_api\internal\pkg\p_js D:\go\release\zwPub\goservice\internal\pkg\p_js /E /Y /I
xcopy D:\job\cache_manager_api\internal\pkg\p_node D:\go\release\zwPub\goservice\internal\pkg\p_node /E /Y /I
xcopy D:\job\cache_manager_api\script\zw_start.bat D:\go\release\zwPub\ /y
xcopy D:\job\cache_manager_api\go.mod D:\go\release\zwPub\goservice\ /y
xcopy D:\job\cache_manager_web\public\favicon.ico D:\go\release\zwPub\devtool\ /y
xcopy D:\job\cache_manager_web\dist D:\go\release\zwPub\devtool\dist\ /E /Y /I
if exist "D:\go\release\zwPub\goservice\playwright.RunLock" (
    del /f /q "D:\go\release\zwPub\goservice\playwright.RunLock"
)
if exist "D:\go\release\zwPub.zip" (
    del /f /q "D:\go\release\zwPub.zip"
)

"C:\Program Files\WinRAR\winrar.exe" a -afzip -r -ep1 D:\go\release\zwPub.zip D D:\go\release\zwPub