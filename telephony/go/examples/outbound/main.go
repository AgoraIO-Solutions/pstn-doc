// Example: outbound
//
// Places an outbound call, waits for call events (answered, bridge, DTMF),
// then hangs up after a short hold.
//
// Usage:
//
//	export CM_HOST="wss://your-cm-host"
//	export AUTH_TOKEN="Basic YOUR_TOKEN"
//	export APPID="your_appid"
//	export TO_NUMBER="+18005551234"
//	export FROM_NUMBER="+15551234567"
//	export SIP="your-lb-host:5081;transport=tls"
//	go run .
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	telephony "github.com/AgoraIO/telephony-go"
)

type handler struct {
	bridged chan struct{}
}

func (h *handler) OnConnected(sessionID string) {
	logEvent("connected", map[string]string{"session_id": sessionID})
}

func (h *handler) OnCallIncoming(call *telephony.Call) bool { return false }

func (h *handler) OnCallRinging(call *telephony.Call) {
	logEvent("call_ringing", map[string]string{"callid": call.CallID})
}

func (h *handler) OnCallAnswered(call *telephony.Call) {
	logEvent("call_answered", map[string]string{"callid": call.CallID})
}

func (h *handler) OnBridgeStart(call *telephony.Call) {
	logEvent("agora_bridge_start", map[string]string{"callid": call.CallID, "channel": call.Channel})
	select {
	case h.bridged <- struct{}{}:
	default:
	}
}

func (h *handler) OnBridgeEnd(call *telephony.Call) {
	logEvent("agora_bridge_end", map[string]string{"callid": call.CallID})
}

func (h *handler) OnCallHangup(call *telephony.Call) {
	logEvent("call_hangup", map[string]string{"callid": call.CallID})
}

func (h *handler) OnError(err error) {
	log.Printf("Error: %v", err)
}

func (h *handler) OnDTMFReceived(call *telephony.Call, digits string) {
	logEvent("dtmf_received", map[string]string{"callid": call.CallID, "digits": digits})
}

func main() {
	cmHost := envOrDefault("CM_HOST", "wss://your-cm-host")
	authToken := requireEnv("AUTH_TOKEN")
	appID := requireEnv("APPID")
	toNumber := requireEnv("TO_NUMBER")
	fromNumber := envOrDefault("FROM_NUMBER", "+15551234567")
	region := envOrDefault("REGION", "AREA_CODE_NA")
	sip := os.Getenv("SIP")

	wsURL := cmHost + "/v1/ws/events"
	clientID := fmt.Sprintf("outbound-example-%d", time.Now().UnixMilli())
	channel := fmt.Sprintf("example_%d", time.Now().UnixMilli())

	client := telephony.NewClient(wsURL, authToken, clientID, appID)
	h := &handler{bridged: make(chan struct{}, 1)}
	client.SetHandler(h)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// 1. Connect
	fmt.Printf("Connecting to %s ...\n", cmHost)
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	// 2. Place outbound call
	fmt.Printf("Dialing %s from %s ...\n", toNumber, fromNumber)
	result, err := client.Dial(ctx, telephony.DialParams{
		To:      toNumber,
		From:    fromNumber,
		Channel: channel,
		UID:     "100",
		Token:   appID, // use appid as token if RTC tokens not enabled
		Region:  region,
		Sip:     sip, // route via load balancer (optional)
		Timeout: "60",
	})
	if err != nil {
		log.Fatalf("Dial failed: %v", err)
	}
	if !result.Success {
		log.Fatalf("Call not successful — no gateways available")
	}
	fmt.Printf("Call placed: callid=%s channel=%s\n", result.CallID, channel)

	// 3. Wait for bridge (call answered + Agora bridge established)
	select {
	case <-h.bridged:
		fmt.Println("Call bridged to Agora channel")
	case <-time.After(30 * time.Second):
		fmt.Println("Timeout waiting for bridge — hanging up")
	}

	// 4. Optional: send DTMF
	fmt.Println("Sending DTMF: 1234#")
	if err := client.SendDTMF(ctx, result.CallID, "1234#"); err != nil {
		log.Printf("SendDTMF failed: %v", err)
	}

	// 5. Hold briefly, then hangup
	time.Sleep(2 * time.Second)
	fmt.Println("Hanging up...")
	if err := client.Hangup(ctx, result.CallID); err != nil {
		log.Printf("Hangup failed: %v", err)
	}

	// Wait for hangup event
	time.Sleep(2 * time.Second)
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
