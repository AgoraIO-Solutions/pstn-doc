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

**Example code:** [Inbound Call Lookup Lambda](CALL-LOOKUP-LAMBDA.md) · [PSTN to ConvoAI](convoAI.md) · [Telephony Go SDK](TELEPHONY-GO.md) · [Video SIP calls: match the SIP side's codecs in the browser](WEB-CODEC-MATCHING.md)

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
- `video` (boolean) [optional]: Video is set by the SIP port the call arrives on, not by this flag: have your provider send the call to a video entrypoint (see [SIP Entrypoints](#sipentry)). Default: `false`
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
- `prompt` (string) [optional]: controls the connection behavior when the callee answers. Default "true". Values:
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
    "reason": "busy",
    "sip_code": "486"
}
```

- `success`: If 'true' the call has connected and 'callid' will be included in the response. If 'false' the call has failed and 'reason' will be included in the response
- `callid` [success only]: This can be used to end the call
- `reason` [failure only]: Why the call failed, usually one of `busy`, `no_answer`, `unavailable`, `declined`, `forbidden`, `not_found`, `cancelled`, `rejected` or `failed`
- `sip_code` [failure only]: The SIP response code, when one is known (may be empty)

### Error Code Responses
401  Unauthorized
500  Missing Parameters, or no result from the gateway (`"msg":"action failed"`)
503  No resource currently available

### Notes
If this API returns success 'true' the call has been connected. If it returns success 'false' there will be a reason explaining the failure.

The request stays open while the number rings, so set your client timeout accordingly: ringing stops after about 60 seconds, and the API answers `500` if the gateway gives no result within 90 seconds.

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
- `video` (boolean) [optional]: Enable video for this call. When `true`, the returned `sip` address uses the video-capable SIP port 5090 (H264+VP8). Default: `false`
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

A call to an unknown, expired or cancelled SIP address is rejected with SIP `403 Forbidden` without being answered.

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
  "sip_headers":{ "X-Session-Id":"abc", "X-Trace-Id":"xyz" },
  "video":true,
  "audio_codecs":["PCMU","OPUS"],
  "video_codecs":["H264","VP8"],
  "audio_codec":"PCMU",
  "video_codec":"H264"
}
```
- `did`: the phone number dialed
- `pin`: the pin entered by caller (`"0000"` if DID not configured for PIN prompts)
- `callerid`: the phone number of caller
- `callid` (string): the gateway's unique identifier for this call — it's the
  same `callid` you will see on lifecycle webhook events. Available at this
  lookup exchange, before any lifecycle event fires. Use it with
  [End Call](#endcall) or [Re-bridge](#rebridge) once your endpoint has
  responded with `200` and an `appid`; until then the API answers `404`.
- `sip_headers` (object): custom `X-*` SIP headers captured on the inbound
  SIP `INVITE`. Always present (empty object `{}` when no custom headers were
  on the wire). String-valued — one value per header name (the gateway never
  sees true on-wire repeats; if the same header name appeared twice, the
  value is the comma-joined string). This applies to calls to a DID too, so
  the object is empty there only if your provider sends no custom headers.
  Inbound SIP `direct_` numbers bypass this webhook.

  Header values come from arbitrary SIP peers and should be treated as
  **untrusted user input** (SQL/HTML/log-injection considerations on your
  side). Size-bounded: ≤20 names, ≤1024 bytes per value, ≤4 KB serialized
  total. Header names prefixed `X-Agora-` are reserved for Agora internal use
  and are never surfaced.
- `video` (boolean): `true` when the call arrived on a video entrypoint (5090, 5091), so it may
  carry video; `false` on the voice entrypoints (5080, 5081), which take audio only.
- `audio_codecs`, `video_codecs` (arrays of strings): the codecs the caller offered, in its order
  of preference (for example `OPUS`, `PCMU`, `PCMA`, `G722`; `H264`, `VP8`, `AV1`). `video_codecs`
  is empty on the voice entrypoints and when the caller offered no video.
- `audio_codec`, `video_codec` (strings): the codecs the call will use, empty when none. On the
  video entrypoints media passes through, so a web client in the channel should publish these:
  create it with them and it joins with matching codecs at once (see
  [WEB-CODEC-MATCHING.md](WEB-CODEC-MATCHING.md)).

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
- `appid` (string) [optional]: the Agora appid for your project. Include it to control this call through the API ([End Call](#endcall), [Cancel Call](#cancelcall), [Bridge](#bridge) and the others): the call is found by `appid` + `callid`, so without it they answer `404`
- `webhook_url` (string) [optional]: your webhook endpoint to receive call lifecycle events (see [Webhook Events](#webhooks))
- `sdk_options` (string) [optional]: JSON string of Agora SDK options (e.g., `{"rtc.client_type":"71"}`)
- `audio_scenario` (string) [optional]: Audio optimization mode. Values:
  - "0": Automatic Human Conferencing Scenarios (default)
  - "10": Optimized To Talk to AI Agents
- `video` (boolean) [optional]: `false` makes the destination voice only: a call with video is
  refused with SIP `488 Video Not Accepted`, so the caller can call again without video. Leave it
  out to accept video when the call arrived on a video entrypoint (see [SIP Entrypoints](#sipentry)).

### Error Code Responses
404  Not Found

### Call answer timing
The gateway does not answer an inbound call while your endpoint decides. The caller hears ringing (SIP `180 Ringing`) until your endpoint responds, and the call is answered only after a successful response, so a rejected call is never answered.

| Your endpoint | Caller receives |
|---------------|-----------------|
| `200 OK` with `token`, `uid` and `channel` | call answered and bridged into the channel |
| `404`, any other error, a body without `token`/`uid`/`channel`, or a connection failure | SIP `403 Forbidden`, call not answered |
| no response within **8 seconds** | SIP `480 Temporarily Unavailable`, call not answered |
| `200 OK` with `"video": false`, on a call with video | SIP `488 Video Not Accepted`, call not answered (unlike `403`, the destination exists: call again without video) |

Respond well within the 8-second limit. Exceptions: numbers configured for PIN prompts are answered first so the caller can hear the prompt and enter the PIN (the lookup runs after the PIN is entered; a failed lookup replays the prompt, up to three attempts, then plays a goodbye message instead of a SIP error), and inbound SIP `direct_` numbers do not use this webhook.

The caller also receives SIP `480 Temporarily Unavailable` if the gateway cannot reach Agora's call service, and SIP `503 Service Unavailable` while that service is not accepting new calls.

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
404  Not Found (callid not active, or the hangup could not be confirmed)
500  Missing Parameters

### Notes

Works for both **outbound and inbound** calls.

`success:true` means the gateway confirmed the call is gone. A `404` with `"msg":"action failed"` means it did not confirm in time (at most 30 seconds): the call may still be up, and you can retry with the same `callid`.

**How to obtain `callid`:**
- **Outbound:** returned in the [Outbound PSTN](#outbound) success response,
  and on webhook lifecycle events from `call_answered` on.
- **Inbound:** delivered in the [Inbound Call Lookup](#calllookup) webhook
  body, *and* on lifecycle events. You can call `endcall` as soon as your
  pinlookup endpoint has responded with `200` and an `appid`, using the
  `callid` you just received — no need to wait for `call_initiated`.

## Cancel Call <a name="cancelcall"></a>
Cancel a call setup request created by Inbound PSTN, Inbound SIP API or Static PIN webhook request. If the call is already in progress, it will be stopped. You can use one of three methods to cancel a call:

- **URL**: `https://sipcm.agora.io/v1/api/pstn`
- **Method**: `POST`

### Method 1: Cancel by Bundle
Use the same `token`, `uid` and `channel` you used when creating the inbound or inboundsip request. Every pending request with these values is cancelled, and a call in progress with them is ended:

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
- `token` (string): the token used to join the channel. If the request was created without a token, pass the appid
- `uid` (string) [optional]: the uid used to join the channel. Omit it only if the request was created without a uid
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
404  Not Found (nothing to cancel, or the call in progress could not be confirmed ended)
500  Missing Parameters

### Notes
This API allows you to cancel a previous call setup request using any of the three methods above. You can cancel calls both before and after they connect - if the call is already in progress, it will be terminated.

A cancelled SIP address is rejected with SIP `403 Forbidden` without being answered. A cancelled PIN no longer connects: a caller who enters it is handled as if the PIN were wrong. A call still setting up when you cancel it, for example during the greeting, ends without joining the channel (see [Webhook Events](#webhooks)). Cancelling the same request twice returns `404`.

## Bridge <a name="bridge"></a>

Attach a connected call's audio to an Agora channel. Used to (re-)establish
the bridge after [Unbridge](#unbridge) or a failed [Re-bridge](#rebridge).
For switching a *currently-bridged* call to a different channel, use
[Re-bridge](#rebridge) instead.

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
- `sdk_options` (string) [optional]: JSON string of Agora SDK options.
  Refused on a call that is still bridged
- `audio_scenario` (string) [optional]: audio optimization mode (see
  [Inbound Call Lookup](#calllookup)). Refused on a call that is still bridged
- `webhook_url` (string) [optional]: accepted but not applied; events keep
  going to the `webhook_url` the call started with

### Success Response Example
```json
{ "success":true }
```

### Error Code Responses
404  Not Found (callid not active)
500  Missing Parameters, or no result from the gateway

### Notes
Works for both inbound and outbound calls. The SIP/PSTN leg stays up
throughout. After an [Unbridge](#unbridge), emits `agora_bridge_start` on
success; the matching `agora_bridge_end` was sent by the unbridge.

On a call that is still bridged, this switches channel in place like
[Re-bridge](#rebridge): the same `rebridge:true` events, `segment_billsec` in
the response, and on a failed join `agora_bridge_failed` with
`{ "success":false, "reason":"join failed", "detail":"..." }`. Passing
`sdk_options` or `audio_scenario` there returns `success:false`; use
[Unbridge](#unbridge) then Bridge to change them.

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
{ "success":true, "parked":true }
```

### Error Code Responses
404  Not Found (callid not active)
500  Missing Parameters, or no result from the gateway

### Notes
Emits `agora_bridge_end` on success. The SIP leg remains active until you
call [End Call](#endcall) or the remote party hangs up. If the call ends
while unbridged, `call_hangup` follows with empty `uid` and `channel` and no
further `agora_bridge_end`.

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
- `sdk_options` (string) [optional]: accepted but not applied; the call keeps
  its SDK settings. To change them, use [Unbridge](#unbridge) then
  [Bridge](#bridge)
- `audio_scenario` (string) [optional]: accepted but not applied, as for
  `sdk_options`
- `webhook_url` (string) [optional]: accepted but not applied; events keep
  going to the `webhook_url` the call started with

### Success Response Example
```json
{ "success":true, "segment_billsec":42 }
```

### Failure Response Examples
```json
{ "success":false, "reason":"current channel/uid mismatch", "channel":"<live channel>", "uid":"<live uid>" }
{ "success":false, "reason":"call not bridged; use bridge" }
{ "success":false, "reason":"rebridge already in progress" }
{ "success":false, "reason":"join failed", "detail":"invalid_token" }
```
On `join failed`, `detail` gives the cause (for example `invalid_token`,
`token_expired`, `connect_timeout`, `connect_rejected` or
`no_bridge_response`), the gateway emits an
[`agora_bridge_failed`](#agora_bridge_failed) lifecycle event and the SIP
leg stays up. **Recovery**: the failure is break-before-make — the call
now has no active Agora bridge, so retrying `/rebridge` will fail with
`call not bridged; use bridge`. Call [Bridge](#bridge) to establish a
fresh Agora session on the same `callid`, or [End Call](#endcall) to
hang up.

### Error Code Responses
404  Not Found (callid not active)
500  Missing required parameters, or no result from the gateway

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

### Error Code Responses
404  Not Found (callid not active)
500  Missing Parameters, or no result from the gateway

### Notes
Once the digits are accepted, one `dtmf_received` webhook event carrying the
whole `digits` string is sent to your `webhook_url`.

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
  "destination":"+18005551234"
}
```
- `appid` (string): the Agora project appid
- `callid` (string): the call id of the connected call
- `destination` (string): phone number or SIP URI to refer the call to. A
  `sip:` or `sips:` URI is referred as given. A phone number goes to the host
  the call is already with: the host that was dialled on an outbound call, the
  domain in the caller's `From` header on an inbound call. If that domain does
  not route calls, pass a SIP URI.
- `leg` (string) [optional]: accepted but has no effect

### Success Response Example
```json
{ "success":true }
```

### Failure Response Example
```json
{ "success":false, "reason":"call not found" }
```

### Error Code Responses
404  Not Found (callid not active)
500  Missing Parameters, or no result from the gateway

### Notes
SIP REFER outcome is up to the upstream provider — success here means the
gateway sent the REFER, not that it was accepted or that the new leg was
answered. The gateway ends its leg once the REFER is sent, so **a refused
REFER still ends the call**. Either way the call's webhook events finish with
`agora_bridge_end`, then `call_hangup`.

## Webhook Events <a name="webhooks"></a>

When you provide a `webhook_url` parameter (in outbound, inbound and inboundsip requests, or in PIN lookup responses), the SIP gateway will send HTTP POST requests to your endpoint with call lifecycle events.

### Events

These webhook events fire for any call regardless of how it was initiated (HTTP API or SDK over WebSocket). An SDK client that sets no `webhook_url` receives them over its WebSocket connection, including `call_hangup` for a call the client ended itself.

| Event | When fired | Direction |
|-------|-----------|-----------|
| [`call_initiated`](#1-call_initiated) | API received the call request (outbound) or PIN lookup succeeded (inbound) | outbound, inbound |
| [`call_ringing`](#2-call_ringing) | Remote party is ringing | outbound |
| [`call_answered`](#3-call_answered) | Remote party answered (SIP-out adds `wav_file`) | outbound, outbound_sip |
| [`agora_bridge_start`](#4-agora_bridge_start) | Audio bridge to Agora RTC starting; the join may still be in progress (on re-bridge: includes `rebridge:true`, `previous_channel`, `previous_uid`) | outbound, inbound |
| [`dtmf_received`](#dtmf_received) | DTMF digit detected on the PSTN leg, or echoed after `/send_dtmf` | outbound, inbound |
| [`agora_bridge_end`](#5-agora_bridge_end) | Agora RTC session ended, including on [Unbridge](#unbridge) (on re-bridge: includes `rebridge:true`, `segment_billsec`) | outbound, inbound |
| [`agora_bridge_failed`](#agora_bridge_failed) | Re-bridge destination-join failed (also a channel switch through [Bridge](#bridge)); SIP leg stays up; `call_hangup` is NOT emitted | outbound, inbound |
| [`call_hangup`](#6-call_hangup) | Call ended (**guaranteed to fire** once a `webhook_url` is known for the call; does NOT fire on re-bridge) | outbound, inbound, outbound_sip |

**Outbound flow**: `call_initiated` → `call_ringing` → `call_answered` → `agora_bridge_start` → `agora_bridge_end` → `call_hangup`

**Inbound flow**: `call_initiated` → `agora_bridge_start` → `agora_bridge_end` → `call_hangup`

**Outbound call never answered**: `call_initiated` → `call_ringing` → `call_hangup` with an empty `callid`.

**Inbound call ended during the greeting** (for example by [Cancel Call](#cancelcall)): sends no `agora_bridge_start` and no `agora_bridge_end`; it still gets `call_hangup`.

**Inbound call rejected at lookup**: no events, because no `webhook_url` is known for it.

**Unbridge and Bridge**: `agora_bridge_end` on [Unbridge](#unbridge) → `agora_bridge_start` on the next [Bridge](#bridge).

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
| `callid` | string | FreeSWITCH UUID for this call session (empty on outbound events before `call_answered`, and on `call_hangup` for an outbound call never answered) |
| `timestamp` | integer | Unix timestamp when event was generated |
| `uid` | string | Agora user ID (may be empty for early events) |
| `channel` | string | Agora channel ID (may be empty for early events) |
| `to` | string | Destination phone number, as dialled |
| `from` | string | Caller ID / origination number, as supplied |
| `direction` | string | Call direction: "inbound", "outbound", or "outbound_sip" |
| `appid` | string | Agora App ID (may be empty if not provided) |
| `video` | boolean | `true` if this is a video call (uses SIP ports 5090/5091); `false` for audio-only |

Exceptions: `to` and `from` are not sent on `dtmf_received`, on `agora_bridge_failed`, or on the `rebridge:true` events; the last two also omit `direction`, as does the `dtmf_received` that follows `/send_dtmf`. `outbound_sip` events carry empty `uid`, `channel`, `to` and `from`, and no `video`.

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

Sent when the audio bridge to Agora RTC starts. The join may still be in progress.

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

**GUARANTEED to fire** when call ends for any reason. An outbound call that was never answered sends it with an empty `callid`.

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
  "sip_response_code": "",
  "sip_response_phrase": "",
  "call_status": "answered",
  "appid": "abc123def456"
}
```

**Additional fields**:
- `hangup_cause` - FreeSWITCH cause (e.g., `NORMAL_CLEARING`, `USER_BUSY`, `NO_ANSWER`, `CALL_REJECTED`)
- `duration` - Total call duration in seconds
- `billsec` - Billable seconds (conversation time)
- `sip_disposition` - SIP hangup disposition (e.g., `recv_bye`, `send_bye`)
- `sip_response_code` - SIP response code that ended the call (empty if none)
- `sip_response_phrase` - SIP response phrase (empty if none)
- `call_status` - Outcome: `answered`, `busy`, `no_answer`, `unavailable`, `declined`, `forbidden`, `not_found`, `cancelled`, `rejected` or `failed`

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
- `pin` - The PIN that was entered by user (`"0000"` on a DID not configured for PIN prompts, the address digits on an inbound SIP address, empty on a `direct_` number)
- `sip_headers` (object) - Custom `X-*` headers
  captured on the inbound `INVITE`. String-valued, one value per name.
  Empty object `{}` when no custom headers were on the wire. Mirrors the field delivered in the
  [Inbound Call Lookup](#calllookup) webhook body — same data, different
  delivery point.

#### 2. agora_bridge_start

Sent when the audio bridge to Agora RTC starts. The join may still be in progress.

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

**GUARANTEED to fire** once a `webhook_url` is known for the call, which for inbound calls means after a successful lookup. A call rejected at lookup sends no events. `uid` and `channel` are empty if the call ended while unbridged.

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

Sent whenever a DTMF digit is detected on the PSTN/SIP leg of the call. Also fired once when digits are pushed onto the leg via the `/send_dtmf` API, so a client that sends DTMF will see its own digits echoed back as a webhook event. Fires for both inbound and outbound calls, any time between `agora_bridge_start` and `call_hangup`. Digits pressed during the bridge and the `/send_dtmf` echo are sent outside the per-call queue (see [Delivery Semantics](#delivery-semantics)).

```json
{
  "event": "dtmf_received",
  "callid": "3636eaab-7dfe-4030-8b06-a7d0ff464360",
  "timestamp": 1736953220,
  "uid": "43455",
  "channel": "agora_iok2rg",
  "direction": "outbound",
  "appid": "abc123def456",
  "video": false,
  "digits": "1"
}
```

**Additional field**: `digits` - The DTMF digit detected, a single character (`0`-`9`, `*`, `#`) per event. On the `/send_dtmf` echo, the whole string that was sent.

#### agora_bridge_failed

Sent when a [Re-bridge](#rebridge) destination-join fails, or the same
failure on a channel switch through [Bridge](#bridge). The SIP leg is
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
  "appid": "abc123def456",
  "video": false,
  "rebridge": true,
  "reason": "join failed",
  "detail": "invalid_token"
}
```

**Additional fields**:
- `rebridge` - always `true` on this event.
- `reason` - always `join failed` on this event.
- `detail` - the cause (e.g. `invalid_token`, `token_expired`, `connect_timeout`, `no_bridge_response`).

### SIP-only Outbound (`outbound_sip`)

`sip_outbound.lua` emits a reduced event set: only `call_answered` and `call_hangup`. The `call_answered` event includes an extra `wav_file` field with the path to the audio file played on the call. `call_initiated`, `call_ringing`, `agora_bridge_start`, and `agora_bridge_end` are not produced for SIP-only outbound calls.

### Webhook Endpoint Requirements

Your webhook endpoint should:
- Accept HTTP POST requests with JSON body
- Return `200 OK` status
- Respond quickly (< 2 seconds recommended): a slow response holds back that call's later events
- Tolerate gaps: events are not retried (see below)

### Delivery Semantics

- **Fire-and-forget.** The gateway sends each webhook in the background and does not block call processing on the response.
- **In order, one at a time, per call.** `call_answered`, `agora_bridge_start`, `agora_bridge_end`, `call_hangup` and inbound `call_initiated` go through a queue per call: each is sent after the one before it has been answered or has timed out, so `agora_bridge_end` arrives before `call_hangup`. A slow or dead endpoint delays only that call's later events, by up to 5 seconds per event. Other calls are not affected.
- **Outside the queue.** These are sent independently and can overlap or overtake queued events: `dtmf_received` for digits pressed during the bridge or echoed after `/send_dtmf`, `call_initiated` and `call_ringing` on outbound calls, `call_hangup` on an outbound call never answered, and the `rebridge:true` and `agora_bridge_failed` events. Compare `timestamp` (whole seconds) if order matters.
- **Timeouts.** 2 second connect timeout, 5 second maximum total timeout. Events outside the queue, other than digits pressed during the bridge, use a single 5 second timeout.
- **No retries.** After a non-200 response, a timeout, or a TCP error the event is **not** redelivered. Make sure `call_hangup` consumers tolerate occasional gaps.
- **No `token` ever included.** The Agora RTC token passed in the initiating API call is intentionally stripped from all webhook payloads.

## SIP Entrypoints <a name="sipentry"></a>

Each region has two pairs of ports:

- **Voice (5080, 5081):** audio only. The gateway transcodes the audio between your SIP codec and Agora. A video offer on these ports is answered with audio only.
- **Video + Voice (5090, 5091):** video calls, and voice-only calls too. Audio and video pass through without transcoding when the Agora side uses the same codecs (see [WEB-CODEC-MATCHING.md](WEB-CODEC-MATCHING.md)).

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

On video entrypoints the gateway passes video and audio through, so the browser must publish the SIP client's codecs: see [Video SIP calls: match the SIP side's codecs](WEB-CODEC-MATCHING.md).

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
