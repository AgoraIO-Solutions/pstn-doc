# Agora PSTN 


1. [Overview](#overview)
2. [Inbound PSTN](#inbound)
3. [Outbound PSTN](#outbound)
4. [Inbound SIP](#inboundsip)
5. [Agora Gateway IPs](#gatewayips)
 
## Overview <a name="overview"></a>
These REST APIs allow developers to trigger inbound and outbound PSTN calls which then connect into an Agora channel enabling end-users to participate with their phone for the audio leg of the conference call.     
Please contact us to provision your appid with this service. We will provide you with an authorization header to include in the API requests below.

## Inbound PSTN <a name="inbound"></a>
In this scenario, the end-user dials a phone number displayed to them and enters the PIN when prompted. With the correct PIN, they are connected into the Agora channel.

- **URL**: `https://sipcm.agora.io/v1/api/pstn`
- **Method**: `POST`

### Request Parameters
```json
{
  "action":"inbound", 
  "appid":"fs9f52d9dcc1f406b93d97ff1f43c554f",
  "token":"NGMxYWRlMGFkYTRjNGQ2ZWFiNTFmYjMz",
  "uid":"123",
  "channel":"pstn",
  "region":"AREA_CODE_NA",
}
```
- `appid` (string) the Agora project appid
- `token` (string) [optional]: a generated access token
- `uid` (string or int) [optional]: a user uid
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
    "pin": "780592",
}
```    

- `did` the phone number to dial
- `display` the phone number to dial in a more friendly display format
- `pin` the pin to enter when prompted

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

### Request Parameters
```json
{
  "action":"outbound", 
  "appid":"fs9f52d9dcc1f406b93d97ff1f43c554f",
  "token":"NGMxYWRlMGFkYTRjNGQ2ZWFiNTFmYjMz",
  "uid":"123",
  "channel":"pstn",
  "region":"AREA_CODE_NA",
  "prompt":"true",
  "to":"+447712886400",
  "from":"+1800222333",
  "sip":"acme.pstn.ashburn.twilio.com"
}
```
- `appid` (string) the Agora project appid
- `token` (string) [optional]: a generated access token
- `uid` (string or int) [optional]: a user uid
- `channel` (string): an Agora channel name
- `prompt` (string): play the callee a voice prompt and wait for them to press a digit. If set to "lazy" then any DTMF may be pressed.
- `to` (string): the end-user's phone number to dial
- `from` (string): the calling number displayed on the end-user's phone during ringing
- `sip` (string): termination sip uri or leave blank if using shared pool    
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
    "tid": "8887755Asdd",
}
```
- `tid` a transaction id

### Error Code Responses       
401  Unauthorized      
500  Missing Parameters       
503  No resource currently available      


## Inbound SIP <a name="inboundsip"></a>
In this scenario, an inbound SIP address is requested. When the SIP address is dialled, the call will be routed to the requested user/channel session.

- **URL**: `https://sipcm.agora.io/v1/api/pstn`
- **Method**: `POST`

### Request Parameters
```json
{
  "action":"inboundsip", 
  "appid":"fs9f52d9dcc1f406b93d97ff1f43c554f",
  "token":"NGMxYWRlMGFkYTRjNGQ2ZWFiNTFmYjMz",
  "uid":"123",
  "channel":"test",
  "region":"AREA_CODE_NA",
}
```
- `appid` (string) the Agora project appid
- `token` (string) [optional]: a generated access token
- `uid` (string or int) [optional]: a user uid
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
    "sip": "sip:lookup_783410988998@2.4.6.7:5080",
}
```    

- `sip` the sip address to dial to join the call

### Error Code Responses       
401  Unauthorized      
500  Missing Parameters     
503  No resource currently available      


### Notes
Using this API you can bridge an outbound call from your provider with an inbound sip address into Agora.

## Agora Gateway IPs <a name="gatewayips"></a>       

Please add the following IP addresses to any Access Control Lists which restrict calling by IP.

3.9.67.24    
52.3.185.227     
52.9.29.181     
34.233.232.16       

