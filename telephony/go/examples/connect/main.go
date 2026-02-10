// Example: connect
//
// Verifies that your credentials work by connecting to the CM WebSocket,
// registering, and printing the session ID.
//
// Usage:
//
//	export CM_HOST="wss://sipcm.agora.io"
//	export AUTH_TOKEN="Basic YOUR_TOKEN"
//	export APPID="your_appid"
//	go run .
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	telephony "github.com/AgoraIO/telephony-go"
)

type handler struct{}

func (h *handler) OnConnected(sessionID string)                      { fmt.Printf("Connected: session=%s\n", sessionID) }
func (h *handler) OnCallIncoming(call *telephony.Call) bool          { return false }
func (h *handler) OnCallRinging(call *telephony.Call)                {}
func (h *handler) OnCallAnswered(call *telephony.Call)               {}
func (h *handler) OnBridgeStart(call *telephony.Call)                {}
func (h *handler) OnBridgeEnd(call *telephony.Call)                  {}
func (h *handler) OnCallHangup(call *telephony.Call)                 {}
func (h *handler) OnError(err error)                                 { log.Printf("Error: %v", err) }

func main() {
	cmHost := envOrDefault("CM_HOST", "wss://sipcm.agora.io")
	authToken := requireEnv("AUTH_TOKEN")
	appID := requireEnv("APPID")

	wsURL := cmHost + "/v1/ws/events"
	clientID := fmt.Sprintf("connect-example-%d", time.Now().UnixMilli())

	client := telephony.NewClient(wsURL, authToken, clientID, appID)
	client.SetHandler(&handler{})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	fmt.Printf("Connecting to %s ...\n", cmHost)
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	fmt.Println("OK — authenticated and registered successfully")
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("Required env var %s is not set", key)
	}
	return v
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
