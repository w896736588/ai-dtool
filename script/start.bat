@echo off
start "" cmd /c ".\build\zhima.exe --IsProd=true & pause"
start http://localhost:17170/