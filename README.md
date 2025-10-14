# Agora PSTN & SIP Gateway


1. [Overview](#overview)
2. [Inbound PSTN](#inbound)
3. [Outbound PSTN](#outbound)
4. [Inbound SIP](#inboundsip)
5. [Static PIN](#staticpin)
6. [End Call](#endcall)
7. [Cancel Call](#cancelcall)
8. [SIP Entrypoints](#sipentry)
9. [Agora Gateway IPs](#gatewayips)  
10. [Twilio Configuration](#configtwilio)
 
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
  "to":"+447712886400#333",
  "from":"+1800222333",
  "timeout":"3600",
  "sip":"acme.pstn.ashburn.twilio.com"
}
```
- `appid` (string): the Agora project appid
- `token` (string) [optional]: a generated access token
- `uid` (string) [optional]: a user uid
- `channel` (string): an Agora channel name
- `prompt` (string): play the callee a voice prompt and wait for them to press a digit. If set to "lazy" then any DTMF may be pressed
- `to` (string): the end-user's phone number to dial. You can optionally add a # followed by numbers which will be played as DTMF once the call connects      
- `from` (string): the calling number displayed on the end-user's phone during ringing
- `sip` (string) [optional]: termination sip uri or leave blank if being routed by this service   
- `timeout` (string) [optional]: max duration for outbound call in seconds. Default 3600 seconds which is 1 hour   
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

## Static PIN <a name="staticpin"></a>
If provisioned, the service can call out to an external REST endpoint (that you implement) providing the number dialed and pin entered. Your REST endpoint can choose to accept the PIN and return the details needed to join the user to the channel or return error status 404 if PIN is not valid.

**This is a webhook that YOU implement, not an Agora API endpoint.**

- **URL**: `https://example.customer.com/api/pinlookup` *(your endpoint)*
- **Method**: `POST`

### Request Body Parameters as JSON (sent by Agora to your endpoint)
```json
{
  "did":"17177440111", 
  "pin":"334455"
}
```
- `did`: the phone number dialed
- `pin`: the pin entered

### Success Response Example (from your endpoint)
*Status Code*: `200 OK`    
*Content Type*: `application/json`    
Body:
```json
{  
  "token":"006d2776fef40dc864dc7a438e3b871IACGnkorQsr0iCpBmzNwKEdzKuAv1b1zRcMy0cOradw6mHN/il8AAAAAIgCQqbAFeaVzZgQAAQDw0oBmAgDw0oBmAwDw0oBmBADw0oBm",
  "uid":"123",
  "channel":"agora_channel",
  "appid":"your_appid_here"
}
```    

- `token` (string): a generated access token OR the appid if tokens are not enabled
- `uid` (string): a user uid
- `channel` (string): an Agora channel name
- `appid` (string): the Agora appid for your project

### Error Code Responses       
404  Not Found  

### Notes
This webhook allows you to give your users a PIN that will not expire. When users dial your DID and enter a PIN, Agora will call your endpoint to validate the PIN and get the channel details.

## End Call <a name="endcall"></a>
Use the callid returned by the outbound call API to terminate an outbound call.    

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
This API allows you to stop an outbound call that is currently in progress.

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

## SIP Entrypoints <a name="sipentry"></a>       

### Europe
**Region**: Europe   
**Transport**: UDP, TCP  
     `sip.eu.lb.01.agora.io:5080`   
**Transport**: TLS  
     `sip.eu.lb.01.agora.io:5081;transport=tls`   

### USA
**Region**: USA   
**Transport**: UDP, TCP  
     `sip.usa.lb.01.agora.io:5080`    
**Transport**: TLS   
     `sip.usa.lb.01.agora.io:5081;transport=tls`    

### Asia 
**Region**: Asia    
**Transport**: UDP, TCP    
     `sip.as.lb.01.agora.io:5080`    
**Transport**: TLS  
     `sip.as.lb.01.agora.io:5081;transport=tls`    

**Media Encryption**: SRTP (SDES) Optional      
**Recommended**: TLS transport for secure signaling       

## Agora Gateway IPs <a name="gatewayips"></a>       

Please add the following IP addresses to any Access Control Lists which restrict outbound calls from Agora's SIP Gateway by IP addresses.

13.41.31.20       
3.9.67.24           
52.3.185.227     
52.9.29.181     
34.233.232.16       
3.142.129.19     
52.15.168.71      
3.150.139.106       
3.18.93.182       
13.40.252.243      
13.41.139.134      
13.204.36.207      
43.204.1.53        
 

## Configure Twilio <a name="configtwilio"></a>       
Configure your own Twilio account to work with the Inbound and Outbound calling APIs above.      

[Twilio Inbound](https://drive.google.com/file/d/1HK0vTP9pEsYLFaCP884uLw075qVvbVuv/view?usp=sharing)      

[Twilio Outbound](https://drive.google.com/file/d/18XvvCLDhPkhbTJB1YC1JCjP5z9ZTWx06/view?usp=sharing)
