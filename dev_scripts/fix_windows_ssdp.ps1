Write-Host "=== Windows SSDP Port 1900 Conflict Resolution ===" -ForegroundColor Green
Write-Host

# Check what's using port 1900
Write-Host "Checking what's using port 1900:" -ForegroundColor Yellow
netstat -ano | findstr :1900

Write-Host
Write-Host "SSDP Discovery Service Status:" -ForegroundColor Yellow
Get-Service SSDPSRV | Format-Table Name, Status, StartType -AutoSize

Write-Host
Write-Host "=== Resolution Options ===" -ForegroundColor Green
Write-Host
Write-Host "Option 1 (RECOMMENDED): Stop Windows SSDP service temporarily" -ForegroundColor Cyan
Write-Host "  Run as Administrator:" -ForegroundColor Gray
Write-Host "  net stop SSDPSRV" -ForegroundColor White
Write-Host "  .\goSSDPkit.exe Wi-Fi -t scanner" -ForegroundColor White
Write-Host "  net start SSDPSRV  # (when done testing)" -ForegroundColor White
Write-Host

Write-Host "Option 2: Use different port (less realistic testing)" -ForegroundColor Cyan
Write-Host "  .\goSSDPkit.exe Wi-Fi -t scanner -p 8888" -ForegroundColor White
Write-Host

Write-Host "=== Testing Network Discovery ===" -ForegroundColor Green
Write-Host "After resolving the port conflict, test discovery with:" -ForegroundColor Yellow
Write-Host "1. Open Windows Explorer" -ForegroundColor Gray
Write-Host "2. Look in 'Network' section" -ForegroundColor Gray
Write-Host "3. You should see a new device appear!" -ForegroundColor Gray

Write-Host
Write-Host "=== Attempting to stop SSDP service ===" -ForegroundColor Green
try {
    $currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
    if ($currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
        Write-Host "Running as Administrator - attempting to stop SSDP service..." -ForegroundColor Green
        Stop-Service SSDPSRV -Force
        Write-Host "SSDP service stopped successfully!" -ForegroundColor Green
        Write-Host "You can now run: .\goSSDPkit.exe Wi-Fi -t scanner" -ForegroundColor White
        Write-Host
        Write-Host "Don't forget to restart it later with: Start-Service SSDPSRV" -ForegroundColor Yellow
    } else {
        Write-Host "Not running as Administrator. Please run PowerShell as Administrator and use:" -ForegroundColor Red
        Write-Host "net stop SSDPSRV" -ForegroundColor White
    }
} catch {
    Write-Host "Failed to stop SSDP service: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Please run as Administrator and use: net stop SSDPSRV" -ForegroundColor Yellow
}