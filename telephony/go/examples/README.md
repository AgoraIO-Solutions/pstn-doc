# Telephony Go SDK — Examples

Three runnable examples that verify the SDK works end-to-end.

## Credentials (dev environment)

```bash
export CM_HOST="wss://sip.dev.cm.01.agora.io"
export AUTH_TOKEN="Basic LkP3sQ8jOvG7fI4mW1uA9eT2rH0yN5oX6zD2kV7p"
export APPID="20b7c51ff4c644ab80cf5a4e646b0537"
export SIP="sip.dev.lb.01.agora.io:5080"
export FROM_NUMBER="+15551234567"
```

Test DID: `18005551234` — a loopback number that routes calls back through the gateway.

## Step 1: Verify connection

```bash
cd connect && go run .
```

Expected: `OK — authenticated and registered successfully`

## Step 2: Outbound call

```bash
cd outbound
export TO_NUMBER="+18005551234"
go run .
```

Expected events: `call_ringing` → `call_answered` → DTMF sent → `call_hangup`

Note: For loopback test calls, `agora_bridge_start` only fires on the B-leg (inbound/accept side), not the outbound A-leg. This is normal. For real PSTN calls to a phone, the A-leg also receives bridge events.

## Step 3: Inbound call (two terminals)

**Terminal 1** — start the listener:
```bash
cd inbound
export DID="18005551234"
go run .
```

**Terminal 2** — trigger the call:
```bash
cd outbound
export TO_NUMBER="+18005551234"
go run .
```

Terminal 1 expected events: `call_incoming` → `call_answered` → `agora_bridge_start` → `call_hangup`

## Transport options

The `SIP` env var controls which transport protocol the load balancer uses:

| Transport | SIP value |
|-----------|-----------|
| UDP (default) | `sip.dev.lb.01.agora.io:5080` |
| TCP | `sip.dev.lb.01.agora.io:5080;transport=tcp` |
| TLS | `sip.dev.lb.01.agora.io:5081;transport=tls` |

If `SIP` is not set, the CM picks a gateway directly (bypasses load balancer).
