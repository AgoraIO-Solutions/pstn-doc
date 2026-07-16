# Inbound Call Lookup — Example Lambda

A reference implementation of the [Inbound Call Lookup](README.md#calllookup) webhook: an AWS Lambda that answers the Agora PSTN gateway's lookup request with an RTC token and channel, so the gateway knows where to connect the caller.

Use this when you want inbound callers connected to an Agora RTC channel. If you want them connected to a **Conversational AI agent** instead, see [ConvoAI](convoAI.md).

**Code:** [`lambda/token_gen.py`](lambda/token_gen.py)

## What it does

The gateway POSTs the caller's details to your endpoint:

```json
{
  "did": "17177440111",
  "pin": "334455",
  "callerid": "1765740333"
}
```

The Lambda generates an Agora RTC token (v007, 24h expiry) and a random channel name, and responds:

```json
{
  "token": "007eJx...",
  "uid": "101",
  "channel": "A1B2C3D4E5",
  "appid": "your_app_id",
  "audio_scenario": "0"
}
```

The gateway then joins that channel and bridges the caller's audio into it. Your own client app joins the same channel to talk to the caller.

See [Inbound Call Lookup](README.md#calllookup) for the full request/response contract, including the `callid` and `sip_headers` fields the gateway also sends.

## Setup

1. Create a Python AWS Lambda and paste in [`lambda/token_gen.py`](lambda/token_gen.py).
2. Enable a **Lambda Function URL** (or front it with API Gateway) so the gateway can reach it over HTTPS.
3. Set the environment variables below.
4. Give the resulting URL to Agora PSTN admin to assign to your inbound phone number.

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `APP_ID` | yes | Agora App ID |
| `APP_CERTIFICATE` | no | Agora App Certificate. Leave empty if security/tokens are not enabled for the App ID — the App ID is returned as the token. |
| `USER_UID` | no | Agora uid the gateway joins as. Default `101`. |
| `AUDIO_SCENARIO` | no | Audio optimization mode. `0` = human conferencing (default), `10` = optimized for talking to AI agents. |
| `WEBHOOK_URL` | no | Your endpoint for call lifecycle events (see [Webhook Events](README.md#webhooks)). Returned to the gateway in the lookup response. |
| `SDK_OPTIONS` | no | JSON string of Agora SDK options, e.g. `{"rtc.client_type":"71"}`. |

## Notes

- **Channel naming** — this example generates a random 10-character channel per call. In production you'll usually derive the channel from the `did`, `pin`, or `callerid` so the caller lands in the channel your app expects.
- **Rejecting a call** — return HTTP `404` to reject the inbound call.
- **Token expiry** — tokens are generated with a 24-hour lifetime.
