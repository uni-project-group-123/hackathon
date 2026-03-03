@echo off
REM Sentinel Build Script for Windows
REM Run from the backend-go directory

echo Building Sentinel...

REM Download dependencies
echo Downloading dependencies...
go mod download

REM Build for current architecture (x64)
echo Building binary...
go build -ldflags="-s -w" -o sentinel.exe .

echo Build complete: sentinel.exe
