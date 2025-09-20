package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"realtime-chat/internal/hub"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket upgrader configuration
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin (in production, be more restrictive)
		return true
	},
}

// Message represents a chat message
type Message struct {
	Type      string `json:"type"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
	RoomID    string `json:"roomId,omitempty"`
}

// RoomMessage represents a room-specific message
type RoomMessage struct {
	Type      string `json:"type"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
	RoomID    string `json:"roomId"`
}

// RoomAction represents room operations
type RoomAction struct {
	Type     string `json:"type"` // "join", "leave", "create", "list"
	RoomID   string `json:"roomId,omitempty"`
	RoomName string `json:"roomName,omitempty"`
	Username string `json:"username,omitempty"`
}

// HandleWebSocket handles WebSocket connections
func HandleWebSocket(h *hub.Hub, w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Get username from query parameter
	username := r.URL.Query().Get("username")
	if username == "" {
		username = "Anonymous"
	}

	// Create a new client
	client := &hub.Client{
		ID:       generateClientID(),
		Username: username,
		Send:     make(chan []byte, 256),
		Hub:      h,
		RoomID:   "", // Will be set when joining a room
	}

	// Register the client with the hub
	h.Register <- client

	// Start goroutines for reading and writing
	go writePump(client, conn)
	go readPump(client, conn)
}

// readPump pumps messages from the WebSocket connection to the hub
func readPump(c *hub.Client, conn *websocket.Conn) {
	defer func() {
		c.Hub.Unregister <- c
		conn.Close()
	}()

	// Set read deadline and pong handler
	conn.SetReadLimit(512)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Try to parse as a room action first
		var roomAction RoomAction
		if err := json.Unmarshal(messageBytes, &roomAction); err == nil && roomAction.Type != "" {
			// Handle room operations
			handleRoomAction(c, roomAction, conn)
			continue
		}

		// Try to parse as a regular message
		var msg Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Set the username and timestamp
		msg.Username = c.Username
		msg.Timestamp = time.Now().Format(time.RFC3339)
		msg.RoomID = c.RoomID

		// If client is in a room, send to that room
		if c.RoomID != "" {
			roomMessage := RoomMessage{
				Type:      msg.Type,
				Username:  msg.Username,
				Content:   msg.Content,
				Timestamp: msg.Timestamp,
				RoomID:    c.RoomID,
			}

			messageJSON, err := json.Marshal(roomMessage)
			if err != nil {
				log.Printf("Error marshaling room message: %v", err)
				continue
			}

			// Broadcast to the specific room
			c.Hub.RoomManager.BroadcastToRoom(c.RoomID, messageJSON, nil)
		} else {
			// Broadcast to all clients (global chat)
			messageJSON, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				continue
			}

			c.Hub.Broadcast <- messageJSON
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func writePump(c *hub.Client, conn *websocket.Conn) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

// handleRoomAction handles room-related operations
func handleRoomAction(c *hub.Client, action RoomAction, conn *websocket.Conn) {
	switch action.Type {
	case "create":
		// Create a new room
		roomID := c.Hub.RoomManager.CreateRoomAsync(action.RoomName, c.Username)

		// Send room created response
		response := map[string]interface{}{
			"type":     "room_created",
			"roomId":   roomID,
			"roomName": action.RoomName,
			"message":  "Room created successfully",
		}

		responseJSON, _ := json.Marshal(response)
		c.Send <- responseJSON

		// Auto-join the created room
		joinAction := RoomAction{
			Type:   "join",
			RoomID: roomID,
		}
		handleRoomAction(c, joinAction, conn)

	case "join":
		// Join a room
		response := c.Hub.RoomManager.JoinRoomAsync(c, action.RoomID)

		if response.Success {
			c.RoomID = action.RoomID

			// Send join success response
			joinResponse := map[string]interface{}{
				"type":     "room_joined",
				"roomId":   action.RoomID,
				"roomName": response.Room.Name,
				"message":  "Successfully joined room",
			}

			joinResponseJSON, _ := json.Marshal(joinResponse)
			c.Send <- joinResponseJSON
		} else {
			// Send join error response
			errorResponse := map[string]interface{}{
				"type":    "room_error",
				"message": response.Message,
			}

			errorResponseJSON, _ := json.Marshal(errorResponse)
			c.Send <- errorResponseJSON
		}

	case "leave":
		// Leave current room
		if c.RoomID != "" {
			success := c.Hub.RoomManager.LeaveRoomAsync(c, c.RoomID)

			if success {
				c.RoomID = ""

				// Send leave success response
				leaveResponse := map[string]interface{}{
					"type":    "room_left",
					"message": "Successfully left room",
				}

				leaveResponseJSON, _ := json.Marshal(leaveResponse)
				c.Send <- leaveResponseJSON
			}
		}

	case "list":
		// List all available rooms
		rooms := c.Hub.RoomManager.GetRooms()

		roomList := make([]map[string]interface{}, 0, len(rooms))
		for _, room := range rooms {
			roomList = append(roomList, map[string]interface{}{
				"id":          room.ID,
				"name":        room.Name,
				"clientCount": room.GetClientCount(),
				"createdBy":   room.CreatedBy,
				"createdAt":   room.CreatedAt.Format(time.RFC3339),
			})
		}

		response := map[string]interface{}{
			"type":  "room_list",
			"rooms": roomList,
		}

		responseJSON, _ := json.Marshal(response)
		c.Send <- responseJSON
	}
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
