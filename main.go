package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/go-stomp/stomp"
)

// Define constants
const (
	webSocketPath = "/ws"
	stompQueue    = "/queue/test"
)

// var upgrader = websocket.NewUpgrader()

func main() {
	http.HandleFunc(webSocketPath, handleWebSocket)

	serverAddr := ":8080"
	log.Printf("Server starting on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer conn.Close(websocket.StatusAbnormalClosure, "abnormal closure from handleWebSocket")

	log.Println("WebSocket connection established")
	processStomp(conn)
}

func processStomp(conn *websocket.Conn) {
	// Wrap WebSocket connection for STOMP
	stompConn, err := stomp.Connect(&stompWebSocketConn{conn}, stomp.ConnOpt.HeartBeat(10*time.Second, 10*time.Second))
	if err != nil {
		log.Printf("Failed to establish STOMP connection: %v", err)
		return
	}
	defer stompConn.Disconnect()

	log.Println("STOMP connection established")

	// Subscribe to a queue
	sub, err := stompConn.Subscribe(stompQueue, stomp.AckAuto)
	if err != nil {
		log.Printf("Failed to subscribe to queue: %v", err)
		return
	}
	defer sub.Unsubscribe()

	// Listen for messages
	for {
		msg := <-sub.C
		if msg.Err != nil {
			log.Printf("Error receiving message: %v", msg.Err)
			break
		}

		log.Printf("Received message: %s", string(msg.Body))
	}
}

// stompWebSocketConn implements stomp.Conn interface for WebSocket
type stompWebSocketConn struct {
	conn *websocket.Conn
}

func (s *stompWebSocketConn) Read(p []byte) (int, error) {
	messageType, data, err := s.conn.Read(context.Background())
	if err != nil {
		return 0, err
	}
	fmt.Println("MessageType:", messageType)
	return copy(p, data), nil
}

func (s *stompWebSocketConn) Write(p []byte) (int, error) {
	err := s.conn.Write(context.Background(), websocket.MessageText, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (s *stompWebSocketConn) Close() error {
	return s.conn.Close(websocket.StatusBadGateway, "Simply closing")
}
