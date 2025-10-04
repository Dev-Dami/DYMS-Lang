@echo off
REM Batch build script for DYMS (Holy-Go) project

set BINARY_NAME=dyms
set MAIN_FILE=main.go
set BUILD_DIR=build
set BINARY_PATH=%BUILD_DIR%\%BINARY_NAME%.exe

if "%1"=="" (
    set COMMAND=build
) else (
    set COMMAND=%1
)

if "%COMMAND%"=="build" goto build
if "%COMMAND%"=="clean" goto clean
if "%COMMAND%"=="test" goto test
if "%COMMAND%"=="deps" goto deps
if "%COMMAND%"=="repl" goto repl
if "%COMMAND%"=="help" goto help
goto help

:build
echo Building DYMS...
if not exist %BUILD_DIR% mkdir %BUILD_DIR%
go build -o %BINARY_PATH% %MAIN_FILE%
if %errorlevel% equ 0 (
    echo Built %BINARY_NAME% successfully!
) else (
    echo Build failed!
    exit /b 1
)
goto end

:clean
echo Cleaning build directory...
if exist %BUILD_DIR% rmdir /s /q %BUILD_DIR%
echo Cleaned build directory
goto end

:test
call :build
echo Running tests...
echo Running basic features test:
%BINARY_PATH% test\01_basic_features.dy
echo.
echo Running math comprehensive benchmark:
%BINARY_PATH% test\12_math_comprehensive_benchmark.dy
echo.
echo Running sub-150ms performance test:
%BINARY_PATH% test\19_sub150_test.dy
echo.
echo Running fast loop test:
%BINARY_PATH% test\18_fast_loop_test.dy
echo.
echo Running new features test:
echo   - Break/Continue test:
%BINARY_PATH% test\21_simple_test.dy
echo   - Increment/Decrement test:
%BINARY_PATH% test\22_inc_dec_test.dy
echo   - Try/Catch test:
%BINARY_PATH% test\23_try_catch_test.dy
goto end

:deps
echo Installing dependencies...
go mod tidy
go mod download
echo Dependencies installed!
goto end

:repl
call :build
%BINARY_PATH%
goto end

:help
echo DYMS Build Script Commands:
echo.
echo   build     Build the DYMS executable
echo   clean     Clean build artifacts  
echo   test      Run all test files
echo   deps      Install Go dependencies
echo   repl      Start the DYMS REPL
echo   help      Show this help message
echo.
echo Examples:
echo   build.bat build
echo   build.bat test
echo   build.bat repl
goto end

:end