@echo off
REM Wrapper для запуска main.exe через CMD
cd /d %~dp0\..
tmp\main.exe
