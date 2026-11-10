# SIP calls on the video entrypoints: publish the SIP side's codecs from the browser

On the video entrypoints (5090, 5091) the gateway passes video and audio through without
transcoding, for video calls and voice-only calls alike. This
keeps quality and latency at their best and uses much less gateway CPU, but it means the browser in the Agora channel must publish the **same codecs the SIP client
negotiated**. If it publishes something else, the SIP client gets no video from it (and, for
audio, hears nothing).

| Call | What the browser must publish |
|---|---|
| Video call on a video entrypoint (5090, 5091; see [SIP Entrypoints](README.md#sipentry)) | the SIP client's video codec (H264, VP8 or AV1) and audio codec (Opus, PCMU, PCMA or G.722) |
| Voice-only call on a video entrypoint | the SIP client's audio codec |
| Call on a voice entrypoint (5080, 5081) | nothing special: the gateway transcodes the audio and publishes Opus; no video |

The SIP client chooses its codecs, so the browser cannot know them in advance. The sample
below finds them from the first packets it receives and switches to them.

## How it works

The Agora Web SDK fixes both codecs when the client is created
(`AgoraRTC.createClient({ codec, audioCodec })`). To change them:

1. Join as usual (for example with H264 and Opus) and subscribe to the gateway's user.
2. Read the codec the gateway sends: `client.getRemoteVideoStats()[uid].codecType` and
   `client.getRemoteAudioStats()[uid].codecType`. They appear about a second after the first
   packets.
3. Wait until **both** are known, then switch once: if either differs from yours, leave,
   create a new client with the gateway's codecs, join again with the same uid (fetch a fresh
   token from your server) and publish the same tracks. The camera and microphone stay open.
   A SIP caller's audio usually arrives about a second before its video, so after the audio
   codec wait up to 3 s for a video.
4. Render the remote video only once its codecs match yours. With a switch, the user then sees
   the remote picture once, from the new client, instead of a picture that disappears and comes
   back. Ignore events from the client you replaced.

The switch takes a few seconds in total. Do not switch back and forth when two SIP users with
different codecs are in one channel: the sample stops after a few switches.

```js
async function connect(codec, acodec) {
  const client = AgoraRTC.createClient({ mode: 'rtc', codec, audioCodec: acodec });
  st.client = client; st.codec = codec; st.acodec = acodec;
  client.on('user-published', (user, kind) => { if (st.client === client) onPublished(client, user, kind); });
  await client.join(APP_ID, channel, await getToken(), uid);   // token from your server
  await client.publish([st.mic, st.cam]);
}

async function onPublished(client, user, kind) {
  await client.subscribe(user, kind);
  if (st.client !== client) return;                            // replaced while subscribing
  const r = st.remotes.get(user.uid) || {}; st.remotes.set(user.uid, r);
  if (kind === 'audio') user.audioTrack.play();                // video waits for the codecs
  for (let i = 0; i < 40 && st.client === client; i++) {       // up to 10 s
    const s = kind === 'video' ? client.getRemoteVideoStats()[user.uid]
                               : client.getRemoteAudioStats()[user.uid];
    if (s && s.codecType) {
      const c = String(s.codecType).toLowerCase();             // h264, vp8, av1 | opus, pcmu, pcma, g722
      r[kind] = c === 'av1x' ? 'av1' : c;
      return decide(client, user, r);
    }
    await new Promise((res) => setTimeout(res, 250));
  }
}

async function decide(client, user, r) {                       // one switch for both codecs
  if (st.switching || r.deciding) return;
  r.deciding = true;
  for (let t = 0; t < 3000 && st.client === client; t += 250) {
    if ((user.hasVideo ? r.video : false) && (!user.hasAudio || r.audio)) break;
    await new Promise((res) => setTimeout(res, 250));
  }
  if (st.client !== client) return;
  const codec = r.video || st.codec, acodec = r.audio || st.acodec;
  if (codec !== st.codec || acodec !== st.acodec) return rejoin(codec, acodec);
  if (user.hasVideo) user.videoTrack.play('remote');           // codecs match: render once
}

async function rejoin(codec, acodec) {
  if (st.switches >= 4) return;
  st.switching = true; st.switches++;
  try {
    const old = st.client; st.client = null; st.remotes.clear();
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
- The same logic runs in our hosted test page (one re-join, remote rendered once; checked on
  2026-11-10 against Linphone and baresip with H264, VP8 and AV1, Opus, PCMU, PCMA and G.722).
- Tested on 2026-11-09 with Chrome and the Agora Web SDK 4.24.9 against SIP clients sending
  VP8 with PCMU, AV1 with Opus, and H264 with PCMA: the page switched each time, and video and
  audio flowed both ways.
- Ask us for a link to our hosted test page, which does the same and shows the codecs, size,
  frame rate and loss on both sides.
