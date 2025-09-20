# Real-time Chat App

A real-time chat application built with Go, showcasing Go's powerful concurrency features including goroutines, channels, and mutexes.

## Features

- **Real-time messaging** using WebSockets
- **Concurrent client handling** with goroutines
- **Thread-safe operations** using mutexes
- **Message broadcasting** to all connected clients
- **Modern web interface** with responsive design
- **Auto-reconnection** on connection loss
- **System notifications** for user join/leave events

## Go Concurrency Features Used

### 1. Goroutines
- Each WebSocket connection runs in its own goroutine
- Separate goroutines for reading and writing messages
- Hub runs in a dedicated goroutine for message broadcasting

### 2. Channels
- `Register` channel for new client connections
- `Unregister` channel for client disconnections
- `Broadcast` channel for message broadcasting
- `Send` channel for each client's outgoing messages

### 3. Mutex (Mutex)
- `sync.RWMutex` for thread-safe access to the clients map
- Prevents race conditions when multiple goroutines access shared data

### 4. Select Statements
- Non-blocking channel operations in the hub's main loop
- Handles multiple channel operations concurrently

## Project Structure

```
realtime-chat/
├── main.go                 # Main application entry point
├── internal/
│   ├── hub/
│   │   └── hub.go         # Client management and message broadcasting
│   └── websocket/
│       └── websocket.go   # WebSocket connection handling
├── web/
│   └── index.html         # Frontend interface
├── go.mod                 # Go module file
└── README.md             # This file
```

## How to Run

1. **Start the server:**
   ```bash
   go run main.go
   ```

2. **Access the application:**
   - **Local Access:** `http://localhost:8080`
   - **Network Access:** `http://[YOUR_IP]:8080` (shown when server starts)

3. **Start chatting:**
   - Enter a username
   - Create or join chat rooms
   - Type messages and press Enter to send
   - Open multiple browser tabs to test with multiple users

## 🌐 Network Access

The server is configured to accept connections from your local network, making it accessible from other devices:

### **From the same computer:**
- Use `http://localhost:8080`

### **From other devices on your network:**
- Use `http://[YOUR_IP]:8080` (IP address shown when server starts)
- Make sure all devices are connected to the same WiFi network
- Some devices might need firewall permission

### **Find your IP address:**
```bash
# Run the helper script
./get-network-info.sh

# Or manually find your IP
hostname -I
```

### **Troubleshooting Network Access:**
- Ensure all devices are on the same network
- Check firewall settings on your computer
- Try accessing from a mobile device on the same WiFi
- Verify the IP address is correct

## Architecture Overview

### Hub Pattern
The application uses a hub pattern where:
- The `Hub` manages all connected clients
- Clients register/unregister through channels
- Messages are broadcast to all clients through channels
- All operations are thread-safe using mutexes

### Concurrency Flow
1. **Client Connection**: Each WebSocket connection spawns two goroutines
   - `readPump`: Reads messages from client and sends to hub
   - `writePump`: Writes messages from hub to client

2. **Message Broadcasting**: The hub runs in its own goroutine
   - Listens for new clients, disconnections, and messages
   - Broadcasts messages to all connected clients concurrently

3. **Thread Safety**: All shared data access is protected by mutexes
   - Client map access is synchronized
   - No race conditions when multiple goroutines access shared state

## Dependencies

- `github.com/gorilla/websocket` - WebSocket implementation
- Standard Go libraries for concurrency (`sync`, `time`)

## Testing

1. Open multiple browser tabs/windows
2. Connect with different usernames
3. Send messages and observe real-time updates
4. Test reconnection by refreshing the page
5. Monitor server logs for connection events

## Future Enhancements

- Multiple chat rooms
- Private messaging
- Message persistence
- User authentication
- File sharing
- Emoji support
- Message history

## Learning Go Concurrency

This project demonstrates several key Go concurrency concepts:

- **Goroutines**: Lightweight threads for concurrent operations
- **Channels**: Communication mechanism between goroutines
- **Select**: Non-blocking channel operations
- **Mutex**: Mutual exclusion for shared data
- **Context**: Graceful shutdown and cancellation

The code is well-commented to help understand how Go's concurrency features work together to create a robust real-time application.
