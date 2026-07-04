@echo off
setlocal EnableExtensions
title PolyForge Dev Menu
set "ROOT=%~dp0"
set "WEB_PORT=8080"

:menu
cls
echo.
echo  ============================================
echo    PolyForge Dev Menu
echo  ============================================
echo.
echo    APP
echo    [1] Start app dev mode        (wails dev)
echo    [2] Build app release         (wails-build.ps1)
echo    [3] Build frontend only       (npm run build)
echo    [4] Stop app dev mode
echo.
echo    WEBSITE
echo    [5] Start website localhost   (http://localhost:%WEB_PORT%)
echo    [6] Open website in browser
echo    [7] Stop website localhost
echo.
echo    [8] Exit
echo.
choice /C 12345678 /N /M "   Select an option: "

if errorlevel 8 goto end
if errorlevel 7 goto webstop
if errorlevel 6 goto webopen
if errorlevel 5 goto webstart
if errorlevel 4 goto appstop
if errorlevel 3 goto frontend
if errorlevel 2 goto appbuild
if errorlevel 1 goto appdev
goto menu

:appdev
echo.
echo   Starting app dev mode in a new window...
start "PolyForge-AppDev" pwsh -NoExit -ExecutionPolicy Bypass -File "%ROOT%scripts\wails-dev.ps1"
timeout /t 2 >nul
goto menu

:appbuild
echo.
echo   Building app release in a new window...
start "PolyForge-AppBuild" pwsh -NoExit -ExecutionPolicy Bypass -File "%ROOT%scripts\wails-build.ps1"
timeout /t 2 >nul
goto menu

:frontend
echo.
echo   Building frontend...
pushd "%ROOT%frontend"
call npm run build
popd
echo.
pause
goto menu

:appstop
echo.
echo   Stopping app dev window...
taskkill /FI "WINDOWTITLE eq PolyForge-AppDev*" /T /F >nul 2>&1
taskkill /IM wails.exe /F >nul 2>&1
echo   Done.
timeout /t 2 >nul
goto menu

:webstart
where php >nul 2>&1
if errorlevel 1 (
  echo.
  echo   PHP was not found on PATH. Install PHP or add it to PATH first.
  pause
  goto menu
)
echo.
echo   Starting website at http://localhost:%WEB_PORT% ...
start "PolyForge-Web" /MIN php -S localhost:%WEB_PORT% -t "%ROOT%website" "%ROOT%website\router.php"
timeout /t 1 >nul
start "" http://localhost:%WEB_PORT%/
goto menu

:webopen
start "" http://localhost:%WEB_PORT%/
goto menu

:webstop
echo.
echo   Stopping website server...
taskkill /FI "WINDOWTITLE eq PolyForge-Web*" /T /F >nul 2>&1
echo   Done.
timeout /t 2 >nul
goto menu

:end
endlocal
exit /b 0
