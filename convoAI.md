# PSTN to Agora ConvoAI

When an inbound phone number is dialled, the Agora PSTN gateway will prompt the user for a variable length PIN followed by the # key.  
The PSTN gateway can then call a URL to find out what Agora RTC channel the phone should be connected to and to trigger sending a convoAI agent into the same channel.     

The python code in [lambda_function_convoAI_pstn.py](https://github.com/BenWeekes/agora-rtc-lambda/blob/main/lambda_function_convoAI_pstn.py) can be copied to a lambda function and configured in your AWS account as shown in this demo video: 
📹 [Watch Demo Video](https://drive.google.com/file/d/13mw4jCw62K0YsgffvkCO1KKrme1GP7XB/view?usp=sharing)    

The environment variables below can be set in your lambda function to configure the agent behavior per pin.     
The lambda function URL can then be provided to Agora PSTN admin to be assigned to your inbound phone number.            

## Environment Variables

The function uses PIN-based environment variable lookup. For each variable, it first checks for `VARIABLE_NAME_{PIN}`, then falls back to `VARIABLE_NAME`.

### Mandatory Variables
- **`APP_ID`** - Agora App ID
- **`APP_CERTIFICATE`** - Agora App Certificate (can be empty string if security not enabled for the Agora App ID)
- **`AGENT_AUTH_HEADER`** - Authorization header for Agora API
- **`DEFAULT_PROMPT_{PIN}`** - System prompt for the AI agent (PIN-specific)
- **`DEFAULT_GREETING_{PIN}`** - Greeting message (PIN-specific)
- **`LLM_URL`** - LLM API endpoint
- **`LLM_API_KEY`** - LLM API key
- **`LLM_MODEL`** - LLM model name
- **`TTS_VENDOR`** - Text-to-speech vendor
- **`TTS_KEY`** - TTS API key
- **`TTS_MODEL`** - TTS model name
- **`TTS_VOICE_ID`** - TTS voice ID

### Optional Variables
- **`TTS_VOICE_STABILITY`** - TTS voice stability (default: "1")
- **`TTS_VOICE_SPEED`** - TTS voice speed (default: "1.0")
- **`TTS_VOICE_SAMPLE_RATE`** - TTS sample rate (default: "24000")
- **`DEFAULT_PROMPT`** - Fallback system prompt
- **`DEFAULT_GREETING`** - Fallback greeting message

## Example Environment Variables

```bash
# Agora Configuration
APP_ID=your_agora_app_id
APP_CERTIFICATE=your_agora_app_certificate_or_empty_string
AGENT_AUTH_HEADER=Bearer your_agora_agent_api_token

# LLM Configuration (Groq example)
LLM_URL=https://api.groq.com/openai/v1/chat/completions
LLM_API_KEY=gsk_your_groq_api_key
LLM_MODEL=llama-3.3-70b-versatile

# TTS Configuration (ElevenLabs example)
TTS_VENDOR=elevenlabs
TTS_KEY=your_elevenlabs_api_key
TTS_MODEL=eleven_flash_v2_5
TTS_VOICE_ID=21m00Tcm4TlvDq8ikWAM

# PIN-specific prompts and greetings
DEFAULT_PROMPT_1234=You are a helpful AI assistant. Keep responses short and conversational.
DEFAULT_GREETING_1234=Hello! How can I help you today?
```

## Agora PSTN .pools configuration example entry

```bash
pinlookup_4412345678=https://zeebonggpkkllkks.lambda-url.us-east-1.on.aws
```

📚 [Agora PSTN API Documentation](https://github.com/AgoraIO-Solutions/pstn-doc)
