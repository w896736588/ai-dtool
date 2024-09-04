@echo on
xcopy D:\go\redis_manager\build D:\go\redis_manager\tool\goservice\build /E /Y /I
xcopy D:\go\redis_manager\config D:\go\redis_manager\tool\goservice\config /E /Y /I
xcopy D:\go\redis_manager\start.bat D:\go\redis_manager\tool\ /y
xcopy D:\go\redis_manager\go.mod D:\go\redis_manager\tool\goservice /y
xcopy D:\go\devtool\public\favicon.ico D:\go\redis_manager\tool\devtool /y
xcopy D:\go\devtool\dist D:\go\redis_manager\tool\devtool\dist /E /Y /I
"C:\Program Files\WinRAR\winrar.exe" a -afzip -r tool.zip D tool