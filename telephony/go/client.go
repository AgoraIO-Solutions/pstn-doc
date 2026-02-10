package telephony

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// Call represents the state of a SIP call.
type Call struct {
	CallID    string `json:"callid"`
	State     string `json:"state"` // incoming, ringing, answered, bridged, unbridged, hangup
	Direction string `json:"direction"`
	From      string `json:"from"`
	To        string `json:"to"`
	Channel   string `json:"channel"`
	UID       string `json:"uid"`
	AppID     string `json:"appid,omitempty"`
}

// EventHandler is the interface that consumers implement to receive call events.
type EventHandler interface {
	OnConnected(sessionID string)
	OnCallIncoming(call *Call) bool // return true to claim
	OnCallRinging(call *Call)
	OnCallAnswered(call *Call)
	OnBridgeStart(call *Call)
	OnBridgeEnd(call *Call)
	OnCallHangup(call *Call)
	OnError(err error)
}

// DTMFHandler is an optional interface for receiving DTMF events.
// Implement this on your EventHandler to receive OnDTMFReceived callbacks.
type DTMFHandler interface {
	OnDTMFReceived(call *Call, digits string)
}

// DialParams contains parameters for placing an outbound call.
type DialParams struct {
	To        string `json:"to"`
	From      string `json:"from"`
	Channel   string `json:"channel"`
	UID       string `json:"uid"`
	Token     string `json:"token"`
	Region    string `json:"region"`
	Timeout   string `json:"timeout"`
	Sip       string `json:"sip,omitempty"`
	SipDomain string `json:"sip_domain,omitempty"`
	AppID     string `json:"appid,omitempty"`
}

// DialResult contains the response from a Dial request.
type DialResult struct {
	Success bool   `json:"success"`
	CallID  string `json:"callid"`
	Data    map[string]interface{}
}

// AcceptParams contains parameters for accepting an inbound call.
type AcceptParams struct {
	Token         string `json:"token"`
	Channel       string `json:"channel"`
	UID           string `json:"uid"`
	AppID         string `json:"appid,omitempty"`
	WebhookURL    string `json:"webhook_url,omitempty"`
	SDKOptions    string `json:"sdk_options,omitempty"`
	AudioScenario string `json:"audio_scenario,omitempty"`
}

// BridgeParams contains parameters for bridging a call to Agora.
type BridgeParams struct {
	Token         string `json:"token"`
	Channel       string `json:"channel"`
	UID           string `json:"uid"`
	AppID         string `json:"appid,omitempty"`
	SDKOptions    string `json:"sdk_options,omitempty"`
	AudioScenario string `json:"audio_scenario,omitempty"`
}

// Client is the Telephony WebSocket SDK client.
type Client struct {
	wsURL            string
	authToken        string
	clientID         string
	appID            string
	subscribeNumbers []string
	conn             *websocket.Conn
	calls            map[string]*Call // callid -> call state
	mu               sync.RWMutex    // protects calls, conn, handler
	handler          EventHandler
	connected        atomic.Bool
	done             chan struct{}

	// pending responses keyed by request_id
	pendingMu sync.Mutex
	pending   map[string]chan map[string]interface{}
	nextReqID atomic.Uint64

	// writeMu serializes WS writes — gorilla/websocket doesn't support concurrent writers
	writeMu sync.Mutex
}

// NewClient creates a new Telephony WebSocket client.
func NewClient(wsURL, authToken, clientID, appID string) *Client {
	return &Client{
		wsURL:     wsURL,
		authToken: authToken,
		clientID:  clientID,
		appID:     appID,
		calls:     make(map[string]*Call),
		done:      make(chan struct{}),
		pending:   make(map[string]chan map[string]interface{}),
	}
}

// SetHandler sets the event handler for receiving call events.
func (c *Client) SetHandler(h EventHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handler = h
}

// getHandler returns the current event handler under read lock.
func (c *Client) getHandler() EventHandler {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.handler
}

// getConn returns the current websocket connection under read lock.
func (c *Client) getConn() *websocket.Conn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

// SetSubscribeNumbers sets the phone numbers to subscribe to for inbound call filtering.
// Numbers are sent to the server during Connect(). Call before Connect().
func (c *Client) SetSubscribeNumbers(numbers []string) {
	c.subscribeNumbers = numbers
}

// Connect dials the WebSocket server and sends a register message.
func (c *Client) Connect(ctx context.Context) error {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.DialContext(ctx, c.wsURL, nil)
	if err != nil {
		return fmt.Errorf("ws dial failed: %w", err)
	}

	// Read connected message
	_, msg, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to read connected message: %w", err)
	}

	var connMsg map[string]interface{}
	if err := json.Unmarshal(msg, &connMsg); err != nil {
		conn.Close()
		return fmt.Errorf("invalid connected message: %w", err)
	}

	if connMsg["status"] != "connected" {
		conn.Close()
		return fmt.Errorf("unexpected status: %v", connMsg["status"])
	}

	sessionID, _ := connMsg["session_id"].(string)

	// Send register
	regMsg := map[string]interface{}{
		"action":     "register",
		"auth_token": c.authToken,
		"client_id":  c.clientID,
		"appid":      c.appID,
	}
	if len(c.subscribeNumbers) > 0 {
		regMsg["subscribe_numbers"] = c.subscribeNumbers
	}
	if err := conn.WriteJSON(regMsg); err != nil {
		conn.Close()
		return fmt.Errorf("register send failed: %w", err)
	}

	// Read register response
	_, msg, err = conn.ReadMessage()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to read register response: %w", err)
	}

	var regResp map[string]interface{}
	if err := json.Unmarshal(msg, &regResp); err != nil {
		conn.Close()
		return fmt.Errorf("invalid register response: %w", err)
	}

	if regResp["status"] != "registered" {
		conn.Close()
		return fmt.Errorf("registration failed: %v", regResp["error"])
	}

	// Store conn under lock
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	c.connected.Store(true)

	// Set up ping/pong for keepalive
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		return nil
	})
	conn.SetPingHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		c.writeMu.Lock()
		err := conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(5*time.Second))
		c.writeMu.Unlock()
		return err
	})

	h := c.getHandler()
	if h != nil {
		h.OnConnected(sessionID)
	}

	// Start read loop and ping loop
	go c.readLoop(conn)
	go c.pingLoop(conn)

	return nil
}

// Subscribe updates the phone number subscriptions on a live connection.
func (c *Client) Subscribe(ctx context.Context, numbers []string) error {
	c.subscribeNumbers = numbers
	msg := map[string]interface{}{
		"action":  "subscribe",
		"numbers": numbers,
	}
	resp, err := c.sendCommand(ctx, "subscribe", msg)
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("connection lost")
	}
	if errMsg, ok := resp["error"].(string); ok {
		return errors.New(errMsg)
	}
	return nil
}

// Close gracefully closes the WebSocket connection.
func (c *Client) Close() error {
	c.connected.Store(false)
	select {
	case <-c.done:
	default:
		close(c.done)
	}

	// Drain all pending command channels
	c.pendingMu.Lock()
	for action, ch := range c.pending {
		select {
		case ch <- nil:
		default:
		}
		delete(c.pending, action)
	}
	c.pendingMu.Unlock()

	c.mu.Lock()
	conn := c.conn
	c.conn = nil
	c.mu.Unlock()

	if conn != nil {
		return conn.Close()
	}
	return nil
}

// Dial places an outbound call.
func (c *Client) Dial(ctx context.Context, params DialParams) (*DialResult, error) {
	msg := map[string]interface{}{
		"action":  "outbound",
		"to":      params.To,
		"from":    params.From,
		"channel": params.Channel,
		"uid":     params.UID,
		"token":   params.Token,
		"region":  params.Region,
		"timeout": params.Timeout,
	}
	if params.Sip != "" {
		msg["sip"] = params.Sip
	}
	if params.SipDomain != "" {
		msg["sip_domain"] = params.SipDomain
	}
	if params.AppID != "" {
		msg["appid"] = params.AppID
	}

	// Pre-track call by channel:uid
	c.mu.Lock()
	chanKey := params.Channel + ":" + params.UID
	c.calls[chanKey] = &Call{
		State:     "ringing",
		Direction: "outbound",
		To:        params.To,
		From:      params.From,
		Channel:   params.Channel,
		UID:       params.UID,
		AppID:     params.AppID,
	}
	c.mu.Unlock()

	resp, err := c.sendCommand(ctx, "outbound", msg)
	if err != nil {
		// Clean up pre-tracked entry on failure
		c.mu.Lock()
		delete(c.calls, chanKey)
		c.mu.Unlock()
		return nil, err
	}
	if resp == nil {
		c.mu.Lock()
		delete(c.calls, chanKey)
		c.mu.Unlock()
		return nil, errors.New("connection lost")
	}

	result := &DialResult{Data: resp}

	// Extract data from response
	data, _ := resp["data"].(map[string]interface{})
	if data != nil {
		result.Success, _ = data["success"].(bool)
		result.CallID, _ = data["callid"].(string)
	}

	c.mu.Lock()
	if result.CallID != "" {
		// Promote pre-tracked entry to real callid key
		call := c.calls[chanKey]
		if call != nil {
			call.CallID = result.CallID
			c.calls[result.CallID] = call
		}
	}
	// Always remove the temporary chanKey entry
	delete(c.calls, chanKey)
	c.mu.Unlock()

	return result, nil
}

// Accept accepts an inbound call with credentials.
func (c *Client) Accept(ctx context.Context, callid string, creds AcceptParams) error {
	msg := map[string]interface{}{
		"action":  "accept",
		"callid":  callid,
		"token":   creds.Token,
		"channel": creds.Channel,
		"uid":     creds.UID,
	}
	if creds.AppID != "" {
		msg["appid"] = creds.AppID
	}
	if creds.WebhookURL != "" {
		msg["webhook_url"] = creds.WebhookURL
	}
	if creds.SDKOptions != "" {
		msg["sdk_options"] = creds.SDKOptions
	}
	if creds.AudioScenario != "" {
		msg["audio_scenario"] = creds.AudioScenario
	}

	resp, err := c.sendCommand(ctx, "accept", msg)
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("connection lost")
	}
	if errMsg, ok := resp["error"].(string); ok {
		return errors.New(errMsg)
	}

	// Update call's AppID so subsequent commands (send_dtmf, hangup) include it
	if creds.AppID != "" {
		c.mu.Lock()
		if call := c.calls[callid]; call != nil {
			call.AppID = creds.AppID
		}
		c.mu.Unlock()
	}
	return nil
}

// Reject rejects an inbound call.
func (c *Client) Reject(ctx context.Context, callid, reason string) error {
	msg := map[string]interface{}{
		"action": "reject",
		"callid": callid,
		"reason": reason,
	}

	resp, err := c.sendCommand(ctx, "reject", msg)
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("connection lost")
	}
	if errMsg, ok := resp["error"].(string); ok {
		return errors.New(errMsg)
	}
	return nil
}

// Bridge bridges the call to an Agora channel.
func (c *Client) Bridge(ctx context.Context, callid string, creds BridgeParams) error {
	msg := map[string]interface{}{
		"action":  "bridge",
		"callid":  callid,
		"token":   creds.Token,
		"channel": creds.Channel,
		"uid":     creds.UID,
	}
	if creds.AppID != "" {
		msg["appid"] = creds.AppID
	}
	if creds.SDKOptions != "" {
		msg["sdk_options"] = creds.SDKOptions
	}
	if creds.AudioScenario != "" {
		msg["audio_scenario"] = creds.AudioScenario
	}

	resp, err := c.sendCommand(ctx, "bridge", msg)
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("connection lost")
	}
	if errMsg, ok := resp["error"].(string); ok {
		return errors.New(errMsg)
	}
	return nil
}

// Unbridge removes the Agora channel bridge from the call.
func (c *Client) Unbridge(ctx context.Context, callid string) error {
	msg := map[string]interface{}{
		"action": "unbridge",
		"callid": callid,
	}
	c.mu.RLock()
	if call := c.calls[callid]; call != nil && call.AppID != "" {
		msg["appid"] = call.AppID
	}
	c.mu.RUnlock()
	resp, err := c.sendCommand(ctx, "unbridge", msg)
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("connection lost")
	}
	if errMsg, ok := resp["error"].(string); ok {
		return errors.New(errMsg)
	}
	return nil
}

// Hangup ends a call. Sends endcall for outbound, hangup for inbound.
func (c *Client) Hangup(ctx context.Context, callid string) error {
	c.mu.RLock()
	call := c.calls[callid]
	c.mu.RUnlock()

	action := "hangup"
	if call != nil && call.Direction == "outbound" {
		action = "endcall"
	}

	msg := map[string]interface{}{
		"action": action,
		"callid": callid,
	}
	if call != nil && call.AppID != "" {
		msg["appid"] = call.AppID
	}

	resp, err := c.sendCommand(ctx, action, msg)
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("connection lost")
	}
	if errMsg, ok := resp["error"].(string); ok {
		return errors.New(errMsg)
	}

	c.mu.Lock()
	delete(c.calls, callid)
	c.mu.Unlock()

	return nil
}

// Transfer transfers a call to another destination.
func (c *Client) Transfer(ctx context.Context, callid, destination, leg string) error {
	msg := map[string]interface{}{
		"action":      "transfer",
		"callid":      callid,
		"destination": destination,
	}
	if leg != "" {
		msg["leg"] = leg
	}
	c.mu.RLock()
	if call := c.calls[callid]; call != nil && call.AppID != "" {
		msg["appid"] = call.AppID
	}
	c.mu.RUnlock()

	resp, err := c.sendCommand(ctx, "transfer", msg)
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("connection lost")
	}
	if errMsg, ok := resp["error"].(string); ok {
		return errors.New(errMsg)
	}
	return nil
}

// SendDTMF sends DTMF digits on an active call.
func (c *Client) SendDTMF(ctx context.Context, callid, digits string) error {
	msg := map[string]interface{}{
		"action": "send_dtmf",
		"callid": callid,
		"digits": digits,
	}
	c.mu.RLock()
	if call := c.calls[callid]; call != nil && call.AppID != "" {
		msg["appid"] = call.AppID
	}
	c.mu.RUnlock()
	resp, err := c.sendCommand(ctx, "send_dtmf", msg)
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("connection lost")
	}
	if errMsg, ok := resp["error"].(string); ok {
		return errors.New(errMsg)
	}
	return nil
}

// GetActiveCalls returns all currently tracked calls.
func (c *Client) GetActiveCalls() []*Call {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*Call, 0, len(c.calls))
	for _, call := range c.calls {
		result = append(result, call)
	}
	return result
}

// IsConnected returns whether the client is currently connected.
func (c *Client) IsConnected() bool {
	return c.connected.Load()
}

// --- Internal methods ---

func (c *Client) pingLoop(conn *websocket.Conn) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.writeMu.Lock()
			err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second))
			c.writeMu.Unlock()
			if err != nil {
				return
			}
		}
	}
}

func (c *Client) readLoop(conn *websocket.Conn) {
	defer func() {
		c.connected.Store(false)

		// Drain all pending command channels so sendCommand callers don't block forever
		c.pendingMu.Lock()
		for action, ch := range c.pending {
			select {
			case ch <- nil:
			default:
			}
			delete(c.pending, action)
		}
		c.pendingMu.Unlock()
	}()

	for {
		select {
		case <-c.done:
			return
		default:
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			if c.connected.Load() {
				h := c.getHandler()
				if h != nil {
					h.OnError(fmt.Errorf("read error: %w", err))
				}
				go c.reconnect()
			}
			return
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(msg, &parsed); err != nil {
			continue
		}

		// Match response to pending command by request_id
		if reqID, ok := parsed["request_id"].(string); ok && reqID != "" {
			c.pendingMu.Lock()
			ch, exists := c.pending[reqID]
			if exists {
				delete(c.pending, reqID)
			}
			c.pendingMu.Unlock()

			if exists {
				ch <- parsed
				continue
			}
		}

		// Otherwise treat as event
		c.handleEvent(parsed)
	}
}

func (c *Client) sendCommand(ctx context.Context, action string, msg map[string]interface{}) (map[string]interface{}, error) {
	if !c.connected.Load() {
		return nil, errors.New("not connected")
	}
	conn := c.getConn()
	if conn == nil {
		return nil, errors.New("not connected")
	}

	// Generate unique request_id for response matching
	reqID := fmt.Sprintf("%s_%d", action, c.nextReqID.Add(1))
	msg["request_id"] = reqID

	respCh := make(chan map[string]interface{}, 1)
	c.pendingMu.Lock()
	c.pending[reqID] = respCh
	c.pendingMu.Unlock()

	c.writeMu.Lock()
	err := conn.WriteJSON(msg)
	c.writeMu.Unlock()
	if err != nil {
		c.pendingMu.Lock()
		delete(c.pending, reqID)
		c.pendingMu.Unlock()
		return nil, fmt.Errorf("send failed: %w", err)
	}

	deadline := time.NewTimer(30 * time.Second)
	defer deadline.Stop()

	select {
	case resp := <-respCh:
		if resp == nil {
			return nil, errors.New("connection lost")
		}
		return resp, nil
	case <-ctx.Done():
		c.pendingMu.Lock()
		delete(c.pending, reqID)
		c.pendingMu.Unlock()
		return nil, ctx.Err()
	case <-deadline.C:
		c.pendingMu.Lock()
		delete(c.pending, reqID)
		c.pendingMu.Unlock()
		return nil, errors.New("command timeout")
	}
}

func (c *Client) handleEvent(msg map[string]interface{}) {
	h := c.getHandler()
	if h == nil {
		return
	}

	eventType, _ := msg["event"].(string)
	callid, _ := msg["callid"].(string)
	channel, _ := msg["channel"].(string)
	uid, _ := msg["uid"].(string)

	// Find or create call state, update fields — all under lock
	c.mu.Lock()
	call := c.calls[callid]
	if call == nil && channel != "" && uid != "" {
		chanKey := channel + ":" + uid
		call = c.calls[chanKey]
	}
	if call == nil && callid != "" {
		call = &Call{
			CallID:  callid,
			Channel: channel,
			UID:     uid,
		}
		c.calls[callid] = call
	}

	if call == nil {
		c.mu.Unlock()
		return
	}

	// Update call fields from event while still holding lock
	if channel != "" {
		call.Channel = channel
	}
	if uid != "" {
		call.UID = uid
	}
	if callid != "" {
		call.CallID = callid
	}
	if from, ok := msg["from"].(string); ok && from != "" {
		call.From = from
	}
	if to, ok := msg["to"].(string); ok && to != "" {
		call.To = to
	}
	if dir, ok := msg["direction"].(string); ok && dir != "" {
		call.Direction = dir
	}
	if appid, ok := msg["appid"].(string); ok && appid != "" {
		call.AppID = appid
	}
	c.mu.Unlock()

	// Dispatch event — handler callbacks run outside lock
	switch eventType {
	case "call_incoming":
		c.mu.Lock()
		call.State = "incoming"
		call.Direction = "inbound"
		c.mu.Unlock()
		claimed := h.OnCallIncoming(call)
		if !claimed {
			c.mu.Lock()
			delete(c.calls, callid)
			c.mu.Unlock()
		}
	case "call_ringing":
		c.mu.Lock()
		call.State = "ringing"
		c.mu.Unlock()
		h.OnCallRinging(call)
	case "call_answered":
		c.mu.Lock()
		call.State = "answered"
		c.mu.Unlock()
		h.OnCallAnswered(call)
	case "agora_bridge_start":
		c.mu.Lock()
		call.State = "bridged"
		c.mu.Unlock()
		h.OnBridgeStart(call)
	case "agora_bridge_end":
		c.mu.Lock()
		call.State = "unbridged"
		c.mu.Unlock()
		h.OnBridgeEnd(call)
	case "call_hangup":
		c.mu.Lock()
		call.State = "hangup"
		delete(c.calls, callid)
		if call.Channel != "" && call.UID != "" {
			delete(c.calls, call.Channel+":"+call.UID)
		}
		c.mu.Unlock()
		h.OnCallHangup(call)
	case "dtmf_received":
		digits, _ := msg["digits"].(string)
		if dh, ok := h.(DTMFHandler); ok {
			dh.OnDTMFReceived(call, digits)
		}
	}
}

func (c *Client) reconnect() {
	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-c.done:
			return
		default:
		}

		time.Sleep(backoff)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := c.Connect(ctx)
		cancel()

		if err == nil {
			return
		}

		h := c.getHandler()
		if h != nil {
			h.OnError(fmt.Errorf("reconnect failed: %w", err))
		}

		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}
