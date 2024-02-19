# Agora PSTN 


## Overview    
These REST APIs allow developers to trigger inbound and outbound PSTN calls which then connect into an Agora channel enabling end-users to participate with their phone for the audio leg of the conference call.

## Inbound PSTN
In this scenario, the end-user dials a phone number displayed to them and enters the PIN when prompted.

- **URL**: `/v1/api/pstn`
- **Method**: `POST`

### Request Parameters
```json
{
  "action":"inbound", 
  "appId":"fs9f52d9dcc1f406b93d97ff1f43c554f",
  "token":"NGMxYWRlMGFkYTRjNGQ2ZWFiNTFmYjMz",
  "uid":"123",
  "channel":"pstn",
  "region":"AREA_CODE_NA",
}
```
- `appId` (string) the Agora project appId
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
    "status": 200,
    "did": "17377583411",
    "display": "+1 (737) 758 3411",
    "pin": "780592",
}
```    

- `did` the phone number to dial
- `display` the phone number to dial in a more friendly display format
- `pin` the pin to enter when prompted

### Notes
Direct Inward Dialing (DID) providers such as Twilio allow you to buy a phone number and have the calls forwarded to a SIP address. We will provide you with the SIP address to forward your calls to when we provision you on this service. You will also be able to customise the voice prompts played to your end-users. 


## Outbound PSTN
In this scenario, the end-user receives a phone call which connects them directly to the channel when they answer. 

- **URL**: `/v1/api/pstn`
- **Method**: `POST`

### Request Parameters
```json
{
  "action":"outbound", 
  "appId":"fs9f52d9dcc1f406b93d97ff1f43c554f",
  "token":"NGMxYWRlMGFkYTRjNGQ2ZWFiNTFmYjMz",
  "uid":"123",
  "channel":"pstn",
  "to":"+447712886400",
  "from":"+1800222333",
  "sip":{provider:"twilio",username:"tw",password:"tw","uri":"twilio22.pstn.ashburn.twilio.com"}
}
```
- `appId` (string) the Agora project appId
- `token` (string) [optional]: a generated access token
- `uid` (string or int) [optional]: a user uid
- `channel` (string): an Agora channel name
- `to` (string): the end-user's phone number to dial
- `from` (string): the calling number displayed on the end-user's phone during ringing
- `sip` (string): the credentials needed to place the sip call via a supported provider.


### Success Response Example
*Status Code*: `200 OK`    
*Content Type*: `application/json`    
Body:
```json
{  
    "status": 200,
    "tid": "8887755Asdd",
}
```
- `tid` a transaction id 

