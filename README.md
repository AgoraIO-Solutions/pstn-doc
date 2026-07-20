# Agora PSTN & SIP Gateway


1. [Overview](#overview)
2. [Inbound PSTN](#inbound)
3. [Outbound PSTN](#outbound)
4. [Inbound SIP](#inboundsip)
5. [Inbound Call Lookup](#calllookup)
6. [End Call](#endcall)
7. [Cancel Call](#cancelcall)
8. [Bridge](#bridge)
9. [Unbridge](#unbridge)
10. [Re-bridge](#rebridge)
11. [Send DTMF](#senddtmf)
12. [Transfer (SIP REFER)](#transfer)
13. [Webhook Events](#webhooks)
14. [SIP Entrypoints](#sipentry)
15. [Agora Gateway IPs](#gatewayips)
16. [Twilio Configuration](#configtwilio)

**Example code:** [Inbound Call Lookup Lambda](CALL-LOOKUP-LAMBDA.md) · [PSTN to ConvoAI](convoAI.md) · [Telephony Go SDK](TELEPHONY-GO.md)

## Overview <a name="overview"></a>
These REST APIs allow developers to trigger inbound and outbound PSTN and SIP calls which then connect into an Agora channel enabling end-users to participate with their phone for the audio leg of the conference call.

Please contact us to provision your appid with this service. We will provide you with an authorization header to include in all API requests.

### Example API Request
```bash
curl --location 'https://sipcm.agora.io/v1/api/pstn' \
--header 'Authorization: Basic your_auth_token_here' \
--header 'Content-Type: application/json' \
--data '{
    "action": "outbound",
    "appid": "your_appid_here",
    "region": "AREA_CODE_EU",
    "uid": "43455",
    "channel": "agora_channel",
    "from": "441473943851",
    "to": "+447751997737",
    "prompt": "false"
}'
```

## Inbound PSTN <a name="inbound"></a>
In this scenario, the end-user dials a phone number displayed to them and enters the PIN when prompted. With the correct PIN, they are connected into the Agora channel.

- **URL**: `https://sipcm.agora.io/v1/api/pstn`
- **Method**: `POST`

### Request Body Parameters as JSON
```json
{
  "action":"inbound",
  "appid":"your_appid_here",
  "token":"006d2776fef40dc864dc7a438e3b871IACGnkorQsr0iCpBmzNwKEdzKuAv1b1zRcMy0cOradw6mHN/il8AAAAAIgCQqbAFeaVzZgQAAQDw0oBmAgDw0oBmAwDw0oBmBADw0oBm",
  "uid":"123",
  "channel":"agora_channel",
  "region":"AREA_CODE_NA"
}
```
- `appid` (string): the Agora project appid
- `token` (string) [optional]: a generated Agora RTC access token
- `uid` (string) [optional]: a user uid
- `channel` (string): an Agora channel name
- `webhook_url` (string) [optional]: your webhook endpoint to receive call lifecycle events (see [Webhook Events](#webhooks))
- `sdk_options` (string) [optional]: JSON string of Agora SDK options (e.g., `{"rtc.client_type":"71"}`)
- `audio_scenario` (string) [optional]: Audio optimization mode. Values:
  - "0": Automatic Human Conferencing Scenarios (default)
  - "10": Optimized To Talk to AI Agents
- `video` (boolean) [optional]: Enable video for this call. When `true`, the gateway uses video-capable SIP ports (H264+VP8). Default: `false`
- `region` (string): the user's region where they will likely be located and calling from. Values:

      AREA_CODE_NA: North America
      AREA_CODE_EU: Europe
      AREA_CODE_AS: Asia, excluding Mainland China
      AREA_CODE_JP: Japan
      AREA_CODE_IN: India
      AREA_CODE_CN: Mainland China

### Success Response Example
*Status Code*: `200 OK`
*Content Type*: `application/json`
Body:
```json
{
    "did": "17377583411",
    "display": "+1 (737) 758 3411",
    "pin": "780592"
}
```

- `did`: the phone number to dial
- `display`: the phone number to dial in a more friendly display format
- `pin`: the pin to enter when prompted

### Error Code Responses
401  Unauthorized
500  Missing Parameters
503  No resource currently available

### Notes
Direct Inward Dialing (DID) providers such as Twilio allow you to buy a phone number and have the calls forwarded to a SIP address. We will provide you with the SIP address to forward your calls to when we provision you on this service. You will also be able to customise the voice prompts played to your end-users.

## Outbound PSTN <a name="outbound"></a>
In this scenario, the end-user receives a phone call which connects them directly to the channel when they answer.

- **URL**: `https://sipcm.agora.io/v1/api/pstn`
- **Method**: `POST`

### Request Body Parameters as JSON
```json
{
  "action":"outbound",
  "appid":"your_appid_here",
  "token":"006d2776fef40dc864dc7a438e3b871IACGnkorQsr0iCpBmzNwKEdzKuAv1b1zRcMy0cOradw6mHN/il8AAAAAIgCQqbAFeaVzZgQAAQDw0oBmAgDw0oBmAwDw0oBmBADw0oBm",
  "uid":"123",
  "channel":"agora_channel",
  "region":"AREA_CODE_NA",
  "prompt":"true",
  "to":"+447712886400;dtmf=1234#",
  "from":"+1800222333",
  "timeout":"3600",
  "sip":"trunk.provider.com:5061;transport=tls;username=+15551234567;password=your_password_here;srtp=true",
  "sip_domain":"sip.gateway.agora.io",
  "webhook_url":"https://example.com/webhooks/call-events",
  "sdk_options":"{\"rtc.client_type\":\"71\"}",
  "audio_scenario":"0"
}
```
- `appid` (string): the Agora project appid
- `token` (string) [optional]: a generated access token
- `uid` (string) [optional]: a user uid
- `channel` (string): an Agora channel name
- `prompt` (string): controls the connection behavior when the callee answers. Values:
  - "false": Silent direct connect - No prompt, no beep, connects immediately to Agora
  - "beep": Beep then connect - Plays beep sound, then connects to Agora (no prompt)
  - "lazy": Lenient prompt - Plays voice prompt, accepts any DTMF digit to proceed (with beep)
  - "true": Strict PIN mode - Plays voice prompt, requires pressing "1" to proceed (with beep). Any other digit plays goodbye message and hangs up
- `to` (string): the end-user's phone number to dial. Supports two methods for sending DTMF tones after the call connects:
  - **Method 1** (`;dtmf=`): Append `;dtmf=DIGITS` to the number. The digits are sent as DTMF after the call is answered. Example: `"+15551234567;dtmf=696969#"` dials the number, then sends `696969#` as DTMF. Use this to automate PIN entry or navigate IVR menus.
  - **Method 2** (`#`): Append `#` followed by digits. Example: `"+447712886400#333"` dials the number, then sends `333` as DTMF.
- `from` (string): the calling number displayed on the end-user's phone during ringing
- `sip` (string) [optional]: SIP trunk URI for call termination, or leave blank to use Agora's routing. Supports optional parameters appended with `;`:
  - `transport=tls` — use TLS signaling (recommended)
  - `username=+15551234567` — SIP authentication username
  - `password=your_password_here` — SIP authentication password
  - `srtp=true` — enable SRTP media encryption

  Example: `"trunk.provider.com:5061;transport=tls;username=+15551234567;password=your_password_here;srtp=true"`
- `sip_domain` (string) [optional]: the domain the gateway uses in the SIP `From` header, which is required by some providers (e.g., WhatsApp/Meta). This does **not** affect gateway selection — the call always routes through an available gateway in the requested `region`.
- `timeout` (string) [optional]: max duration for outbound call in seconds. Default 3600 seconds which is 1 hour
- `webhook_url` (string) [optional]: your webhook endpoint to receive call lifecycle events (see [Webhook Events](#webhooks))
- `sdk_options` (string) [optional]: JSON string of Agora SDK options (e.g., `{"rtc.client_type":"71"}`)
- `audio_scenario` (string) [optional]: Audio optimization mode. Values:
  - "0": Automatic Human Conferencing Scenarios (default)
  - "10": Optimized To Talk to AI Agents
- `video` (boolean) [optional]: Enable video for this call. When `true`, the gateway uses video-capable SIP ports (H264+VP8). Default: `false`
- `region` (string): the user's region where they will likely be located and calling from. Values:

      AREA_CODE_NA: North America
      AREA_CODE_EU: Europe
      AREA_CODE_AS: Asia, excluding Mainland China
      AREA_CODE_JP: Japan
      AREA_CODE_IN: India
      AREA_CODE_CN: Mainland China

### Success Response Example
*Status Code*: `200 OK`
*Content Type*: `application/json`
Body:
```json
{
    "success": true,
    "callid": "88877-55Asdd7-55Asdd"
}
```

### Failure Response Example
*Status Code*: `200 OK`
*Content Type*: `application/json`
Body:
```json
{
    "success": false,
    "reason": "Busy"
}
```

- `success`: If 'true' the call has connected and 'callid' will be included in the response. If 'false' the call has failed and 'reason' will be included in the response
- `callid` [success only]: This can be used to end the call
- `reason` [failure only]: The call failed because the user was 'Busy' or the destination was 'Invalid'

### Error Code Responses
401  Unauthorized
500  Missing Parameters
503  No resource currently available

### Notes
If this API returns success 'true' the call has been connected. If it returns success 'false' there will be a reason explaining the failure.

## Inbound SIP <a name="inboundsip"></a>
In this scenario, an inbound SIP address is requested. When the SIP address is dialled, the call will be routed to the requested user/channel session.

- **URL**: `https://sipcm.agora.io/v1/api/pstn`
- **Method**: `POST`

### Request Body Parameters as JSON
```json
{
  "action":"inboundsip",
  "appid":"your_appid_here",
  "token":"006d2776fef40dc864dc7a438e3b871IACGnkorQsr0iCpBmzNwKEdzKuAv1b1zRcMy0cOradw6mHN/il8AAAAAIgCQqbAFeaVzZgQAAQDw0oBmAgDw0oBmAwDw0oBmBADw0oBm",
  "uid":"123",
  "channel":"agora_channel",
  "region":"AREA_CODE_NA"
}
```
- `appid` (string): the Agora project appid
- `token` (string) [optional]: a generated access token
- `uid` (string) [optional]: a user uid
- `channel` (string): an Agora channel name
- `webhook_url` (string) [optional]: your webhook endpoint to receive call lifecycle events (see [Webhook Events](#webhooks))
- `sdk_options` (string) [optional]: JSON string of Agora SDK options (e.g., `{"rtc.client_type":"71"}`)
- `audio_scenario` (string) [optional]: Audio optimization mode. Values:
  - "0": Automatic Human Conferencing Scenarios (default)
  - "10": Optimized To Talk to AI Agents
- `video` (boolean) [optional]: Enable video for this call. When `true`, the gateway uses video-capable SIP ports (H264+VP8). Default: `false`
- `region` (string): the user's region where they will likely be located and calling from. Values:

      AREA_CODE_NA: North America
      AREA_CODE_EU: Europe
      AREA_CODE_AS: Asia, excluding Mainland China
      AREA_CODE_JP: Japan
      AREA_CODE_IN: India
      AREA_CODE_CN: Mainland China

### Success Response Example
*Status Code*: `200 OK`
*Content Type*: `application/json`
Body:
```json
{
    "sip": "sip:pstn_783410988998@2.4.6.7:5080"
}
```

- `sip`: the sip address to dial to join the call

### Error Code Responses
401  Unauthorized
500  Missing Parameters
503  No resource currently available

### Notes
Using this API you can bridge an outbound call from your provider with an inbound sip address into Agora.

## Inbound Call Lookup <a name="calllookup"></a>

When provisioned, the gateway calls your REST endpoint when an inbound call arrives, providing the DID dialed, PIN entered (if configured), and caller ID. Your endpoint returns the Agora channel details to connect the caller to.

**This is a webhook that YOU implement, not an Agora API endpoint.**

- **URL**: `https://example.customer.com/api/pinlookup` *(your endpoint)*
- **Method**: `POST`
- **Authorization**: optional — see [Securing your endpoint](#securing-your-endpoint)

### Securing your endpoint

Your lookup endpoint is publicly reachable, so you should authenticate the requests it receives.

Provide an authorization value when your number is provisioned, and Agora will send it as the `Authorization` header on every lookup request:

```
Authorization: Bearer <your-secret>
```

The value is sent exactly as supplied, so include the scheme (`Bearer `, `Basic `, …) if your endpoint expects one. Validate it on every request and reject anything unauthenticated — returning `404` rejects the inbound call.

**Example implementations** — ready-to-deploy AWS Lambdas that implement this webhook:

| Example | Use when |
|---------|----------|
| [Inbound Call Lookup Lambda](CALL-LOOKUP-LAMBDA.md) — [`lambda/token_gen.py`](lambda/token_gen.py) | You want the caller connected to a regular Agora RTC channel |
| [PSTN to ConvoAI](convoAI.md) — [`lambda/convoai_pstn.py`](lambda/convoai_pstn.py) | You want the caller connected to a Conversational AI agent, configured per PIN |

### Request Body Parameters as JSON (sent by Agora to your endpoint)
```json
{
  "did":"17177440111",
  "pin":"334455",
  "callerid":"1765740333",
  "callid":"f577605c-eb3a-4efe-af1b-ee66d5297569",
  "sip_headers":{ "X-Session-Id":"abc", "X-Trace-Id":"xyz" }
}
```
- `did`: the phone number dialed
- `pin`: the pin entered by caller (empty string if DID not configured for PIN prompts)
- `callerid`: the phone number of caller
- `callid` (string): the gateway's unique identifier for this call. Use it
  immediately to [End Call](#endcall) or [Re-bridge](#rebridge) — it's the
  same `callid` you will see on lifecycle webhook events. Synchronously
  available at this lookup exchange, before any lifecycle event fires.
- `sip_headers` (object): custom `X-*` SIP headers captured on the inbound
  SIP `INVITE`. Always present (empty object `{}` when no custom headers were
  on the wire). String-valued — one value per header name (the gateway never
  sees true on-wire repeats; if the same header name appeared twice, the
  value is the comma-joined string). Inbound SIP only — this object is
  always empty for inbound PSTN (DID) calls and for inbound SIP `direct_`
  numbers (which bypass this webhook).

  Header values come from arbitrary SIP peers and should be treated as
  **untrusted user input** (SQL/HTML/log-injection considerations on your
  side). Size-bounded: ≤20 names, ≤1024 bytes per value, ≤4 KB serialized
  total. Header names prefixed `X-Agora-` are reserved for Agora internal use
  and are never surfaced.

### Success Response Example (from your endpoint)
*Status Code*: `200 OK`
*Content Type*: `application/json`
Body:
```json
{
  "token":"006d2776fef40dc864dc7a438e3b871IACGnkorQsr0iCpBmzNwKEdzKuAv1b1zRcMy0cOradw6mHN/il8AAAAAIgCQqbAFeaVzZgQAAQDw0oBmAgDw0oBmAwDw0oBmBADw0oBm",
  "uid":"123",
  "channel":"agora_channel",
  "appid":"your_appid_here",
  "webhook_url":"https://example.com/webhooks/call-events",
  "sdk_options":"{\"rtc.client_type\":\"71\"}",
  "audio_scenario":"0"
}
```

- `token` (string): a generated access token OR the appid if tokens are not enabled
- `uid` (string): a user uid
- `channel` (string): an Agora channel name
- `appid` (string) [optional]: the Agora appid for your project
- `webhook_url` (string) [optional]: your webhook endpoint to receive call lifecycle events (see [Webhook Events](#webhooks))
- `sdk_options` (string) [optional]: JSON string of Agora SDK options (e.g., `{"rtc.client_type":"71"}`)
- `audio_scenario` (string) [optional]: Audio optimization mode. Values:
  - "0": Automatic Human Conferencing Scenarios (default)
  - "10": Optimized To Talk to AI Agents
- `video` (boolean) [optional]: Enable video for this call. When `true`, the gateway uses video-capable SIP ports (H264+VP8). Default: `false`

### Error Code Responses
404  Not Found

### Notes
This webhook allows dynamic call routing to Agora channels. Your DID can be configured to prompt for PIN entry or route directly based on DID/caller ID. Return 404 to reject the call.

## End Call <a name="endcall"></a>
Terminate a connected call (inbound or outbound) by `callid` + `appid`.

- **URL**: `https://sipcm.agora.io/v1/api/pstn`
- **Method**: `POST`

### Request Body Parameters as JSON
```json
{
  "action":"endcall",
  "appid":"your_appid_here",
  "callid":"f577605c-eb3a-4efe-af1b-ee66d5297569"
}
```
- `appid` (string): the Agora project appid
- `callid` (string): the call id of the ongoing call

### Success Response Example
*Status Code*: `200 OK`
*Content Type*: `application/json`
Body:
```json
{
  "success":true
}
```

### Error Code Responses
404  Not Found

### Notes

Works for both **outbound and inbound** calls.

**How to obtain `callid`:**
- **Outbound:** returned in the [Outbound PSTN](#outbound) success response,
  and on every webhook lifecycle event.
- **Inbound:** delivered in the [Inbound Call Lookup](#calllookup) webhook
  body, *and* on lifecycle events. You can call `endcall` immediately from
  your pinlookup endpoint with the `callid` you just received — no need to
  wait for `call_initiated`.

## Cancel Call <a name="cancelcall"></a>
Cancel a call setup request created by Inbound PSTN, Inbound SIP API or Static PIN webhook request. If the call is already in progress, it will be stopped. You can use one of three methods to cancel a call:

- **URL**: `https://sipcm.agora.io/v1/api/pstn`
- **Method**: `POST`

### Method 1: Cancel by Bundle
Use the same parameters you used when creating the inbound or inboundsip request:

```json
{
  "action":"cancelcall",
  "appid":"your_appid_here",
  "token":"006d2776fef40dc864dc7a438e3b871IACGnkorQsr0iCpBmzNwKEdzKuAv1b1zRcMy0cOradw6mHN/il8AAAAAIgCQqbAFeaVzZgQAAQDw0oBmAgDw0oBmAwDw0oBmBADw0oBm",
  "uid":"123",
  "channel":"agora_channel"
}
```
- `appid` (string): the Agora project appid
- `token` (string): the token used to join the channel
- `uid` (string) [optional]: the uid used to join the channel
- `channel` (string): the channel used for the call

### Method 2: Cancel by SIP Address
Use the sip address returned by the inboundsip API:

```json
{
  "action":"cancelcall",
  "appid":"your_appid_here",
  "sip":"sip:pstn_783410988998@2.4.6.7:5080"
}
```
- `appid` (string): the Agora project appid
- `sip` (string): the sip address returned by inboundsip api

### Method 3: Cancel by DID/PIN
Use the did and pin returned by the inbound API:

```json
{
  "action":"cancelcall",
  "appid":"your_appid_here",
  "did":"17377583411",
  "pin":"780592"
}
```
- `appid` (string): the Agora project appid
- `did` (string): the did returned by inbound api
- `pin` (string): the pin returned by inbound api

### Success Response Example
*Status Code*: `200 OK`
*Content Type*: `application/json`
Body:
```json
{
  "success":true
}
```

### Error Code Responses
404  Not Found

### Notes
This API allows you to cancel a previous call setup request using any of the three methods above. You can cancel calls both before and after they connect - if the call is already in progress, it will be terminated.

## Bridge <a name="bridge"></a>

Attach a connected call's audio to an Agora channel. Used to (re-)establish
the bridge after [Unbridge](#unbridge), or to bridge a call that was placed
without an Agora channel attached. For switching a *currently-bridged*
call to a different channel, use [Re-bridge](#rebridge) instead.

- **URL**: `https://sipcm.agora.io/v1/api/pstn`
- **Method**: `POST`

### Request Body Parameters as JSON
```json
{
  "action":"bridge",
  "appid":"your_appid_here",
  "callid":"f577605c-eb3a-4efe-af1b-ee66d5297569",
  "channel":"agora_channel",
  "token":"<RTC token>",
  "uid":"123"
}
```
- `appid` (string): the Agora project appid
- `callid` (string): the call id of the connected call
- `channel` (string): destination Agora channel
- `token` (string): RTC token for the destination channel (use the appid if
  tokens are not enabled for your project)
- `uid` (string): Agora uid the bridge should join as
- `sdk_options` (string) [optional]: JSON string of Agora SDK options
- `audio_scenario` (string) [optional]: audio optimization mode (see
  [Inbound Call Lookup](#calllookup))
- `webhook_url` (string) [optional]: your webhook endpoint for lifecycle events

### Success Response Example
```json
{ "success":true }
```

### Error Code Responses
404  Not Found (callid not active)

### Notes
Works for both inbound and outbound calls. The SIP/PSTN leg stays up
throughout. Emits `agora_bridge_start` on success (and `agora_bridge_end`
+ `agora_bridge_start` on a subsequent bridge after an unbridge).

## Unbridge <a name="unbridge"></a>

Remove the Agora channel bridge from a connected call. The SIP/PSTN leg
stays up; only the Agora connection is closed. Use [Bridge](#bridge) to
reattach.

- **URL**: `https://sipcm.agora.io/v1/api/pstn`
- **Method**: `POST`

### Request Body Parameters as JSON
```json
{
  "action":"unbridge",
  "appid":"your_appid_here",
  "callid":"f577605c-eb3a-4efe-af1b-ee66d5297569"
}
```
- `appid` (string): the Agora project appid
- `callid` (string): the call id of the connected call

### Success Response Example
```json
{ "success":true }
```

### Error Code Responses
404  Not Found (callid not active)

### Notes
Emits `agora_bridge_end` on success. The SIP leg remains active until you
call [End Call](#endcall) or the remote party hangs up.

## Re-bridge <a name="rebridge"></a>

Move a connected call's Agora bridge to a *different* Agora channel
**without dropping the SIP/PSTN leg**. Works identically for audio-only
and audio+video calls (no audio gap on the SIP side). This is NOT the same
as [Transfer](#transfer), which is a SIP REFER to an external phone number.

- **URL**: `https://sipcm.agora.io/v1/api/pstn`
- **Method**: `POST`

### Request Body Parameters as JSON
```json
{
  "action":"rebridge",
  "appid":"your_appid_here",
  "callid":"f577605c-eb3a-4efe-af1b-ee66d5297569",
  "channel":"agora_new_channel",
  "token":"<RTC token for the new channel>",
  "uid":"123",
  "current_channel":"agora_old_channel",
  "current_uid":"123"
}
```
- `appid` (string): the Agora project appid
- `callid` (string): the call id of the connected call (inbound or outbound)
- `channel` (string): destination Agora channel
- `token` (string): RTC token for the destination channel. **Always
  required** — there is no fallback to the appid here.
- `uid` (string): Agora uid in the destination channel (may differ from
  `current_uid`)
- `current_channel` (string): the channel the call is currently bridged to.
  The gateway validates this against live state and rejects mismatches.
- `current_uid` (string): the uid the bridge is using now.
- `sdk_options` (string) [optional]: JSON string of Agora SDK options
- `audio_scenario` (string) [optional]: audio optimization mode
- `webhook_url` (string) [optional]: your webhook endpoint for lifecycle events

### Success Response Example
```json
{ "success":true }
```

### Failure Response Examples
```json
{ "success":false, "reason":"current channel/uid mismatch" }
{ "success":false, "reason":"call not bridged; use bridge" }
{ "success":false, "reason":"rebridge already in progress" }
{ "success":false, "reason":"join failed: <detail>" }
```
On `join failed`, the gateway emits an
[`agora_bridge_failed`](#agora_bridge_failed) lifecycle event and the SIP
leg stays up. **Recovery**: the failure is break-before-make — the call
now has no active Agora bridge, so retrying `/rebridge` will fail with
`call not bridged; use bridge`. Call [Bridge](#bridge) to establish a
fresh Agora session on the same `callid`, or [End Call](#endcall) to
hang up.

### Error Code Responses
404  Not Found (callid not active)
500  Missing required parameters

### Notes
- Works for **inbound and outbound** calls in all directions.
- The SIP/PSTN leg is never torn down. `call_hangup` is **NOT** emitted on
  rebridge success or destination-join failure.
- Webhook event sequence on success:
  1. `agora_bridge_end { rebridge:true, segment_billsec, ... }` — closes
     the segment-1 CDR for the old channel.
  2. `agora_bridge_start { rebridge:true, previous_channel, previous_uid, ... }`
     — opens the segment-2 CDR for the new channel.
  - Both carry the **same `callid`**. Per-segment billing is derived from
    these labelled events.
- Concurrency: the gateway serializes switches per call and rejects an
  overlapping rebridge with `rebridge already in progress`.

## Send DTMF <a name="senddtmf"></a>

Send DTMF tones on the PSTN/SIP side of a connected call.

- **URL**: `https://sipcm.agora.io/v1/api/pstn`
- **Method**: `POST`

### Request Body Parameters as JSON
```json
{
  "action":"send_dtmf",
  "appid":"your_appid_here",
  "callid":"f577605c-eb3a-4efe-af1b-ee66d5297569",
  "digits":"1234#"
}
```
- `appid` (string): the Agora project appid
- `callid` (string): the call id of the connected call
- `digits` (string): the DTMF digit string to send (`0`-`9`, `*`, `#`)

### Success Response Example
```json
{ "success":true }
```

### Notes
Each digit fires a `dtmf_received` webhook event back on your `webhook_url`
(echoed from the SIP side).

## Transfer (SIP REFER) <a name="transfer"></a>

Hand off a connected call's PSTN/SIP leg to a different external phone
number (or SIP URI) via a SIP REFER. The Agora bridge ends; the remote
PSTN/SIP party continues talking to the new destination. **This is NOT
related to moving an Agora channel** — for that, use [Re-bridge](#rebridge).

- **URL**: `https://sipcm.agora.io/v1/api/pstn`
- **Method**: `POST`

### Request Body Parameters as JSON
```json
{
  "action":"transfer",
  "appid":"your_appid_here",
  "callid":"f577605c-eb3a-4efe-af1b-ee66d5297569",
  "destination":"+18005551234",
  "leg":"aleg"
}
```
- `appid` (string): the Agora project appid
- `callid` (string): the call id of the connected call
- `destination` (string): phone number or SIP URI to refer the call to
- `leg` (string) [optional]: `aleg` (default) or `bleg`

### Success Response Example
```json
{ "success":true }
```

### Notes
SIP REFER outcome is up to the upstream provider — success here means the
REFER was accepted by the gateway, not that the new leg was answered.

## Webhook Events <a name="webhooks"></a>

When you provide a `webhook_url` parameter (in outbound calls or PIN lookup responses), the SIP gateway will send HTTP POST requests to your endpoint with call lifecycle events.

### Events

These webhook events fire for any call regardless of how it was initiated (HTTP API or SDK over WebSocket).

| Event | When fired | Direction |
|-------|-----------|-----------|
| [`call_initiated`](#1-call_initiated) | API received the call request (outbound) or PIN lookup succeeded (inbound) | outbound, inbound |
| [`call_ringing`](#2-call_ringing) | Remote party is ringing | outbound |
| [`call_answered`](#3-call_answered) | Remote party answered (SIP-out adds `wav_file`) | outbound, outbound_sip |
| [`agora_bridge_start`](#4-agora_bridge_start) | Audio bridge to Agora RTC established (on re-bridge: includes `rebridge:true`, `previous_channel`, `previous_uid`) | outbound, inbound |
| [`dtmf_received`](#dtmf_received) | DTMF digit detected on the PSTN leg, or echoed after `/send_dtmf` | outbound, inbound |
| [`agora_bridge_end`](#5-agora_bridge_end) | Agora RTC session ended (on re-bridge: includes `rebridge:true`, `segment_billsec`) | outbound, inbound |
| [`agora_bridge_failed`](#agora_bridge_failed) | Re-bridge destination-join failed; SIP leg stays up; `call_hangup` is NOT emitted | outbound, inbound |
| [`call_hangup`](#6-call_hangup) | Call ended (**guaranteed to fire**; does NOT fire on re-bridge) | outbound, inbound, outbound_sip |

**Outbound flow**: `call_initiated` → `call_ringing` → `call_answered` → `agora_bridge_start` → `agora_bridge_end` → `call_hangup`

**Inbound flow**: `call_initiated` → `agora_bridge_start` → `agora_bridge_end` → `call_hangup`

**SIP-out flow** (SIP-only outbound, `sip_outbound.lua`): `call_answered` → `call_hangup` only.

**Re-bridge** (mid-call channel switch via [Re-bridge](#rebridge)): inserts
a labelled segment-boundary pair into either flow:
`agora_bridge_end{rebridge:true, segment_billsec}` (old channel) →
`agora_bridge_start{rebridge:true, previous_channel, previous_uid}` (new
channel). Same `callid` throughout. `call_hangup` is **not** emitted on
re-bridge. On destination-join failure, `agora_bridge_failed` fires and
the SIP leg stays up.

`dtmf_received` is asynchronous — it can fire any time between `call_answered` / `agora_bridge_start` and `call_hangup`.

### Common Event Fields

All webhook events include these base fields:

| Field | Type | Description |
|-------|------|-------------|
| `event` | string | Event type identifier (e.g., "call_initiated", "call_hangup") |
| `callid` | string | FreeSWITCH UUID for this call session |
| `timestamp` | integer | Unix timestamp when event was generated |
| `uid` | string | Agora user ID (may be empty for early events) |
| `channel` | string | Agora channel ID (may be empty for early events) |
| `to` | string | Destination phone number (with + prefix) |
| `from` | string | Caller ID / origination number (with + prefix) |
| `direction` | string | Call direction: "inbound", "outbound", or "outbound_sip" |
| `appid` | string | Agora App ID (may be empty if not provided) |
| `video` | boolean | `true` if this is a video call (uses SIP ports 5090/5091); `false` for audio-only |

The `token` value passed in the initiating API call is **deliberately never included** in webhook payloads, so the webhook channel cannot be used to leak channel-join credentials.

### Outbound Call Event Sequence

**Successful call flow**: `call_initiated` → `call_ringing` → `call_answered` → `agora_bridge_start` → `agora_bridge_end` → `call_hangup`

#### 1. call_initiated

Sent when the API receives the call request (Node.js layer). `callid` is empty at this point.

```json
{
  "event": "call_initiated",
  "callid": "",
  "timestamp": 1736953200,
  "uid": "43455",
  "channel": "agora_iok2rg",
  "to": "+447712886300",
  "from": "+441473943851",
  "direction": "outbound",
  "appid": "abc123def456"
}
```

#### 2. call_ringing

Sent 1 second after call_initiated if call hasn't completed yet. `callid` is still empty.

```json
{
  "event": "call_ringing",
  "callid": "",
  "timestamp": 1736953201,
  "uid": "43455",
  "channel": "agora_iok2rg",
  "to": "+447712886300",
  "from": "+441473943851",
  "direction": "outbound",
  "appid": "abc123def456"
}
```

#### 3. call_answered

Sent when the PSTN call is answered. First event with actual `callid`.

```json
{
  "event": "call_answered",
  "callid": "3636eaab-7dfe-4030-8b06-a7d0ff464360",
  "timestamp": 1736953205,
  "uid": "43455",
  "channel": "agora_iok2rg",
  "to": "+447712886300",
  "from": "+441473943851",
  "direction": "outbound",
  "appid": "abc123def456"
}
```

#### 4. agora_bridge_start

Sent when audio bridge to Agora RTC is established.

```json
{
  "event": "agora_bridge_start",
  "callid": "3636eaab-7dfe-4030-8b06-a7d0ff464360",
  "timestamp": 1736953210,
  "uid": "43455",
  "channel": "agora_iok2rg",
  "to": "+447712886300",
  "from": "+441473943851",
  "direction": "outbound",
  "appid": "abc123def456"
}
```

#### 5. agora_bridge_end

Sent when Agora RTC session ends normally.

```json
{
  "event": "agora_bridge_end",
  "callid": "3636eaab-7dfe-4030-8b06-a7d0ff464360",
  "timestamp": 1736953275,
  "uid": "43455",
  "channel": "agora_iok2rg",
  "to": "+447712886300",
  "from": "+441473943851",
  "direction": "outbound",
  "appid": "abc123def456"
}
```

#### 6. call_hangup

**GUARANTEED to fire** when call ends for any reason.

```json
{
  "event": "call_hangup",
  "callid": "3636eaab-7dfe-4030-8b06-a7d0ff464360",
  "timestamp": 1736953276,
  "uid": "43455",
  "channel": "agora_iok2rg",
  "to": "+447712886300",
  "from": "+441473943851",
  "direction": "outbound",
  "hangup_cause": "NORMAL_CLEARING",
  "duration": "75",
  "billsec": "70",
  "sip_disposition": "recv_bye",
  "appid": "abc123def456"
}
```

**Additional fields**:
- `hangup_cause` - FreeSWITCH cause (e.g., `NORMAL_CLEARING`, `USER_BUSY`, `NO_ANSWER`, `CALL_REJECTED`)
- `duration` - Total call duration in seconds
- `billsec` - Billable seconds (conversation time)
- `sip_disposition` - SIP hangup disposition (e.g., `recv_bye`, `send_bye`)

### Inbound Call Event Sequence

**Successful call flow**: `call_initiated` → `agora_bridge_start` → `agora_bridge_end` → `call_hangup`

#### 1. call_initiated

Sent after PIN lookup succeeds. First webhook event for inbound calls.

```json
{
  "event": "call_initiated",
  "callid": "a1b2c3d4-5678-90ef-1234-567890abcdef",
  "timestamp": 1736953115,
  "uid": "54321",
  "channel": "agora_xyz789",
  "to": "7373",
  "from": "benweekes",
  "direction": "inbound",
  "appid": "xyz789abc123",
  "pin": "1234",
  "sip_headers": { "X-Session-Id": "abc" }
}
```

**Additional fields**:
- `pin` - The PIN that was entered by user
- `sip_headers` (object) - **Inbound SIP only.** Custom `X-*` headers
  captured on the inbound `INVITE`. String-valued, one value per name.
  Empty object `{}` when no custom headers were on the wire. Absent for
  inbound PSTN (DID) calls. Mirrors the field delivered in the
  [Inbound Call Lookup](#calllookup) webhook body — same data, different
  delivery point.

#### 2. agora_bridge_start

Sent when audio bridge to Agora RTC is established.

```json
{
  "event": "agora_bridge_start",
  "callid": "a1b2c3d4-5678-90ef-1234-567890abcdef",
  "timestamp": 1736953115,
  "uid": "54321",
  "channel": "agora_xyz789",
  "to": "7373",
  "from": "benweekes",
  "direction": "inbound",
  "appid": "xyz789abc123",
  "sdk_options": "{\"audioProfile\":1,\"audioScenario\":1}"
}
```

**Additional field**: `sdk_options` - JSON string of Agora SDK options

#### 3. agora_bridge_end

Sent when Agora RTC session ends normally.

```json
{
  "event": "agora_bridge_end",
  "callid": "a1b2c3d4-5678-90ef-1234-567890abcdef",
  "timestamp": 1736953245,
  "uid": "54321",
  "channel": "agora_xyz789",
  "to": "+441473943851",
  "from": "+447712886300",
  "direction": "inbound",
  "appid": "xyz789abc123"
}
```

#### 4. call_hangup

**GUARANTEED to fire**. `uid`, `channel`, and `appid` may be empty if call never reached Agora (e.g., invalid PIN).

```json
{
  "event": "call_hangup",
  "callid": "a1b2c3d4-5678-90ef-1234-567890abcdef",
  "timestamp": 1736953246,
  "uid": "54321",
  "channel": "agora_xyz789",
  "to": "+441473943851",
  "from": "+447712886300",
  "direction": "inbound",
  "hangup_cause": "NORMAL_CLEARING",
  "duration": "145",
  "billsec": "131",
  "sip_disposition": "send_bye",
  "appid": "xyz789abc123"
}
```

**Additional fields**:
- `hangup_cause` - FreeSWITCH hangup cause
- `duration` - Total call duration in seconds
- `billsec` - Billable seconds
- `sip_disposition` - SIP hangup disposition

### Asynchronous Events

#### dtmf_received

Sent whenever a DTMF digit is detected on the PSTN/SIP leg of the call. Also fired when digits are pushed onto the leg via the `/send_dtmf` API, so a client that sends DTMF will see its own digits echoed back as a webhook event. Fires for both inbound and outbound calls, any time between `agora_bridge_start` and `call_hangup`.

```json
{
  "event": "dtmf_received",
  "callid": "3636eaab-7dfe-4030-8b06-a7d0ff464360",
  "timestamp": 1736953220,
  "uid": "43455",
  "channel": "agora_iok2rg",
  "to": "+447712886300",
  "from": "+441473943851",
  "direction": "outbound",
  "appid": "abc123def456",
  "digits": "1"
}
```

**Additional field**: `digits` - The DTMF digit(s) detected. Typically a single character (`0`-`9`, `*`, `#`) per event.

#### agora_bridge_failed

Sent when a [Re-bridge](#rebridge) destination-join fails. The SIP leg is
kept up; the call remains active. `call_hangup` is NOT emitted as a result
of this failure. **The call has no active Agora bridge at this point**
(re-bridge is break-before-make) — recover with [Bridge](#bridge) to open
a new Agora session on the same `callid`, or [End Call](#endcall).

```json
{
  "event": "agora_bridge_failed",
  "callid": "3636eaab-7dfe-4030-8b06-a7d0ff464360",
  "timestamp": 1736953220,
  "uid": "123",
  "channel": "agora_new_channel",
  "to": "+447712886300",
  "from": "+441473943851",
  "direction": "inbound",
  "appid": "abc123def456",
  "rebridge": true,
  "reason": "join failed: invalid token"
}
```

**Additional fields**:
- `rebridge` - always `true` on this event.
- `reason` - human-readable failure detail (e.g. invalid token, network error).

### SIP-only Outbound (`outbound_sip`)

`sip_outbound.lua` emits a reduced event set: only `call_answered` and `call_hangup`. The `call_answered` event includes an extra `wav_file` field with the path to the recorded audio. `call_initiated`, `call_ringing`, `agora_bridge_start`, and `agora_bridge_end` are not produced for SIP-only outbound calls.

### Webhook Endpoint Requirements

Your webhook endpoint should:
- Accept HTTP POST requests with JSON body
- Return `200 OK` status
- Respond quickly (< 2 seconds recommended)
- Handle duplicate events gracefully (network retries may occur)

### Delivery Semantics

- **Fire-and-forget.** The gateway sends each webhook in the background and does not block call processing on the response.
- **Timeouts.** 2 second connect timeout, 5 second maximum total timeout.
- **No retries.** A non-200 response, a timeout, or a TCP error is logged but the event is **not** redelivered. Make sure `call_hangup` consumers tolerate occasional gaps.
- **No `token` ever included.** The Agora RTC token passed in the initiating API call is intentionally stripped from all webhook payloads.

## SIP Entrypoints <a name="sipentry"></a>

### Europe
**Region**: Europe

#### Voice

**Transport**: UDP, TCP
     `sip.eu.lb.01.agora.io:5080`

**Transport**: TLS
     `sip.eu.lb.01.agora.io:5081;transport=tls`

#### Video + Voice

**Transport**: UDP, TCP
     `sip.eu.lb.01.agora.io:5090`

**Transport**: TLS
     `sip.eu.lb.01.agora.io:5091;transport=tls`

### USA
**Region**: USA

#### Voice

**Transport**: UDP, TCP
     `sip.usa.lb.01.agora.io:5080`

**Transport**: TLS
     `sip.usa.lb.01.agora.io:5081;transport=tls`

#### Video + Voice

**Transport**: UDP, TCP
     `sip.usa.lb.01.agora.io:5090`

**Transport**: TLS
     `sip.usa.lb.01.agora.io:5091;transport=tls`

### Asia
**Region**: Asia

#### Voice

**Transport**: UDP, TCP
     `sip.as.lb.01.agora.io:5080`

**Transport**: TLS
     `sip.as.lb.01.agora.io:5081;transport=tls`

#### Video + Voice

**Transport**: UDP, TCP
     `sip.as.lb.01.agora.io:5090`

**Transport**: TLS
     `sip.as.lb.01.agora.io:5091;transport=tls`

**Media Encryption**: SRTP (SDES) Optional
**Recommended**: TLS transport for secure signaling

## Agora Gateway IPs <a name="gatewayips"></a>

Please add the following IP addresses to any Access Control Lists which restrict outbound calls from Agora's SIP Gateway by IP addresses.

13.41.31.20/32      
3.9.67.24/32      
52.3.185.227/32      
52.9.29.181/32      
34.233.232.16/32      
3.142.129.19/32      
52.15.168.71/32      
3.150.139.106/32      
3.18.93.182/32      
13.40.252.243/32      
13.41.139.134/32      
13.204.36.207/32      
43.204.1.53/32      

## Configure Twilio <a name="configtwilio"></a>
Configure your own Twilio account to work with the Inbound and Outbound calling APIs above.

[Twilio Inbound](https://drive.google.com/file/d/1HK0vTP9pEsYLFaCP884uLw075qVvbVuv/view?usp=sharing)

[Twilio Outbound](https://drive.google.com/file/d/18XvvCLDhPkhbTJB1YC1JCjP5z9ZTWx06/view?usp=sharing)
