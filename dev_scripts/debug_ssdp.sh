#!/bin/bash

echo "=== SSDP Troubleshooting Script ==="
echo

echo "1. Checking for processes using port 1900:"
sudo netstat -ulnp | grep 1900 || echo "   No processes found on port 1900"
echo

echo "2. Testing multicast connectivity:"
ping -c 3 239.255.255.250 2>/dev/null || echo "   Multicast ping failed (this is normal)"
echo

echo "3. Available network interfaces:"
ip addr show | grep -E "^[0-9]+:|inet " | grep -v "127.0.0.1"
echo

echo "4. Testing SSDP discovery with nmap (if available):"
if command -v nmap >/dev/null 2>&1; then
    sudo nmap --script broadcast-upnp-info 2>/dev/null || echo "   nmap UPnP scan failed"
else
    echo "   nmap not available"
fi
echo

echo "5. Trying to discover existing SSDP devices:"
timeout 5 python3 -c "
import socket
import select

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
sock.bind(('', 0))
sock.settimeout(3)

msg = '''M-SEARCH * HTTP/1.1\r
HOST: 239.255.255.250:1900\r
MAN: \"ssdp:discover\"\r
ST: ssdp:all\r
MX: 3\r\n\r\n'''

try:
    sock.sendto(msg.encode(), ('239.255.255.250', 1900))
    print('   Sent M-SEARCH, waiting for responses...')
    
    for i in range(10):
        ready = select.select([sock], [], [], 1)
        if ready[0]:
            data, addr = sock.recvfrom(1024)
            print(f'   Response from {addr[0]}: {data[:100]}...')
        else:
            break
    print('   SSDP discovery complete')
except Exception as e:
    print(f'   Error: {e}')
finally:
    sock.close()
" 2>/dev/null || echo "   SSDP discovery test failed"

echo
echo "=== Recommendations ==="
echo "1. Try running: sudo ./goSSDPkit usb0"
echo "2. Verify Windows firewall allows SSDP (port 1900 UDP)"
echo "3. Clear Windows network discovery cache:"
echo "   - Windows: ipconfig /flushdns"  
echo "   - Windows: net stop ssdpsrv && net start ssdpsrv"
echo "4. Test with a real network interface (not loopback)"