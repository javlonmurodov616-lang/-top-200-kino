@echo off
title Top 200 eng zo'r kinolar
netstat -ano | findstr ":8080" | findstr "LISTENING" >nul 2>&1
if errorlevel 1 (
    start "" "%~dp0top-films.exe"
    timeout /t 3 /nobreak >nul
)
start "" http://localhost:8080
exit