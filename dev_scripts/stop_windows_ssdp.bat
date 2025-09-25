@echo off
echo === Stopping Windows SSDP Service for goSSDPkit ===
echo.

echo Checking SSDP service status...
sc query SSDPSRV

echo.
echo Stopping SSDP Discovery service...
net stop SSDPSRV

echo.
echo Service stopped. You can now run goSSDPkit.
echo.
echo To restart the service later, run:
echo net start SSDPSRV
echo.
echo Press any key to continue...
pause >nul