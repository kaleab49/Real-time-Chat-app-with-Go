package room

import (
	"log"
	"sync"
	"time"
)

// Manager manages all chat rooms and their goroutines
type Manager struct {
	Rooms      map[string]*Room
	Mutex      sync.RWMutex
	CreateRoom chan *Room
	DeleteRoom chan string
	JoinRoom   chan *JoinRequest
	LeaveRoom  chan *LeaveRequest
	Broadcast  chan *BroadcastRequest
}

// JoinRequest represents a request to join a room
type JoinRequest struct {
	Client     interface{} // Will be *hub.Client
	RoomID     string
	Response   chan *JoinResponse
}

// LeaveRequest represents a request to leave a room
type LeaveRequest struct {
	Client   interface{} // Will be *hub.Client
	RoomID   string
	Response chan bool
}

// BroadcastRequest represents a request to broadcast a message to a room
type BroadcastRequest struct {
	RoomID  string
	Message []byte
	Sender  interface{} // Will be *hub.Client
}

// JoinResponse represents the response to a join request
type JoinResponse struct {
	Success bool
	Room    *Room
	Message string
}

// NewManager creates a new room manager
func NewManager() *Manager {
	return &Manager{
		Rooms:      make(map[string]*Room),
		CreateRoom: make(chan *Room),
		DeleteRoom: make(chan string),
		JoinRoom:   make(chan *JoinRequest),
		LeaveRoom:  make(chan *LeaveRequest),
		Broadcast:  make(chan *BroadcastRequest),
	}
}

// Run starts the room manager in a goroutine
func (m *Manager) Run() {
	log.Println("Room Manager started")
	
	for {
		select {
		case room := <-m.CreateRoom:
			m.Mutex.Lock()
			m.Rooms[room.ID] = room
			m.Mutex.Unlock()
			
			// Start the room in its own goroutine
			go room.Run()
			
			log.Printf("Room '%s' (%s) created and started", room.Name, room.ID)

		case roomID := <-m.DeleteRoom:
			m.Mutex.Lock()
			if room, exists := m.Rooms[roomID]; exists {
				// Close all client connections in the room
				for client := range room.Clients {
					close(client.Send)
				}
				delete(m.Rooms, roomID)
				log.Printf("Room '%s' (%s) deleted", room.Name, room.ID)
			}
			m.Mutex.Unlock()

		case req := <-m.JoinRoom:
			m.Mutex.RLock()
			room, exists := m.Rooms[req.RoomID]
			m.Mutex.RUnlock()
			
			if exists {
				// Type assert to get the client
				if client, ok := req.Client.(interface {
					GetID() string
					GetUsername() string
					GetSendChannel() chan []byte
				}); ok {
					// Create a new client for this room
					roomClient := &Client{
						ID:       client.GetID(),
						Username: client.GetUsername(),
						Send:     make(chan []byte, 256),
						Room:     room,
					}
					
					// Register the client with the room
					room.Register <- roomClient
					
					req.Response <- &JoinResponse{
						Success: true,
						Room:    room,
						Message: "Successfully joined room",
					}
				} else {
					req.Response <- &JoinResponse{
						Success: false,
						Room:    nil,
						Message: "Invalid client type",
					}
				}
			} else {
				req.Response <- &JoinResponse{
					Success: false,
					Room:    nil,
					Message: "Room not found",
				}
			}

		case req := <-m.LeaveRoom:
			m.Mutex.RLock()
			room, exists := m.Rooms[req.RoomID]
			m.Mutex.RUnlock()
			
			if exists {
				// Type assert to get the client
				if client, ok := req.Client.(interface {
					GetID() string
				}); ok {
					// Find and remove the client from the room
					room.Mutex.Lock()
					for roomClient := range room.Clients {
						if roomClient.ID == client.GetID() {
							room.Unregister <- roomClient
							break
						}
					}
					room.Mutex.Unlock()
					
					req.Response <- true
				} else {
					req.Response <- false
				}
			} else {
				req.Response <- false
			}

		case req := <-m.Broadcast:
			m.Mutex.RLock()
			room, exists := m.Rooms[req.RoomID]
			m.Mutex.RUnlock()
			
			if exists {
				room.Broadcast <- req.Message
			}
		}
	}
}

// CreateRoom creates a new room and starts it in a goroutine
func (m *Manager) CreateRoomAsync(name, createdBy string) string {
	roomID := generateRoomID()
	room := NewRoom(roomID, name, createdBy)
	
	m.CreateRoom <- room
	return roomID
}

// GetRoom returns a room by ID
func (m *Manager) GetRoom(roomID string) (*Room, bool) {
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()
	room, exists := m.Rooms[roomID]
	return room, exists
}

// GetRooms returns a list of all rooms
func (m *Manager) GetRooms() []*Room {
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()
	
	rooms := make([]*Room, 0, len(m.Rooms))
	for _, room := range m.Rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// GetRoomCount returns the number of active rooms
func (m *Manager) GetRoomCount() int {
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()
	return len(m.Rooms)
}

// JoinRoomAsync joins a client to a room
func (m *Manager) JoinRoomAsync(client interface{}, roomID string) *JoinResponse {
	response := make(chan *JoinResponse)
	req := &JoinRequest{
		Client:   client,
		RoomID:   roomID,
		Response: response,
	}
	
	m.JoinRoom <- req
	return <-response
}

// LeaveRoomAsync removes a client from a room
func (m *Manager) LeaveRoomAsync(client interface{}, roomID string) bool {
	response := make(chan bool)
	req := &LeaveRequest{
		Client:   client,
		RoomID:   roomID,
		Response: response,
	}
	
	m.LeaveRoom <- req
	return <-response
}

// BroadcastToRoom sends a message to a specific room
func (m *Manager) BroadcastToRoom(roomID string, message []byte, sender interface{}) {
	req := &BroadcastRequest{
		RoomID:  roomID,
		Message: message,
		Sender:  sender,
	}
	
	m.Broadcast <- req
}

// generateRoomID generates a unique room ID
func generateRoomID() string {
	return "room_" + time.Now().Format("20060102150405") + "_" + randomString(6)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
