#!/bin/bash

echo "🌐 Network Information for Real-time Chat App"
echo "=============================================="
echo ""

# Get local IP address
LOCAL_IP=$(hostname -I | awk '{print $1}')
echo "📱 Your Local IP Address: $LOCAL_IP"
echo ""

# Display access URLs
echo "🔗 Access URLs:"
echo "   Local Access:    http://localhost:8080"
echo "   Network Access:  http://$LOCAL_IP:8080"
echo ""

# Check if port 8080 is open
echo "🔍 Checking if port 8080 is available..."
if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null ; then
    echo "✅ Port 8080 is currently in use (server is running)"
else
    echo "❌ Port 8080 is not in use (server is not running)"
    echo "   Run 'go run main.go' to start the server"
fi
echo ""

# Display network interfaces
echo "📡 Available Network Interfaces:"
ip addr show | grep -E "inet [0-9]" | grep -v "127.0.0.1" | awk '{print "   " $2}' | cut -d'/' -f1
echo ""

echo "💡 Instructions:"
echo "   1. Start the server: go run main.go"
echo "   2. Share the network URL with other devices on your local network"
echo "   3. Make sure all devices are connected to the same WiFi network"
echo "   4. Some devices might need to allow the connection through firewall"
