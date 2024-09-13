@echo off
 

start ./goservice/build/zhima.exe --IsProd=true  >> ./error.log 2>&1


start chrome http://localhost:7170/