@echo off
 

start ./goservice/build/zhima.exe --IsProd=true  >> ./logs/error.log 2>&1

::start chrome http://localhost:7170/
start http://localhost:7170/