# Agora PSTN & SIP Gateway


1. [Overview](#overview)
2. [Inbound PSTN](#inbound)
3. [Outbound PSTN](#outbound)
4. [Inbound SIP](#inboundsip)
5. [Static PIN](#staticpin)
6. [End Call](#endcall)
7. [Agora Gateway IPs](#gatewayips)
8. [Twilio Configuration](#configtwilio)
 
## Overview <a name="overview"></a>
These REST APIs allow developers to trigger inbound and outbound PSTN and SIP calls which then connect into an Agora channel enabling end-users to participate with their phone for the audio leg of the conference call.     
Please contact us to provision your appid with this service. We will provide you with an authorization header to include in the API requests below.

## Inbound PSTN <a name="inbound"></a>
In this scenario, the end-user dials a phone number displayed to them and enters the PIN when prompted. With the correct PIN, they are connected into the Agora channel.

- **URL**: `https://sipcm.agora.io/v1/api/pstn`
- **Method**: `POST`

### Request Body Parameters as JSON
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

### Request Body Parameters as JSON
```json
{
  "action":"outbound", 
  "appid":"fs9f52d9dcc1f406b93d97ff1f43c554f",
  "token":"NGMxYWRlMGFkYTRjNGQ2ZWFiNTFmYjMz",
  "uid":"123",
  "channel":"pstn",
  "region":"AREA_CODE_NA",
  "prompt":"true",
  "to":"+447712886400#333",
  "from":"+1800222333",
  "timeout":"3600",
  "sip":"acme.pstn.ashburn.twilio.com"
}
```
- `appid` (string) the Agora project appid
- `token` (string) [optional]: a generated access token
- `uid` (string or int) [optional]: a user uid
- `channel` (string): an Agora channel name
- `prompt` (string): play the callee a voice prompt and wait for them to press a digit. If set to "lazy" then any DTMF may be pressed.
- `to` (string): the end-user's phone number to dial. You can optionally add a # followed by numbers which will be played as DTMF once the call connects.      
- `from` (string): the calling number displayed on the end-user's phone during ringing
- `sip` (string): termination sip uri or leave blank if being routed by this service   
- `timeout` (string) [optional]: max duration for outbound call in seconds. Default 3600 seconds which is 1 hour.   
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
    "callid": "88877-55Asdd7-55Asdd",
    "reason": "busy",

}
```
- `success` If 'true' the call has connected and 'callid' will be included in the response. If 'false' the call has failed and 'reason' will be included in the response     
- `callid` [optional] This can be used to end the call        
- `reason` [optional] The call failed because the user was 'busy' or the destination was 'invalid'           

### Error Code Responses       
401  Unauthorized      
500  Missing Parameters       
503  No resource currently available      

### Notes     
If this API returns success 'true' the call has been connected. If it returns success false there will be a reason 

## Inbound SIP <a name="inboundsip"></a>
In this scenario, an inbound SIP address is requested. When the SIP address is dialled, the call will be routed to the requested user/channel session.

- **URL**: `https://sipcm.agora.io/v1/api/pstn`
- **Method**: `POST`

### Request Body Parameters as JSON
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

## Static PIN <a name="staticpin"></a>
If provisioned, the service can call out to an external REST endpoint providing the number dialed and pin entered.
The REST endpoint can choose to accept the PIN and return the details needed to join the user to the channel or error status 404 if PIN not valid.

- **URL**: `https://example.customer.com/api/pinlookup`
- **Method**: `POST`

### Request Body Parameters as JSON
```json
{
  "did":"17177440111", 
  "pin":"334455",
}
```
- `did` the phone number dialed
- `pin` the pin entered

### Success Response Example
*Status Code*: `200 OK`    
*Content Type*: `application/json`    
Body:
```json
{  
  "token":"NGMxYWRlMGFkYTRjNGQ2ZWFiNTFmYjMz",
  "uid":"123",
  "channel":"test",
  "appid":"fs9f52d9dcc1f406b93d97ff1f43c554f",
}
```    

- `token` (string): a generated access token OR the appid if tokens are not enabled
- `uid` (string or int): a user uid
- `channel` (string): an Agora channel name
- `appid` (string): the Agora appid for your project

### Error Code Responses       
404  Not Found  

### Notes
This API allows you to give your users a pin that will not expire.

## End Call <a name="endcall"></a>
Use the callid returned by the outbound call API to terminate an outbound call.    

- **URL**: `https://sipcm.agora.io/v1/api/pstn`
- **Method**: `POST`

### Request Body Parameters as JSON
```json
{
  "action":"endcall",
  "appid":"fs9f52d9dcc1f406b93d97ff1f43c554f",
  "callid":"f577605c-eb3a-4efe-af1b-ee66d5297569",
}
```
- `appid` (string) the Agora project appid
- `callid` (string) the call id of the ongoing call    

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
This API allows you to stop an outbound call. 


## Agora Gateway IPs <a name="gatewayips"></a>       

### Outbound IPs      
Please add the following IP addresses to any Access Control Lists which restrict outbound calling by IP.

3.9.67.24    
52.3.185.227     
52.9.29.181     
34.233.232.16       

### Inbound IPs      
Please point inbound calls at one of these IPs with the other being a failover.     
52.3.185.227       
52.9.29.181       

## Configure Twilio <a name="configtwilio"></a>       
Configure your own Twilio account to work with the Inbound and Outbound calling APIs above.      

[Twilio Inbound](https://drive.google.com/file/d/1HK0vTP9pEsYLFaCP884uLw075qVvbVuv/view?usp=sharing)      

[Twilio Outbound](https://drive.google.com/file/d/18XvvCLDhPkhbTJB1YC1JCjP5z9ZTWx06/view?usp=sharing)      


