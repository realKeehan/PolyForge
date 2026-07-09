@echo off
setlocal EnableExtensions
title PolyForge Dev Menu
set "ROOT=%~dp0"
set "WEB_PORT=8080"

rem Prefer PowerShell 7 (pwsh) but fall back to Windows PowerShell
set "PS=pwsh"
where pwsh >nul 2>&1 || set "PS=powershell"

:menu
cls
set "CURVER=?"
if exist "%ROOT%VERSION" set /p CURVER=<"%ROOT%VERSION"
echo.
echo  ============================================
echo    PolyForge Dev Menu          v%CURVER%
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
echo    [8] Package website for cPanel (7-Zip)
echo.
echo    RELEASE
echo    [9] Set app version           (updates VERSION file)
echo.
echo    [0] Exit
echo.
choice /C 1234567890 /N /M "   Select an option: "

if errorlevel 10 goto end
if errorlevel 9 goto setversion
if errorlevel 8 goto webpack
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
start "PolyForge-AppDev" %PS% -NoExit -ExecutionPolicy Bypass -File "%ROOT%scripts\wails-dev.ps1"
timeout /t 2 >nul
goto menu

:appbuild
echo.
echo   Build app release (v%CURVER%)
echo.
set "BFLAGS="
choice /C YN /N /M "   Customize build flags? [Y/N]: "
if errorlevel 2 goto appbuildrun

echo.
echo   --- wails v2.10 build options (Y = enable) ---
choice /C YN /N /M "   [-UPX]          Compress binary with UPX?        [Y/N]: "
if not errorlevel 2 set "BFLAGS=%BFLAGS% -UPX"
choice /C YN /N /M "   [-nsis]         Build NSIS installer?            [Y/N]: "
if not errorlevel 2 set "BFLAGS=%BFLAGS% -nsis"
echo   [-Obfuscated]   WARNING: breaks the app - garbles the backend method
echo                   names the UI calls, so the build fails with "Unable to
echo                   load installer options from backend". Leave this N.
choice /C YN /N /M "   [-Obfuscated]   Garble bound methods anyway?     [Y/N]: "
if not errorlevel 2 set "BFLAGS=%BFLAGS% -Obfuscated"
choice /C YN /N /M "   [-clean]        Clean bin dir before build?      [Y/N]: "
if not errorlevel 2 set "BFLAGS=%BFLAGS% -clean"
choice /C YN /N /M "   [-trimpath]     Strip file paths from binary?    [Y/N]: "
if not errorlevel 2 set "BFLAGS=%BFLAGS% -trimpath"
choice /C YN /N /M "   [-webview2]     Embed WebView2 runtime? (bigger) [Y/N]: "
if not errorlevel 2 set "BFLAGS=%BFLAGS% -webview2 embed"
choice /C YN /N /M "   [-debug]        Debug build with devtools?       [Y/N]: "
if not errorlevel 2 set "BFLAGS=%BFLAGS% -debug"
choice /C YN /N /M "   [-SkipFrontend] Skip frontend rebuild?           [Y/N]: "
if not errorlevel 2 set "BFLAGS=%BFLAGS% -SkipFrontend"

:appbuildrun
echo.
echo   Launching: wails-build.ps1%BFLAGS%
start "PolyForge-AppBuild" %PS% -NoExit -ExecutionPolicy Bypass -File "%ROOT%scripts\wails-build.ps1" %BFLAGS%
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

:webpack
echo.
echo   Packaging website into a cPanel-ready zip...
%PS% -ExecutionPolicy Bypass -File "%ROOT%scripts\package-website.ps1"
echo.
pause
goto menu

:setversion
echo.
echo   Current version: %CURVER%
echo   The VERSION file feeds both the Go binary and the frontend at build
echo   time, and is compared against api/manifest.json for update prompts.
echo.
set "NEWVER="
set /p NEWVER=  New version (e.g. 5.6.0, blank to cancel):
if not defined NEWVER goto menu
echo Checking format...
echo %NEWVER%| findstr /R "^[0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*$" >nul
if errorlevel 1 (
  echo.
  echo   "%NEWVER%" is not a valid x.y.z version. Nothing changed.
  pause
  goto menu
)
<nul set /p=%NEWVER%> "%ROOT%VERSION"
echo.
echo   VERSION set to %NEWVER%. Rebuild the app to apply it, and remember
echo   to update latestVersion in website/api/manifest.json when releasing.
pause
goto menu

:end
endlocal
exit /b 0
