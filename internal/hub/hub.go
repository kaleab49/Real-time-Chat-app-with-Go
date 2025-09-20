package hub

import (
	"log"
	"sync"
	"time"
)

// Client represents a connected WebSocket client
type Client struct {
	ID       string
	Username string
	Send     chan []byte
	Hub      *Hub
}


// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Channel for broadcasting messages to all clients
	broadcast chan []byte

	// Channel for registering new clients
	Register chan *Client

	// Channel for unregistering clients
	Unregister chan *Client

	// Channel for broadcasting messages
	Broadcast chan []byte

	// Mutex for thread-safe operations
	mutex sync.RWMutex
}

// NewHub creates a new hub instance
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan []byte),
	}
}

// Run starts the hub and handles client registration/unregistration and message broadcasting
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			
			log.Printf("Client %s (%s) connected. Total clients: %d", 
				client.ID, client.Username, len(h.clients))
			
			// Send welcome message
			welcomeMsg := []byte(`{"type":"system","message":"` + client.Username + ` joined the chat","timestamp":"` + getCurrentTime() + `"}`)
			h.broadcastMessage(welcomeMsg, client)

		case client := <-h.Unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mutex.Unlock()
			
			log.Printf("Client %s (%s) disconnected. Total clients: %d", 
				client.ID, client.Username, len(h.clients))
			
			// Send goodbye message
			goodbyeMsg := []byte(`{"type":"system","message":"` + client.Username + ` left the chat","timestamp":"` + getCurrentTime() + `"}`)
			h.broadcastMessage(goodbyeMsg, nil)

		case message := <-h.Broadcast:
			h.broadcastMessage(message, nil)
		}
	}
}

// broadcastMessage sends a message to all connected clients
func (h *Hub) broadcastMessage(message []byte, sender *Client) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	for client := range h.clients {
		// Don't send the message back to the sender
		if sender != nil && client == sender {
			continue
		}
		
		select {
		case client.Send <- message:
		default:
			// If client's send channel is full, close the connection
			close(client.Send)
			delete(h.clients, client)
		}
	}
}

// GetClientCount returns the current number of connected clients
func (h *Hub) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// getCurrentTime returns the current timestamp
func getCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}
