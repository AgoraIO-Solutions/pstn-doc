# Video SIP calls: publish the SIP side's codecs from the browser

On a video SIP call the gateway passes video and audio through without transcoding. This
keeps quality and latency at their best and uses much less gateway CPU, but it means the browser in the Agora channel must publish the **same codecs the SIP client
negotiated**. If it publishes something else, the SIP client gets no video from it (and, for
audio, hears nothing).

| Call | What the browser must publish |
|---|---|
| Video SIP call (video entrypoint, see [SIP Entrypoints](README.md#sipentry)) | the SIP client's video codec (H264, VP8 or AV1) and audio codec (Opus, PCMU, PCMA or G.722) |
| Audio-only SIP call | nothing special: the gateway converts the audio and publishes Opus |

The SIP client chooses its codecs, so the browser cannot know them in advance. The sample
below finds them from the first packets it receives and switches to them.

## How it works

The Agora Web SDK fixes both codecs when the client is created
(`AgoraRTC.createClient({ codec, audioCodec })`). To change them:

1. Join as usual (for example with H264 and Opus) and subscribe to the gateway's user.
2. Read the codec the gateway sends: `client.getRemoteVideoStats()[uid].codecType` and
   `client.getRemoteAudioStats()[uid].codecType`. They appear about a second after the first
   packets.
3. If either differs from yours, leave, create a new client with the gateway's codecs, join
   again with the same uid (fetch a fresh token from your server) and publish the same tracks.
   The camera and microphone stay open, so the user sees only a short gap.

The switch takes a few seconds in total. Do not switch back and forth when two SIP users with
different codecs are in one channel: the sample stops after a few switches.

```js
async function connect(codec, acodec) {
  const client = AgoraRTC.createClient({ mode: 'rtc', codec, audioCodec: acodec });
  st.client = client; st.codec = codec; st.acodec = acodec;
  client.on('user-published', onPublished);
  await client.join(APP_ID, channel, await getToken(), uid);   // token from your server
  await client.publish([st.mic, st.cam]);
}

async function onPublished(user, kind) {
  const client = st.client;
  await client.subscribe(user, kind);
  kind === 'video' ? user.videoTrack.play('remote') : user.audioTrack.play();
  for (let i = 0; i < 40 && st.client === client; i++) {          // up to 10 s
    const s = kind === 'video' ? client.getRemoteVideoStats()[user.uid]
                               : client.getRemoteAudioStats()[user.uid];
    if (s && s.codecType) {
      const c = String(s.codecType).toLowerCase();                  // h264, vp8, av1 | opus, pcmu, pcma, g722
      st.want[kind] = c === 'av1x' ? 'av1' : c;
      const codec = st.want.video || st.codec, acodec = st.want.audio || st.acodec;
      if (codec !== st.codec || acodec !== st.acodec) await rejoin(codec, acodec);
      return;
    }
    await new Promise((r) => setTimeout(r, 250));
  }
}

async function rejoin(codec, acodec) {
  if (st.switching || st.switches >= 4) return;
  st.switching = true; st.switches++;
  try {
    const old = st.client; st.client = null;
    await old.leave();                // tracks stay open
    await connect(codec, acodec);
  } finally { st.switching = false; }
}
```

Full page, ready to open in a browser:
[examples/web-codec-matching/index.html](examples/web-codec-matching/index.html). Enter your
App ID, a channel, a token for that channel and uid, and join; then have the SIP client call
into the same channel (for example with the address returned by [Inbound SIP](README.md#inboundsip)).

Notes:

- Never put your App Certificate in a web page. Issue tokens from your own server.
- The page needs `https://` (or `localhost`) for the camera and microphone.
- Tested on 2026-11-09 with Chrome and the Agora Web SDK 4.24.9 against SIP clients sending
  VP8 with PCMU, AV1 with Opus, and H264 with PCMA: the page switched each time, and video and
  audio flowed both ways.
- Ask us for a link to our hosted test page, which does the same and shows the codecs, size,
  frame rate and loss on both sides.
