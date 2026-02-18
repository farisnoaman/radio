package websocket

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHub_Broadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Mock client
	client := &Client{
		hub:  hub,
		send: make(chan []byte, 256),
	}
	
	// Register client
	hub.register <- client
	
	// Wait for registration
	time.Sleep(100 * time.Millisecond)
	
	// Broadcast message
	msg := map[string]string{"test": "message"}
	hub.Broadcast(msg)
	
	// Verify message received
	select {
	case received := <-client.send:
		var decoded map[string]string
		err := json.Unmarshal(received, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, "message", decoded["test"])
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for broadcast message")
	}
	
	// Unregister client
	hub.unregister <- client
	time.Sleep(100 * time.Millisecond)
	
	// Broadcast another message, should not block or panic
	hub.Broadcast(msg)
}
