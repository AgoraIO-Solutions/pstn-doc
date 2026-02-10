// Example: inbound
//
// Subscribes to a DID (phone number) and waits for incoming calls.
// When a call arrives, it auto-accepts with Agora credentials and logs
// all lifecycle events until hangup.
//
// Usage:
//
//	export CM_HOST="wss://sipcm.agora.io"
//	export AUTH_TOKEN="Basic YOUR_TOKEN"
//	export APPID="your_appid"
//	export DID="18005551234"
//	go run .
//
// Then call the DID from a phone or trigger a loopback by dialing it
// via the outbound API.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	telephony "github.com/AgoraIO/telephony-go"
)

type handler struct {
	client *telephony.Client
	appID  string
	done   chan struct{}
}

func (h *handler) OnConnected(sessionID string) {
	logEvent("connected", map[string]string{"session_id": sessionID})
}

func (h *handler) OnCallIncoming(call *telephony.Call) bool {
	logEvent("call_incoming", map[string]string{
		"callid": call.CallID, "from": call.From, "to": call.To,
	})

	// Accept the call asynchronously
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		channel := fmt.Sprintf("inbound_%s", call.CallID[:8])
		fmt.Printf("Accepting call %s → channel=%s\n", call.CallID, channel)

		err := h.client.Accept(ctx, call.CallID, telephony.AcceptParams{
			Token:   h.appID, // use appid as token if RTC tokens not enabled
			Channel: channel,
			UID:     "200",
			AppID:   h.appID,
		})
		if err != nil {
			log.Printf("Accept failed: %v", err)
		}
	}()

	return true // claim the call
}

func (h *handler) OnCallRinging(call *telephony.Call) {
	logEvent("call_ringing", map[string]string{"callid": call.CallID})
}

func (h *handler) OnCallAnswered(call *telephony.Call) {
	logEvent("call_answered", map[string]string{"callid": call.CallID})
}

func (h *handler) OnBridgeStart(call *telephony.Call) {
	logEvent("agora_bridge_start", map[string]string{"callid": call.CallID, "channel": call.Channel})
}

func (h *handler) OnBridgeEnd(call *telephony.Call) {
	logEvent("agora_bridge_end", map[string]string{"callid": call.CallID})
}

func (h *handler) OnCallHangup(call *telephony.Call) {
	logEvent("call_hangup", map[string]string{"callid": call.CallID})
	select {
	case h.done <- struct{}{}:
	default:
	}
}

func (h *handler) OnError(err error) {
	log.Printf("Error: %v", err)
}

func (h *handler) OnDTMFReceived(call *telephony.Call, digits string) {
	logEvent("dtmf_received", map[string]string{"callid": call.CallID, "digits": digits})
}

func main() {
	cmHost := envOrDefault("CM_HOST", "wss://sipcm.agora.io")
	authToken := requireEnv("AUTH_TOKEN")
	appID := requireEnv("APPID")
	did := requireEnv("DID")

	wsURL := cmHost + "/v1/ws/events"
	clientID := fmt.Sprintf("inbound-example-%d", time.Now().UnixMilli())

	client := telephony.NewClient(wsURL, authToken, clientID, appID)
	h := &handler{client: client, appID: appID, done: make(chan struct{}, 1)}
	client.SetHandler(h)

	// Subscribe to the DID before connecting
	client.SetSubscribeNumbers([]string{did})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	fmt.Printf("Connecting to %s ...\n", cmHost)
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	fmt.Printf("Subscribed to DID %s — waiting for incoming calls\n", did)
	fmt.Println("Press Ctrl+C to exit, or wait for a call to complete")

	// Wait for either a hangup event or Ctrl+C
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-h.done:
		fmt.Println("Call completed")
	case <-sigCh:
		fmt.Println("\nShutting down...")
	}

	fmt.Println("Done")
}

func logEvent(event string, data map[string]string) {
	data["event"] = event
	data["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	b, _ := json.Marshal(data)
	fmt.Println(string(b))
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
