@echo off
start "" cmd /c ".\build\zw.exe --IsProd=true & pause"
start http://localhost:17170/