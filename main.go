package main

import (
	"log"
	"net/http"
	"realtime-chat/internal/hub"
	"realtime-chat/internal/websocket"
)

func main() {
	// Create a new hub for managing clients and broadcasting messages
	h := hub.NewHub()
	
	// Start the hub in a goroutine
	go h.Run()

	// WebSocket endpoint
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.HandleWebSocket(h, w, r)
	})

	// Serve static files (HTML, CSS, JS)
	http.Handle("/", http.FileServer(http.Dir("./web/")))

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
