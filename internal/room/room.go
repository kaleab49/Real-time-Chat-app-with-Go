package room

import (
	"log"
	"sync"
	"time"
)

// Room represents a chat room with its own clients and message broadcasting
type Room struct {
	ID          string
	Name        string
	Clients     map[*Client]bool
	Broadcast   chan []byte
	Register    chan *Client
	Unregister  chan *Client
	Mutex       sync.RWMutex
	CreatedAt   time.Time
	CreatedBy   string
}

// Client represents a client in a specific room
type Client struct {
	ID       string
	Username string
	Send     chan []byte
	Room     *Room
}

// NewRoom creates a new chat room
func NewRoom(id, name, createdBy string) *Room {
	return &Room{
		ID:         id,
		Name:       name,
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		CreatedAt:  time.Now(),
		CreatedBy:  createdBy,
	}
}

// Run starts the room's message broadcasting loop in a goroutine
func (r *Room) Run() {
	log.Printf("Room '%s' (%s) started", r.Name, r.ID)
	
	for {
		select {
		case client := <-r.Register:
			r.Mutex.Lock()
			r.Clients[client] = true
			r.Mutex.Unlock()
			
			log.Printf("Client %s (%s) joined room '%s'. Room clients: %d", 
				client.ID, client.Username, r.Name, len(r.Clients))
			
			// Send welcome message to the room
			welcomeMsg := []byte(`{"type":"system","message":"` + client.Username + ` joined the room","timestamp":"` + getCurrentTime() + `"}`)
			r.broadcastMessage(welcomeMsg, client)

		case client := <-r.Unregister:
			r.Mutex.Lock()
			if _, ok := r.Clients[client]; ok {
				delete(r.Clients, client)
				close(client.Send)
			}
			r.Mutex.Unlock()
			
			log.Printf("Client %s (%s) left room '%s'. Room clients: %d", 
				client.ID, client.Username, r.Name, len(r.Clients))
			
			// Send goodbye message to the room
			goodbyeMsg := []byte(`{"type":"system","message":"` + client.Username + ` left the room","timestamp":"` + getCurrentTime() + `"}`)
			r.broadcastMessage(goodbyeMsg, nil)

		case message := <-r.Broadcast:
			r.broadcastMessage(message, nil)
		}
	}
}

// broadcastMessage sends a message to all clients in the room
func (r *Room) broadcastMessage(message []byte, sender *Client) {
	r.Mutex.RLock()
	defer r.Mutex.RUnlock()
	
	for client := range r.Clients {
		// Don't send the message back to the sender
		if sender != nil && client == sender {
			continue
		}
		
		select {
		case client.Send <- message:
		default:
			// If client's send channel is full, close the connection
			close(client.Send)
			delete(r.Clients, client)
		}
	}
}

// GetClientCount returns the number of clients in the room
func (r *Room) GetClientCount() int {
	r.Mutex.RLock()
	defer r.Mutex.RUnlock()
	return len(r.Clients)
}

// GetClients returns a list of client usernames in the room
func (r *Room) GetClients() []string {
	r.Mutex.RLock()
	defer r.Mutex.RUnlock()
	
	clients := make([]string, 0, len(r.Clients))
	for client := range r.Clients {
		clients = append(clients, client.Username)
	}
	return clients
}

// getCurrentTime returns the current timestamp
func getCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}
