@echo on

set "targetDirectory=D:\go\release\zhima\devtool"
if not exist "%targetDirectory%" (
    echo dir：%targetDirectory% not exist,creating...
    mkdir "%targetDirectory%"
)

set "targetDirectory1=D:\go\release\zhima\goservice\build"
if not exist "%targetDirectory1%" (
    echo dir：%targetDirectory1% not exist,creating...
    mkdir "%targetDirectory1%"
)

set "targetDirectory2=D:\go\release\zhimaPub\goservice\build"
if not exist "%targetDirectory2%" (
    echo dir：%targetDirectory2% not exist,creating...
    mkdir "%targetDirectory2%"
)

set "targetDirectory3=D:\go\release"
if not exist "%targetDirectory3%" (
    echo dir：%targetDirectory3% not exist,creating...
    mkdir "%targetDirectory3%"
)

copy D:\go\cache_manager_api\build\zhima.exe D:\go\release\zhima\goservice\build\zhima.exe /Y
xcopy D:\go\cache_manager_api\config D:\go\release\zhima\goservice\config /E /Y /I
xcopy D:\go\cache_manager_api\script\start.bat D:\go\release\zhima\goservice\ /y
xcopy D:\go\cache_manager_api\go.mod D:\go\release\zhima\goservice /y
xcopy D:\go\cache_manager_api\internal\pkg\p_js D:\go\release\zhima\goservice\internal\pkg\p_js /E /Y /I
xcopy D:\go\cache_manager_api\internal\pkg\p_node D:\go\release\zhima\goservice\internal\pkg\p_node /E /Y /I
xcopy D:\go\devtool\public\favicon.ico D:\go\release\zhima\devtool /y
xcopy D:\go\devtool\dist D:\go\release\zhima\devtool\dist /E /Y /I

copy D:\go\cache_manager_api\build\zhimaPub.exe D:\go\release\zhimaPub\goservice\build\zhima.exe /Y
xcopy D:\go\cache_manager_api\config D:\go\release\zhimaPub\goservice\config\ /E /Y /I
xcopy D:\go\cache_manager_api\internal\pkg\p_js D:\go\release\zhimaPub\goservice\internal\pkg\p_js /E /Y /I
xcopy D:\go\cache_manager_api\internal\pkg\p_node D:\go\release\zhimaPub\goservice\internal\pkg\p_node /E /Y /I
xcopy D:\go\cache_manager_api\script\start.bat D:\go\release\zhimaPub\goservice\ /y
xcopy D:\go\cache_manager_api\go.mod D:\go\release\zhimaPub\goservice\ /y
xcopy D:\go\devtool\public\favicon.ico D:\go\release\zhimaPub\devtool\ /y
xcopy D:\go\devtool\dist D:\go\release\zhimaPub\devtool\dist\ /E /Y /I
if exist "D:\go\release\zhimaPub\goservice\playwright.RunLock" (
    del /f /q "D:\go\release\zhimaPub\goservice\playwright.RunLock"
)
if exist "D:\go\release\zhimaPub.zip" (
    del /f /q "D:\go\release\zhimaPub.zip"
)

"C:\Program Files\WinRAR\winrar.exe" a -afzip -r -ep1 D:\go\release\zhimaPub.zip D D:\go\release\zhimaPub