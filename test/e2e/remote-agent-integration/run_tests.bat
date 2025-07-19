@echo off
setlocal enabledelayedexpansion

:: Symphony Remote Agent E2E Test Runner (Windows)
:: This script provides an easy way to run the end-to-end tests on Windows

:: Default values
set TEST_TYPE=all
set VERBOSE=
set TIMEOUT=10m
set DOCKER_CHECK=true
set KUBE_CHECK=true

:: Change to script directory
cd /d "%~dp0"

:parse_args
if "%~1"=="" goto check_prerequisites
if "%~1"=="-t" (
    set TEST_TYPE=%~2
    shift
    shift
    goto parse_args
)
if "%~1"=="--type" (
    set TEST_TYPE=%~2
    shift
    shift
    goto parse_args
)
if "%~1"=="-v" (
    set VERBOSE=-v
    shift
    goto parse_args
)
if "%~1"=="--verbose" (
    set VERBOSE=-v
    shift
    goto parse_args
)
if "%~1"=="-T" (
    set TIMEOUT=%~2
    shift
    shift
    goto parse_args
)
if "%~1"=="--timeout" (
    set TIMEOUT=%~2
    shift
    shift
    goto parse_args
)
if "%~1"=="--skip-docker-check" (
    set DOCKER_CHECK=false
    shift
    goto parse_args
)
if "%~1"=="--skip-kube-check" (
    set KUBE_CHECK=false
    shift
    goto parse_args
)
if "%~1"=="-h" goto show_usage
if "%~1"=="--help" goto show_usage

echo [ERROR] Unknown option: %~1
goto show_usage

:show_usage
echo Usage: %~nx0 [OPTIONS]
echo.
echo Options:
echo   -t, --type TYPE         Test type: all, http, mqtt, cert, reconnect (default: all)
echo   -v, --verbose           Enable verbose output
echo   -T, --timeout DURATION  Test timeout (default: 10m)
echo   --skip-docker-check     Skip Docker availability check
echo   --skip-kube-check       Skip Kubernetes connectivity check
echo   -h, --help              Show this help message
echo.
echo Examples:
echo   %~nx0                      # Run all tests
echo   %~nx0 -t http -v           # Run HTTP tests with verbose output
echo   %~nx0 -t mqtt -T 15m       # Run MQTT tests with 15 minute timeout
echo   %~nx0 --skip-docker-check  # Run tests without checking Docker
goto end

:check_prerequisites
echo [INFO] Checking prerequisites...

:: Check Go
go version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Go is not installed or not in PATH
    exit /b 1
)
echo [SUCCESS] Go found: 
go version

:: Check Docker (if needed)
if "%DOCKER_CHECK%"=="true" (
    if "%TEST_TYPE%"=="all" goto check_docker
    if "%TEST_TYPE%"=="mqtt" goto check_docker
    if "%TEST_TYPE%"=="cert" goto check_docker
    if "%TEST_TYPE%"=="reconnect" goto check_docker
    goto check_kubectl
)

:check_docker
docker --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Docker is not installed or not in PATH
    echo [WARNING] MQTT tests require Docker for the MQTT broker
    exit /b 1
)

docker info >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Docker daemon is not running
    exit /b 1
)
echo [SUCCESS] Docker found and running

:check_kubectl
if "%KUBE_CHECK%"=="false" goto init_modules

kubectl cluster-info >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Cannot connect to Kubernetes cluster
    echo [WARNING] Make sure kubectl is configured and you have access to a cluster
    exit /b 1
)
echo [SUCCESS] Kubernetes connection verified

:init_modules
echo [INFO] Initializing Go modules...
go mod tidy
if errorlevel 1 (
    echo [ERROR] Failed to download Go dependencies
    exit /b 1
)
echo [SUCCESS] Go modules initialized

:run_tests
set TEST_PATH=
set TEST_NAME=

if "%TEST_TYPE%"=="all" (
    set TEST_PATH=./...
    set TEST_NAME=All tests
) else if "%TEST_TYPE%"=="http" (
    set TEST_PATH=./scenarios/http-communication/
    set TEST_NAME=HTTP communication tests
) else if "%TEST_TYPE%"=="mqtt" (
    set TEST_PATH=./scenarios/mqtt-communication/
    set TEST_NAME=MQTT communication tests
) else if "%TEST_TYPE%"=="cert" (
    set TEST_PATH=./scenarios/mqtt-communication/
    set TEST_NAME=MQTT certificate authentication test
    set VERBOSE=%VERBOSE% -run TestMQTTCertificateAuthentication
) else if "%TEST_TYPE%"=="reconnect" (
    set TEST_PATH=./scenarios/mqtt-communication/
    set TEST_NAME=MQTT reconnection test
    set VERBOSE=%VERBOSE% -run TestMQTTReconnection
) else (
    echo [ERROR] Invalid test type: %TEST_TYPE%
    echo [WARNING] Valid types: all, http, mqtt, cert, reconnect
    exit /b 1
)

echo [INFO] Running !TEST_NAME!...
echo [INFO] Command: go test %VERBOSE% -timeout %TIMEOUT% !TEST_PATH!
echo.

:: Run the tests
go test %VERBOSE% -timeout %TIMEOUT% !TEST_PATH!
if errorlevel 1 (
    echo.
    echo [ERROR] !TEST_NAME! failed!
    exit /b 1
)

echo.
echo [SUCCESS] !TEST_NAME! completed successfully!

:cleanup
if not "%TEST_TYPE%"=="all" (
    set /p CLEANUP="Do you want to cleanup test resources? (y/N): "
    if /i "!CLEANUP!"=="y" (
        echo [INFO] Cleaning up any remaining test resources...
        
        :: Clean up any test namespaces
        kubectl delete namespace test-http-ns test-mqtt-ns test-reconnect-ns 2>nul
        
        :: Clean up any test MQTT broker containers
        for /f "tokens=*" %%i in ('docker ps -q --filter "name=test-mqtt-broker" 2^>nul') do docker stop %%i 2>nul
        for /f "tokens=*" %%i in ('docker ps -aq --filter "name=test-mqtt-broker" 2^>nul') do docker rm %%i 2>nul
        
        echo [SUCCESS] Cleanup completed
    )
)

echo [SUCCESS] Test execution completed!

:end
endlocal
