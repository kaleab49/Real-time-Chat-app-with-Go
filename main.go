package main

import (
	"fmt"
	"log"
	"net"
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

	// Serve static files
	//  (HTML, CSS, JS)
	http.Handle("/", http.FileServer(http.Dir("./web/")))

	// Get the local IP address
	localIP := getLocalIP()
	
	// Display server information
	fmt.Println("🚀 Real-time Chat Server Starting...")
	fmt.Println("==================================================")
	fmt.Printf("📱 Local Access:    http://localhost:8080\n")
	fmt.Printf("🌐 Network Access:  http://%s:8080\n", localIP)
	fmt.Println("==================================================")
	fmt.Println("💡 Share the network URL with other devices on your local network")
	fmt.Println("🛑 Press Ctrl+C to stop the server")
	fmt.Println("")
	
	log.Printf("Server starting on 0.0.0.0:8080 (accessible from local network)")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}

// getLocalIP returns the local IP address of the machine
func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
