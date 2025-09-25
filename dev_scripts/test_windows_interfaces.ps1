Write-Host "=== Windows Network Interface Information ===" -ForegroundColor Green
Write-Host

# Get all network adapters
Write-Host "All Network Interfaces:" -ForegroundColor Yellow
Get-NetAdapter | Where-Object {$_.Status -eq "Up"} | Select-Object Name, InterfaceDescription, LinkSpeed | Format-Table -AutoSize

Write-Host
Write-Host "Interface IP Addresses:" -ForegroundColor Yellow
Get-NetIPAddress -AddressFamily IPv4 | Where-Object {$_.AddressState -eq "Preferred"} | Select-Object InterfaceAlias, IPAddress | Format-Table -AutoSize

Write-Host
Write-Host "=== Suggested Interface Names for goSSDPkit ===" -ForegroundColor Green
$wifiAdapters = Get-NetAdapter | Where-Object {$_.Status -eq "Up" -and ($_.InterfaceDescription -like "*Wi-Fi*" -or $_.InterfaceDescription -like "*Wireless*" -or $_.Name -like "*Wi-Fi*")}
if ($wifiAdapters) {
    Write-Host "Wi-Fi Interfaces (recommended for SSDP testing):" -ForegroundColor Cyan
    $wifiAdapters | ForEach-Object {
        $ip = (Get-NetIPAddress -InterfaceAlias $_.Name -AddressFamily IPv4 | Where-Object {$_.AddressState -eq "Preferred"}).IPAddress
        Write-Host "  Interface Name: '$($_.Name)' -> IP: $ip" -ForegroundColor White
        Write-Host "  Try: .\goSSDPkit.exe '$($_.Name)'" -ForegroundColor Gray
    }
} else {
    Write-Host "No Wi-Fi interfaces found. Available interfaces:" -ForegroundColor Yellow
    Get-NetAdapter | Where-Object {$_.Status -eq "Up"} | ForEach-Object {
        $ip = (Get-NetIPAddress -InterfaceAlias $_.Name -AddressFamily IPv4 | Where-Object {$_.AddressState -eq "Preferred"}).IPAddress
        Write-Host "  Interface Name: '$($_.Name)' -> IP: $ip" -ForegroundColor White
        Write-Host "  Try: .\goSSDPkit.exe '$($_.Name)'" -ForegroundColor Gray
    }
}

Write-Host
Write-Host "=== Testing goSSDPkit.exe Interface Detection ===" -ForegroundColor Green
if (Test-Path ".\goSSDPkit.exe") {
    # Test help first
    Write-Host "Testing help:" -ForegroundColor Yellow
    .\goSSDPkit.exe --help 2>&1 | Select-Object -First 5
    
    Write-Host
    Write-Host "Testing interface detection (will fail but show error):" -ForegroundColor Yellow
    $firstInterface = (Get-NetAdapter | Where-Object {$_.Status -eq "Up"} | Select-Object -First 1).Name
    if ($firstInterface) {
        Write-Host "Trying interface: $firstInterface" -ForegroundColor Gray
        # This will likely fail but shows our error handling
        $result = .\goSSDPkit.exe $firstInterface 2>&1 | Select-Object -First 10
        Write-Host $result
    }
} else {
    Write-Host "goSSDPkit.exe not found in current directory" -ForegroundColor Red
}
