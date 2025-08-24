@echo on

set "targetDirectory=C:\work\frog\tool_zw\devtool"
if not exist "%targetDirectory%" (
    echo dir：%targetDirectory% not exist,creating...
    mkdir "%targetDirectory%"
)

set "targetDirectory1=C:\work\frog\tool_zw\goservice\build"
if not exist "%targetDirectory1%" (
    echo dir：%targetDirectory1% not exist,creating...
    mkdir "%targetDirectory1%"
)

set "targetDirectory2=C:\work\frog\tool_zwPub\goservice\build"
if not exist "%targetDirectory2%" (
    echo dir：%targetDirectory2% not exist,creating...
    mkdir "%targetDirectory2%"
)

set "targetDirectory3=D:\go\release"
if not exist "%targetDirectory3%" (
    echo dir：%targetDirectory3% not exist,creating...
    mkdir "%targetDirectory3%"
)

copy C:\work\frog\cache_manager_api\build\zw.exe C:\work\frog\tool_zw\goservice\build\zw.exe /Y
xcopy C:\work\frog\cache_manager_api\config\zw C:\work\frog\tool_zw\goservice\config\zw /E /Y /I
xcopy C:\work\frog\cache_manager_api\script\zw_start.bat C:\work\frog\tool_zw\goservice\ /y
xcopy C:\work\frog\cache_manager_api\go.mod C:\work\frog\tool_zw\goservice /y
xcopy C:\work\frog\cache_manager_api\internal\pkg\p_js C:\work\frog\tool_zw\goservice\internal\pkg\p_js /E /Y /I
xcopy C:\work\frog\cache_manager_api\internal\pkg\p_node C:\work\frog\tool_zw\goservice\internal\pkg\p_node /E /Y /I
xcopy C:\work\frog\cache_manager_web\public\favicon.ico C:\work\frog\tool_zw\devtool /y
xcopy C:\work\frog\cache_manager_web\dist C:\work\frog\tool_zw\devtool\dist /E /Y /I

copy C:\work\frog\cache_manager_api\build\zwPub.exe C:\work\frog\tool_zwPub\goservice\build\zw.exe /Y
xcopy C:\work\frog\cache_manager_api\config C:\work\frog\tool_zwPub\goservice\config\ /E /Y /I
xcopy C:\work\frog\cache_manager_api\internal\pkg\p_js C:\work\frog\tool_zwPub\goservice\internal\pkg\p_js /E /Y /I
xcopy C:\work\frog\cache_manager_api\internal\pkg\p_node C:\work\frog\tool_zwPub\goservice\internal\pkg\p_node /E /Y /I
xcopy C:\work\frog\cache_manager_api\script\zw_start.bat C:\work\frog\tool_zwPub\goservice\ /y
xcopy C:\work\frog\cache_manager_api\go.mod C:\work\frog\tool_zwPub\goservice\ /y
xcopy C:\work\frog\cache_manager_web\public\favicon.ico C:\work\frog\tool_zwPub\devtool\ /y
xcopy C:\work\frog\cache_manager_web\dist C:\work\frog\tool_zwPub\devtool\dist\ /E /Y /I
if exist "C:\work\frog\tool_zwPub\goservice\playwright.RunLock" (
    del /f /q "C:\work\frog\tool_zwPub\goservice\playwright.RunLock"
)
if exist "C:\work\frog\tool_zwPub.zip" (
    del /f /q "C:\work\frog\tool_zwPub.zip"
)

"C:\Program Files\WinRAR\winrar.exe" a -afzip -r -ep1 C:\work\frog\tool_zwPub.zip D C:\work\frog\tool_zwPub