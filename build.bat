@echo off
set "BinDir=%USERPROFILE%\bin"
set "ExePath=%BinDir%\codecrafters.exe"

if "%1"=="install" goto install
if "%1"=="uninstall" goto uninstall
if "%1"=="release" goto release
if "%1"=="test" goto test
goto help

:install
if not exist "%BinDir%" mkdir "%BinDir%"
go build -o "%ExePath%" cmd/codecrafters/main.go
goto end

:uninstall
if exist "%ExePath%" del "%ExePath%"
goto end

:release
for /f "tokens=*" %%i in ('git tag --list "v*"') do set "LastTag=%%i"
if defined LastTag (
    set "CurrentVersion=%LastTag:v=%"
) else (
    set "CurrentVersion=0"
)
set /a "NextVersion=%CurrentVersion%+1"
git tag "v%NextVersion%"
git push origin main "v%NextVersion%"
goto end

:test
go test -v ./...
goto end

:help
echo Usage: build.bat [install^|uninstall^|release^|test]

:end