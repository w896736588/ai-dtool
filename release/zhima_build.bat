@echo on
xcopy D:\go\redis_manager\build D:\go\redis_manager\release\zhima\goservice\build /E /Y /I
xcopy D:\go\redis_manager\config D:\go\redis_manager\release\zhima\goservice\config /E /Y /I
xcopy D:\go\redis_manager\start.bat D:\go\redis_manager\release\zhima\ /y
xcopy D:\go\redis_manager\go.mod D:\go\redis_manager\release\zhima\goservice /y
xcopy D:\go\devtool\public\favicon.ico D:\go\redis_manager\release\zhima\devtool /y
xcopy D:\go\devtool\dist D:\go\redis_manager\release\zhima\devtool\dist /E /Y /I


copy D:\go\redis_manager\build\zhimaPub.exe D:\go\redis_manager\release\zhimaPub\goservice\build\zhima.exe /Y
xcopy D:\go\redis_manager\config D:\go\redis_manager\release\zhimaPub\goservice\config\ /E /Y /I
xcopy D:\go\redis_manager\start.bat D:\go\redis_manager\release\zhimaPub\ /y
xcopy D:\go\redis_manager\go.mod D:\go\redis_manager\release\zhimaPub\goservice\ /y
xcopy D:\go\devtool\public\favicon.ico D:\go\redis_manager\release\zhimaPub\devtool\ /y
xcopy D:\go\devtool\dist D:\go\redis_manager\release\zhimaPub\devtool\dist\ /E /Y /I
"C:\Program Files\WinRAR\winrar.exe" a -afzip -r D:\go\redis_manager\release\zhimaPub.zip D D:\go\redis_manager\release\zhimaPub