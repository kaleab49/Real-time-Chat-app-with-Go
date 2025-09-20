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

2. **Open your browser:**
   Navigate to `http://localhost:8080`

3. **Start chatting:**
   - Enter a username
   - Type messages and press Enter to send
   - Open multiple browser tabs to test with multiple users

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
