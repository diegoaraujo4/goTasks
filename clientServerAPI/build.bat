@echo off
echo Testing Go USD/BRL Exchange Rate Service
echo.

echo Building server...
cd cmd\server
go build -o ..\..\server.exe server.go
if %ERRORLEVEL% neq 0 (
    echo Failed to build server
    exit /b 1
)
cd ..\..

echo Building client...
cd cmd\client  
go build -o ..\..\client.exe client.go
if %ERRORLEVEL% neq 0 (
    echo Failed to build client
    exit /b 1
)
cd ..\..

echo.
echo Built successfully! 
echo.
echo To test:
echo 1. Start server: .\server.exe
echo 2. In another terminal, run client: .\client.exe
echo.
echo Or use Docker:
echo docker-compose up --build
