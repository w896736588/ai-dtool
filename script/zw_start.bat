@echo off
 

start ./goservice/build/zw.exe --IsProd=true  >> ./error.log 2>&1

::start chrome http://localhost:17170/
start http://localhost:17170/