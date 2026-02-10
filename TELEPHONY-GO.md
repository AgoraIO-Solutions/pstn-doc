# Telephony Go SDK

Go SDK for the Agora SIP Call Manager WebSocket API. Places, receives, and manages SIP calls via the Telephony WebSocket interface.

**SDK source:** [`telephony/go/client.go`](telephony/go/client.go)
**Runnable examples:** [`telephony/go/examples/`](telephony/go/examples/)

## Prerequisites

- **Go 1.21+** — [install](https://go.dev/doc/install)
- **Agora App ID** — from your [Agora Console](https://console.agora.io) project
- **Auth token** — provided by Agora when your App ID is provisioned for SIP/PSTN
- **CM WebSocket URL** — provided during provisioning (e.g. `wss://your-cm-host`)
- **For inbound calls:** a DID (phone number) provisioned on the CM

## Installation

The SDK is a single file — [`telephony/go/client.go`](telephony/go/client.go). Copy it into your project or use it as a Go module:

```bash
# Option 1: Use as a module (update module path as needed)
go get github.com/AgoraIO/telephony-go

# Option 2: Copy the SDK file directly
cp telephony/go/client.go your-project/telephony/
```

Dependency: [gorilla/websocket](https://github.com/gorilla/websocket) v1.5.1+

## Quick Start

The fastest way to verify your setup:

```bash
cd telephony/go/examples

# 1. Set your credentials
export CM_HOST="wss://your-cm-host"
export AUTH_TOKEN="Basic YOUR_TOKEN"
export APPID="your_appid"

# 2. Test connection
go run ./connect/

# 3. Place an outbound call
export TO_NUMBER="+18005551234"
export FROM_NUMBER="+15551234567"
go run ./outbound/

# 4. Listen for inbound calls on a DID
export DID="18005551234"
go run ./inbound/
```

### Minimal Code Example

```go
package main

import (
    "context"
    "fmt"
    "log"

    telephony "github.com/AgoraIO/telephony-go"
)

func main() {
    client := telephony.NewClient(
        "wss://your-cm-host/v1/ws/events",
        "Basic YOUR_TOKEN",
        "my-client-1",
        "YOUR_APPID",
    )
    client.SetHandler(&MyHandler{}) // see Complete Handler Example below

    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Place an outbound call
    result, err := client.Dial(ctx, telephony.DialParams{
        To:      "+18005551234",
        From:    "+15551234567",
        Channel: "my-channel",
        UID:     "100",
        Token:   "agora-rtc-token",
        Region:  "AREA_CODE_NA",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Call placed: callid=%s\n", result.CallID)

    // Events arrive automatically via WebSocket (answered, bridge, dtmf, hangup)
    // Send DTMF and hangup when ready
    client.SendDTMF(ctx, result.CallID, "1234#")
    client.Hangup(ctx, result.CallID)
}
```

See [`examples/outbound/main.go`](telephony/go/examples/outbound/main.go) for a complete, runnable version with event logging.

---

## API Reference

### Constructor

#### `NewClient(wsURL, authToken, clientID, appID string) *Client`

Creates a new WebSocket client. Does not connect — call `Connect()` to establish the connection.

| Parameter | Description |
|-----------|-------------|
| `wsURL` | WebSocket endpoint (e.g. `wss://your-cm-host/v1/ws/events`) |
| `authToken` | Auth token from CM config (e.g. `Basic LkP3sQ8j...`) |
| `clientID` | Unique identifier for this client instance |
| `appID` | Agora App ID. Use `"MULTI"` for multi-appid mode (see [MULTI Mode](#multi-appid-mode)) |

```go
client := telephony.NewClient(wsURL, "Basic mytoken", "client-1", "20b7c51f...")
```

---

### Connection

#### `Connect(ctx context.Context) error`

Connects to the WebSocket server, performs the handshake (connected → register → registered), and starts the event loop. Blocks until registration completes or fails.

```go
ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
defer cancel()
if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}
```

**Auto-reconnect:** If the connection drops after a successful `Connect()`, the SDK automatically reconnects with exponential backoff (1s → 2s → 4s → ... → 30s max). Call state is preserved across reconnects. `OnError` fires on each disconnect.

#### `Close() error`

Gracefully closes the connection. Unblocks any pending commands with a "connection lost" error. Stops the reconnect loop. Safe to call multiple times.

```go
defer client.Close()
```

#### `IsConnected() bool`

Returns whether the client currently has an active WebSocket connection.

---

### Calling

#### `Dial(ctx context.Context, params DialParams) (*DialResult, error)`

Places an outbound call. The CM selects a gateway (or uses the `Sip`/`SipDomain` override) and forwards the call. Returns the call ID on success.

```go
result, err := client.Dial(ctx, telephony.DialParams{
    To:      "+18005551234",
    From:    "+15551234567",
    Channel: "my-channel",
    UID:     "100",
    Token:   "agora-rtc-token",
    Region:  "AREA_CODE_NA",
    Timeout: "60",
})
if err != nil {
    log.Fatal(err)
}
if !result.Success {
    log.Fatal("no gateways available")
}
fmt.Println("CallID:", result.CallID)
```

**DialParams:**

| Field | Required | Description |
|-------|----------|-------------|
| `To` | yes | Destination phone number (E.164 format) |
| `From` | yes | Caller ID phone number |
| `Channel` | yes | Agora channel name for the bridge |
| `UID` | yes | Agora user ID |
| `Token` | yes | Agora RTC token |
| `Region` | yes | Gateway region (`AREA_CODE_NA`, `AREA_CODE_EU`, `AREA_CODE_AS`, etc.) |
| `Timeout` | no | Call timeout in seconds (default: gateway default) |
| `Sip` | no | SIP URI — routes through a load balancer (e.g. `host:5081;transport=tls`) |
| `SipDomain` | no | Force a specific gateway by domain |
| `AppID` | no | Required in [MULTI mode](#multi-appid-mode) — Agora App ID for this call |

**DialResult:**

| Field | Type | Description |
|-------|------|-------------|
| `Success` | `bool` | Whether the gateway accepted the call |
| `CallID` | `string` | Unique call identifier for subsequent operations |
| `Data` | `map[string]interface{}` | Raw response data from the CM |

#### `Accept(ctx context.Context, callid string, creds AcceptParams) error`

Accepts an inbound call received via `OnCallIncoming`. Provides Agora credentials for the bridge.

```go
func (h *MyHandler) OnCallIncoming(call *telephony.Call) bool {
    go func() {
        err := h.client.Accept(context.Background(), call.CallID, telephony.AcceptParams{
            Token:   "agora-rtc-token",
            Channel: "inbound-channel",
            UID:     "200",
            AppID:   "20b7c51f...",
        })
        if err != nil {
            log.Printf("Accept failed: %v", err)
        }
    }()
    return true // claim the call
}
```

**AcceptParams:**

| Field | Required | Description |
|-------|----------|-------------|
| `Token` | yes | Agora RTC token |
| `Channel` | yes | Agora channel name |
| `UID` | yes | Agora user ID |
| `AppID` | no | Required in [MULTI mode](#multi-appid-mode) |
| `WebhookURL` | no | Override webhook URL (auto-injected if omitted) |
| `SDKOptions` | no | Agora SDK options JSON string |
| `AudioScenario` | no | Agora audio scenario |

#### `Reject(ctx context.Context, callid string, reason string) error`

Rejects an inbound call. The gateway receives a 404 response.

```go
err := client.Reject(ctx, callID, "busy")
```

#### `Hangup(ctx context.Context, callid string) error`

Ends an active call. Automatically uses `endcall` for outbound calls and `hangup` for inbound calls.

```go
// Hangup outbound call (sends "endcall" to CM → gateway)
err := client.Hangup(ctx, outboundCallID)

// Hangup inbound accepted call (sends "hangup" to gateway directly)
err := client.Hangup(ctx, inboundCallID)
```

The call is removed from `GetActiveCalls()` after a successful hangup.

#### `SendDTMF(ctx context.Context, callid string, digits string) error`

Sends DTMF tones on an active call. Works on both outbound and inbound accepted calls. Valid digits: `0-9`, `*`, `#`.

```go
// Send DTMF on outbound call
err := client.SendDTMF(ctx, outboundCallID, "1234#")

// Send DTMF on inbound accepted call
err := client.SendDTMF(ctx, inboundCallID, "5678*")
```

The gateway echoes back a `dtmf_received` event via `OnDTMFReceived`.

#### `Bridge(ctx context.Context, callid string, creds BridgeParams) error`

Bridges an active call to an Agora channel. Used for re-bridging after `Unbridge`.

```go
err := client.Bridge(ctx, callID, telephony.BridgeParams{
    Token:   "agora-rtc-token",
    Channel: "new-channel",
    UID:     "300",
})
```

**BridgeParams:**

| Field | Required | Description |
|-------|----------|-------------|
| `Token` | yes | Agora RTC token |
| `Channel` | yes | Agora channel name |
| `UID` | yes | Agora user ID |
| `AppID` | no | Required in [MULTI mode](#multi-appid-mode) |
| `SDKOptions` | no | Agora SDK options JSON string |
| `AudioScenario` | no | Agora audio scenario |

#### `Unbridge(ctx context.Context, callid string) error`

Removes the Agora channel bridge from the call. The SIP call stays active.

```go
err := client.Unbridge(ctx, callID)
```

#### `Transfer(ctx context.Context, callid string, destination string, leg string) error`

Transfers a call to another destination.

```go
// Transfer the call
err := client.Transfer(ctx, callID, "+18001234567", "")

// Transfer a specific leg
err := client.Transfer(ctx, callID, "+18001234567", "B")
```

| Parameter | Description |
|-----------|-------------|
| `callid` | Call to transfer |
| `destination` | Target phone number |
| `leg` | Which leg to transfer (`""` = default, `"A"` or `"B"`) |

---

### Phone Number Subscription

Control which inbound calls your client receives. Without subscriptions, the client is catch-all (receives all events for its appid).

#### `SetSubscribeNumbers(numbers []string)`

Set phone numbers to subscribe to **before** calling `Connect()`. Numbers are sent as part of the registration message.

```go
client.SetSubscribeNumbers([]string{"+18005551234", "+18005559876"})
client.Connect(ctx) // subscription sent during registration
```

#### `Subscribe(ctx context.Context, numbers []string) error`

Update subscriptions on a **live** connection. Replaces the previous subscription list.

```go
// Subscribe to new numbers (replaces previous)
err := client.Subscribe(ctx, []string{"+18005551234"})
```

Numbers are normalized to digits on the server (leading `+` and non-digit characters are stripped).

---

### Call State

#### `GetActiveCalls() []*Call`

Returns all currently tracked calls.

```go
for _, call := range client.GetActiveCalls() {
    fmt.Printf("callid=%s state=%s direction=%s\n", call.CallID, call.State, call.Direction)
}
```

**Call struct:**

| Field | Type | Description |
|-------|------|-------------|
| `CallID` | `string` | Unique call identifier |
| `State` | `string` | `incoming`, `ringing`, `answered`, `bridged`, `unbridged`, `hangup` |
| `Direction` | `string` | `outbound` or `inbound` |
| `From` | `string` | Caller number |
| `To` | `string` | Destination number |
| `Channel` | `string` | Agora channel name |
| `UID` | `string` | Agora user ID |
| `AppID` | `string` | Agora App ID (set from Dial/Accept params or incoming events) |

---

### Event Handlers

#### `SetHandler(handler EventHandler)`

Set the event handler before calling `Connect()`. The handler receives all call lifecycle events.

```go
client.SetHandler(&MyHandler{})
```

#### `EventHandler` Interface

All methods are required. Events fire asynchronously from the WebSocket read loop.

```go
type EventHandler interface {
    OnConnected(sessionID string)
    OnCallIncoming(call *Call) bool   // return true to claim the call
    OnCallRinging(call *Call)
    OnCallAnswered(call *Call)
    OnBridgeStart(call *Call)
    OnBridgeEnd(call *Call)
    OnCallHangup(call *Call)
    OnError(err error)
}
```

#### `DTMFHandler` Interface (Optional)

Implement this alongside `EventHandler` to receive DTMF events. The SDK checks for this interface at runtime.

```go
type DTMFHandler interface {
    OnDTMFReceived(call *Call, digits string)
}
```

### Event Reference

Events arrive via WebSocket from the gateway's webhook. The CM routes webhook POSTs to the WS client that owns the call.

| Event | Handler | WS Event Name | When it fires |
|-------|---------|---------------|---------------|
| Connected | `OnConnected(sessionID)` | — | After successful registration |
| Incoming | `OnCallIncoming(call) bool` | `call_incoming` | Inbound call on a subscribed DID. Return `true` to claim, `false` to ignore (call removed from state). |
| Ringing | `OnCallRinging(call)` | `call_ringing` | Outbound call is ringing at destination |
| Answered | `OnCallAnswered(call)` | `call_answered` | Call answered by remote party |
| Bridge Start | `OnBridgeStart(call)` | `agora_bridge_start` | Agora channel bridge established |
| Bridge End | `OnBridgeEnd(call)` | `agora_bridge_end` | Agora channel bridge removed |
| Hangup | `OnCallHangup(call)` | `call_hangup` | Call ended. Call removed from `GetActiveCalls()`. |
| DTMF | `OnDTMFReceived(call, digits)` | `dtmf_received` | DTMF tones received (requires `DTMFHandler` interface) |
| Error | `OnError(err)` | — | Connection error or read failure |

**Event sequence — outbound call:**
```
Dial() → OnCallAnswered → OnBridgeStart → [OnDTMFReceived] → OnCallHangup
```

**Event sequence — inbound call:**
```
OnCallIncoming → Accept() → OnCallAnswered → OnBridgeStart → [OnDTMFReceived] → OnCallHangup
```

---

## MULTI-AppID Mode

MULTI mode allows a single WebSocket client to manage calls across multiple Agora App IDs. This is useful for platforms that serve multiple customers, each with their own App ID.

### Setup

Register with `appID: "MULTI"` and provide a corresponding `authorization_MULTI` token in the CM config:

```go
client := telephony.NewClient(
    "wss://your-cm-host/v1/ws/events",
    "Basic MULTI_TOKEN",
    "my-multi-client",
    "MULTI",                    // special appid — enables multi-appid mode
)
```

CM config (`config/{env}/auth`):
```
authorization_MULTI=Basic MULTI_TOKEN
```

### Usage

In MULTI mode, **every command must include `AppID`**. The server rejects commands without it.

```go
// Outbound call — AppID required in DialParams
result, err := client.Dial(ctx, telephony.DialParams{
    To:      "+18005551234",
    From:    "+15551234567",
    Channel: "multi-ch",
    UID:     "100",
    Token:   "agora-rtc-token",
    Region:  "AREA_CODE_NA",
    AppID:   "20b7c51ff4c644ab80cf5a4e646b0537",    // required in MULTI mode
})

// Accept inbound — AppID required in AcceptParams
err := client.Accept(ctx, callID, telephony.AcceptParams{
    Token:   "agora-rtc-token",
    Channel: "inbound-ch",
    UID:     "200",
    AppID:   "20b7c51ff4c644ab80cf5a4e646b0537",    // required in MULTI mode
})

// SendDTMF, Hangup, Bridge, Unbridge, Transfer — AppID is auto-included
// from the Call state (set during Dial/Accept), no manual action needed.
err = client.SendDTMF(ctx, result.CallID, "1234#")  // appid auto-included
err = client.Hangup(ctx, result.CallID)               // appid auto-included
```

### Scoped Call Tracking

In MULTI mode, calls are tracked by `appid:channel:uid` on the server. Two calls with the same `channel:uid` but different `appid` values are independent — no cross-routing occurs.

```go
// These are two separate calls — same channel:uid, different appids
client.Dial(ctx, telephony.DialParams{
    Channel: "shared-ch", UID: "100",
    AppID: "appid-A", // ...
})
client.Dial(ctx, telephony.DialParams{
    Channel: "shared-ch", UID: "100",
    AppID: "appid-B", // ...
})
```

### Concurrent Calls

The SDK uses `request_id`-based response matching, so multiple concurrent `Dial()` calls from the same client never collide. Each command gets a unique ID; the server echoes it back.

---

## Phone Number Subscription — Inbound Call Routing

### How It Works

1. Client subscribes to one or more phone numbers (DIDs)
2. When a call arrives on that DID, the gateway asks the CM to look up credentials
3. CM sees the WS subscription → holds the HTTP lookup → broadcasts `call_incoming` to the client
4. Client calls `Accept()` with Agora credentials → CM responds to the gateway with the bundle
5. Gateway bridges the call to the Agora channel
6. Lifecycle events (answered, bridge, dtmf, hangup) flow back via webhook → WS

### Subscribe at Connect

```go
client := telephony.NewClient(wsURL, auth, "client-1", appID)
client.SetSubscribeNumbers([]string{"+18005551234", "+18009876543"})
client.SetHandler(&MyHandler{claimCalls: true})
client.Connect(ctx)
```

### Subscribe on Live Connection

```go
// Add new numbers (replaces the entire subscription list)
err := client.Subscribe(ctx, []string{"+18005551234", "+18001112222"})
```

### Receive and Accept Inbound Calls

```go
func (h *MyHandler) OnCallIncoming(call *telephony.Call) bool {
    log.Printf("Incoming: callid=%s from=%s to=%s", call.CallID, call.From, call.To)

    // Decide whether to accept based on the To number, caller, etc.
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        err := h.client.Accept(ctx, call.CallID, telephony.AcceptParams{
            Token:   generateAgoraToken(call),
            Channel: "inbound-" + call.CallID[:8],
            UID:     "200",
        })
        if err != nil {
            log.Printf("Accept failed: %v", err)
        }
    }()

    return true // must return true to claim the call
}
```

### Reject Inbound Calls

```go
func (h *MyHandler) OnCallIncoming(call *telephony.Call) bool {
    if isBlocked(call.From) {
        go h.client.Reject(context.Background(), call.CallID, "blocked")
        return true
    }
    // ...
}
```

### Lookup Precedence

When a gateway sends a `lookup`, the CM resolves it in this order:

1. **pinMap** (local cache) — return bundle immediately
2. **pinlookup URL** (external webhook for the DID) — forward to webhook
3. **WS subscription** (no pinlookup URL, WS client subscribed to DID) — hold request, broadcast `call_incoming`
4. **Not found** — 500

WS subscription only activates for DIDs without a configured `pinlookup` URL.

---

## Webhook URL Auto-Injection

The WS proxy auto-injects `webhook_url` into commands so the gateway sends lifecycle events back to the CM, which routes them to your WS client.

Auto-injected for:
- `Dial()` (outbound) — A-leg webhook
- `Accept()` (inbound) — B-leg webhook
- `SendDTMF()` — DTMF response webhook

You do **not** need to set `WebhookURL` in `AcceptParams` or configure any webhook endpoint. The CM handles it automatically.

---

## Call Flows

### Outbound Call

```
Client                    CM                                  Gateway
  │                              │                            │
  │── Dial() ──────────────────>│                            │
  │   {action:outbound, ...}    │── POST /pstn ────────────>│
  │                              │   {+webhook_url}          │
  │<── {callid, success} ───────│<── {callid} ──────────────│
  │                              │                            │
  │                              │<── POST /webhook ─────────│  call_answered
  │<── OnCallAnswered ──────────│   {event, callid, ...}    │
  │                              │                            │
  │                              │<── POST /webhook ─────────│  agora_bridge_start
  │<── OnBridgeStart ───────────│                            │
  │                              │                            │
  │── SendDTMF() ──────────────>│── POST /pstn ────────────>│
  │                              │<── POST /webhook ─────────│  dtmf_received
  │<── OnDTMFReceived ──────────│                            │
  │                              │                            │
  │── Hangup() ────────────────>│── POST /pstn (endcall) ──>│
  │                              │                            │
```

### Inbound Call

```
Client                    CM                                  Gateway
  │                              │                            │
  │── Subscribe(DID) ──────────>│                            │
  │                              │                            │
  │                              │<── POST /service (lookup)─│  Inbound call on DID
  │                              │   {did, callerid, callid} │
  │                              │                            │
  │                              │── hold HTTP response ──── │  (WS subscription match)
  │<── OnCallIncoming ──────────│   broadcast call_incoming  │
  │   {callid, from, to}       │                            │
  │                              │                            │
  │── Accept(callid) ──────────>│── respond to held lookup ─>│  {token, channel, uid,
  │                              │   (+webhook_url)          │   +webhook_url}
  │                              │                            │
  │                              │<── POST /webhook ─────────│  call_answered
  │<── OnCallAnswered ──────────│                            │
  │                              │<── POST /webhook ─────────│  agora_bridge_start
  │<── OnBridgeStart ───────────│                            │
  │                              │                            │
  │── SendDTMF(callid) ────────>│── POST gateway ──────────>│  DTMF
  │<── OnDTMFReceived ──────────│<── POST /webhook ─────────│
  │                              │                            │
  │── Hangup(callid) ──────────>│── POST gateway (hangup) ─>│  Hangup
  │<── OnCallHangup ────────────│                            │
```

---

## Concurrency & Thread Safety

- All public methods are goroutine-safe
- WebSocket writes are serialized internally (gorilla/websocket requires single writer)
- Call state map protected by `sync.RWMutex`
- Pending command map uses its own `sync.Mutex`
- Connection state is `atomic.Bool`
- Event handler callbacks run outside locks — safe to call SDK methods from handlers (use goroutines for blocking calls like `Accept`)

---

## Complete Handler Example

```go
type MyHandler struct {
    client *telephony.Client
}

func (h *MyHandler) OnConnected(sessionID string) {
    log.Printf("Connected: session=%s", sessionID)
}

func (h *MyHandler) OnCallIncoming(call *telephony.Call) bool {
    log.Printf("Incoming call: %s from %s to %s", call.CallID, call.From, call.To)

    // Accept the call asynchronously
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        err := h.client.Accept(ctx, call.CallID, telephony.AcceptParams{
            Token:   "agora-rtc-token",
            Channel: "inbound-ch",
            UID:     "200",
        })
        if err != nil {
            log.Printf("Accept failed: %v", err)
        }
    }()
    return true // claim the call
}

func (h *MyHandler) OnCallRinging(call *telephony.Call) {
    log.Printf("Ringing: %s", call.CallID)
}

func (h *MyHandler) OnCallAnswered(call *telephony.Call) {
    log.Printf("Answered: %s", call.CallID)
}

func (h *MyHandler) OnBridgeStart(call *telephony.Call) {
    log.Printf("Bridge started: %s channel=%s", call.CallID, call.Channel)
}

func (h *MyHandler) OnBridgeEnd(call *telephony.Call) {
    log.Printf("Bridge ended: %s", call.CallID)
}

func (h *MyHandler) OnCallHangup(call *telephony.Call) {
    log.Printf("Hangup: %s", call.CallID)
}

func (h *MyHandler) OnError(err error) {
    log.Printf("Error: %v", err)
}

// Implement DTMFHandler to receive DTMF events
func (h *MyHandler) OnDTMFReceived(call *telephony.Call, digits string) {
    log.Printf("DTMF received on %s: %s", call.CallID, digits)
}
```

---

## Examples

Runnable examples are in [`telephony/go/examples/`](telephony/go/examples/). Each is a standalone `main.go` with env var configuration.

```bash
cd telephony/go/examples

export CM_HOST="wss://your-cm-host"
export AUTH_TOKEN="Basic YOUR_TOKEN"
export APPID="your_appid"
```

| Example | Command | What it does |
|---------|---------|-------------|
| **connect** | `go run ./connect/` | Verify credentials — connect, register, print session ID, exit |
| **outbound** | `go run ./outbound/` | Place a call, wait for events (answered, bridge), send DTMF `1234#`, hangup |
| **inbound** | `go run ./inbound/` | Subscribe to a DID, auto-accept incoming calls, log all events until hangup |

### Outbound Example

```bash
export TO_NUMBER="+18005551234"
export FROM_NUMBER="+15551234567"
go run ./outbound/
```

Output (one JSON line per event):
```
Connecting to wss://your-cm-host ...
Dialing +18005551234 from +15551234567 ...
Call placed: callid=3636eaab-7dfe-... channel=example_1707654321000
{"event":"call_answered","callid":"3636eaab-7dfe-...","timestamp":"..."}
{"event":"agora_bridge_start","callid":"3636eaab-7dfe-...","channel":"example_1707654321000","timestamp":"..."}
Call bridged to Agora channel
Sending DTMF: 1234#
{"event":"dtmf_received","callid":"3636eaab-7dfe-...","digits":"1234#","timestamp":"..."}
Hanging up...
{"event":"call_hangup","callid":"3636eaab-7dfe-...","timestamp":"..."}
Done
```

### Inbound Example

```bash
export DID="18005551234"
go run ./inbound/
```

The client subscribes to the DID and waits. When someone calls that number (or you trigger a loopback via the outbound API), the call arrives as `call_incoming` and is auto-accepted.

---

## Troubleshooting

| Error | Cause | Fix |
|-------|-------|-----|
| `ws dial failed: dial tcp: ...` | Cannot reach CM server | Verify `CM_HOST` is correct and accessible from your network |
| `registration failed: unauthorized` | Bad auth token | Verify `AUTH_TOKEN` matches the token provisioned for your App ID |
| `unexpected status: error` | Server rejected connection | Check server logs — may be at capacity or misconfigured |
| `not connected` | Calling SDK methods before `Connect()` | Ensure `Connect()` returns `nil` before calling `Dial`, `Accept`, etc. |
| `command timeout` (30s) | Gateway not responding | Gateway may be down or overloaded — retry later |
| `Dial` returns `Success: false` | No gateways available in region | Check that gateways are running in the requested `Region` |
| `no call_incoming` for inbound | DID has a pinlookup URL configured | WS subscription only works for DIDs without a `pinlookup` webhook. Contact Agora to verify DID config |
| `Accept failed: callid not found` | Call timed out before accept | The gateway has a ~15s lookup timeout. Accept faster, or the call was already rejected |
| Connection drops + `OnError` fires | Network interruption | SDK auto-reconnects with exponential backoff (1s → 30s). Call state is preserved |

### Environment Variables (Examples)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `CM_HOST` | yes | — | CM WebSocket host (e.g. `wss://your-cm-host`) |
| `AUTH_TOKEN` | yes | — | Auth token (e.g. `Basic LkP3sQ8j...`) |
| `APPID` | yes | — | Agora App ID |
| `TO_NUMBER` | outbound | — | Destination phone number (E.164 format) |
| `FROM_NUMBER` | outbound | `+15551234567` | Caller ID phone number |
| `DID` | inbound | — | Phone number to subscribe to for inbound calls |
| `REGION` | outbound | `AREA_CODE_NA` | Gateway region |
