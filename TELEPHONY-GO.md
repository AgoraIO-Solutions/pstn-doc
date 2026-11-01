# Telephony Go SDK

Go SDK for the Agora SIP Call Manager WebSocket API. Place, receive, and manage SIP calls via a WebSocket connection.

**Full documentation, source code, examples, and tests:**
[`github.com/AgoraIO-Solutions/telephony-go`](https://github.com/AgoraIO-Solutions/telephony-go)

## Quick Install

```bash
go get github.com/AgoraIO-Solutions/telephony-go
```

## What's in the SDK repo

| Directory | Contents |
|-----------|----------|
| [`client.go`](https://github.com/AgoraIO-Solutions/telephony-go/blob/master/client.go) | SDK source — single-file WebSocket client |
| [`examples/`](https://github.com/AgoraIO-Solutions/telephony-go/tree/master/examples) | Runnable examples: connect, outbound call, inbound call |
| [`test/`](https://github.com/AgoraIO-Solutions/telephony-go/tree/master/test) | E2E tests against live CM and gateway infrastructure |
| [`README.md`](https://github.com/AgoraIO-Solutions/telephony-go/blob/master/README.md) | Full API reference, call flows, MULTI mode, troubleshooting |

## Features

- Outbound calls (Dial, Hangup)
- Inbound call handling (Subscribe, Accept, Reject). The caller hears ringing until a subscribed client accepts; the call is answered only on Accept. Reject, or no client response within 30 seconds, returns SIP `403 Forbidden` to the caller without answering
- DTMF send and receive
- Call bridge/unbridge (connect/disconnect Agora RTC channels)
- Call transfer
- MULTI-AppID mode for multi-tenant platforms
- Automatic reconnection with exponential backoff
- Thread-safe concurrent call tracking
