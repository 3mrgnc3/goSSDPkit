#!/usr/bin/env python3
"""
Quick SSDP M-SEARCH test to verify our Go listener is working
"""

import socket


def send_msearch():
    # SSDP M-SEARCH message
    msg = (
        "M-SEARCH * HTTP/1.1\r\n"
        "HOST: 239.255.255.250:1900\r\n"
        'MAN: "ssdp:discover"\r\n'
        "ST: upnp:rootdevice\r\n"
        "MX: 3\r\n\r\n"
    )

    # Create socket
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.settimeout(5)

    try:
        # Send to multicast group
        sock.sendto(msg.encode(), ("239.255.255.250", 1900))
        print("Sent M-SEARCH to multicast group")

        # Try to receive response
        try:
            data, addr = sock.recvfrom(1024)
            print(f"Received response from {addr}:")
            print(data.decode())
        except socket.timeout:
            print("No response received (this is expected in analyze mode)")

    except Exception as e:
        print(f"Error: {e}")
    finally:
        sock.close()


if __name__ == "__main__":
    send_msearch()
