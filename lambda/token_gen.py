import json
import hmac
from hashlib import sha256
import base64
import struct
import zlib
import secrets
import time
import random
import string
from collections import OrderedDict
import os


def get_env_var(var_name, default_value=None):
    """
    Gets an environment variable with an optional default.

    Args:
        var_name: The environment variable name
        default_value: Default value if variable is not set

    Returns:
        The value of the environment variable or default_value
    """
    value = os.environ.get(var_name)
    if value is not None:
        return value
    return default_value


def generate_random_channel(length=10):
    """
    Generates a random channel name with uppercase letters and numbers.

    Args:
        length: Length of the channel name (default: 10)

    Returns:
        Random channel name string
    """
    return ''.join(random.choices(string.ascii_uppercase + string.digits, k=length))


def lambda_handler(event, context):
    """
    Lambda handler for PSTN CallLookup webhook.

    Receives POST {did, pin, callerid} from the Agora PSTN gateway
    and returns {token, uid, channel, appid, audio_scenario} so the
    gateway knows which RTC channel to connect the caller to.
    """
    # Parse POST body
    body = {}
    if event.get('body'):
        try:
            body = json.loads(event['body'])
        except (json.JSONDecodeError, TypeError):
            body = {}

    did = body.get('did', '')
    pin = body.get('pin', '')
    callerid = body.get('callerid', '')

    print(f"CallLookup request: did={did} pin={pin} callerid={callerid}")

    # Read env vars
    app_id = get_env_var('APP_ID', '')
    app_certificate = get_env_var('APP_CERTIFICATE', '')
    user_uid = get_env_var('USER_UID', '101')
    audio_scenario = get_env_var('AUDIO_SCENARIO', '0')
    webhook_url = get_env_var('WEBHOOK_URL')
    sdk_options = get_env_var('SDK_OPTIONS')

    # Generate random channel name
    channel = generate_random_channel(10)

    # Build token
    has_certificate = app_certificate and app_certificate.strip() != ''
    if has_certificate:
        token = build_rtc_token(app_id, app_certificate, channel, user_uid)
    else:
        token = app_id

    # Build response
    response_data = {
        "token": token,
        "uid": user_uid,
        "channel": channel,
        "appid": app_id,
        "audio_scenario": audio_scenario,
    }

    if webhook_url:
        response_data["webhook_url"] = webhook_url
    if sdk_options:
        response_data["sdk_options"] = sdk_options

    return json_response(200, response_data)


def json_response(status_code, body):
    """Helper function to create a JSON response."""
    return {
        "statusCode": status_code,
        "headers": {
            "Content-Type": "application/json"
        },
        "body": json.dumps(body)
    }


def build_rtc_token(app_id, app_certificate, channel, uid):
    """
    Builds a v007 token with RTC privileges only (no RTM).

    Args:
        app_id: Agora App ID
        app_certificate: Agora App Certificate
        channel: Channel name
        uid: User ID string

    Returns:
        v007 token string
    """
    privilege_expire = 24 * 3600  # 24 hours

    token = AccessToken007(app_id, app_certificate)

    rtc_service = ServiceRtc(channel, uid)
    rtc_service.add_privilege(ServiceRtc.kPrivilegeJoinChannel, privilege_expire)
    rtc_service.add_privilege(ServiceRtc.kPrivilegePublishAudioStream, privilege_expire)
    rtc_service.add_privilege(ServiceRtc.kPrivilegePublishVideoStream, privilege_expire)
    rtc_service.add_privilege(ServiceRtc.kPrivilegePublishDataStream, privilege_expire)
    token.add_service(rtc_service)

    return token.build()


# --- v007 token pack helpers ---

def pack_uint16(x):
    return struct.pack('<H', int(x))


def pack_uint32(x):
    return struct.pack('<I', int(x))


def pack_string(string):
    if isinstance(string, str):
        string = string.encode('utf-8')
    return pack_uint16(len(string)) + string


def pack_map_uint32(m):
    return pack_uint16(len(m)) + b''.join([pack_uint16(k) + pack_uint32(v) for k, v in m.items()])


# --- v007 token classes ---

class Service:
    """Base class for v007 token services."""

    def __init__(self, service_type):
        self.__type = service_type
        self.__privileges = {}

    def __pack_type(self):
        return pack_uint16(self.__type)

    def __pack_privileges(self):
        privileges = OrderedDict(
            sorted(iter(self.__privileges.items()), key=lambda x: int(x[0])))
        return pack_map_uint32(privileges)

    def add_privilege(self, privilege, expire):
        self.__privileges[privilege] = expire

    def service_type(self):
        return self.__type

    def pack(self):
        return self.__pack_type() + self.__pack_privileges()


class ServiceRtc(Service):
    """RTC service for v007 token generation."""

    kServiceType = 1
    kPrivilegeJoinChannel = 1
    kPrivilegePublishAudioStream = 2
    kPrivilegePublishVideoStream = 3
    kPrivilegePublishDataStream = 4

    def __init__(self, channel_name='', uid=0):
        super(ServiceRtc, self).__init__(ServiceRtc.kServiceType)
        self.__channel_name = channel_name.encode('utf-8')
        self.__uid = b'' if uid == 0 else str(uid).encode('utf-8')

    def pack(self):
        return super(ServiceRtc, self).pack() + pack_string(self.__channel_name) + pack_string(self.__uid)


class AccessToken007:
    """Access token generator using v007 service-based architecture."""

    def __init__(self, app_id='', app_certificate='', issue_ts=0, expire=900):
        self.__app_id = app_id
        self.__app_cert = app_certificate
        self.__issue_ts = issue_ts if issue_ts != 0 else int(time.time())
        self.__expire = expire
        self.__salt = secrets.SystemRandom().randint(1, 99999999)
        self.__service = {}

    def __signing(self):
        signing = hmac.new(pack_uint32(self.__issue_ts), self.__app_cert, sha256).digest()
        signing = hmac.new(pack_uint32(self.__salt), signing, sha256).digest()
        return signing

    def __build_check(self):
        def is_uuid(data):
            if len(data) != 32:
                return False
            try:
                bytes.fromhex(data)
            except:
                return False
            return True

        if not is_uuid(self.__app_id) or not is_uuid(self.__app_cert):
            return False
        if not self.__service:
            return False
        return True

    def add_service(self, service):
        self.__service[service.service_type()] = service

    def build(self):
        if not self.__build_check():
            return ''

        self.__app_id = self.__app_id.encode('utf-8')
        self.__app_cert = self.__app_cert.encode('utf-8')
        signing = self.__signing()
        signing_info = pack_string(self.__app_id) + pack_uint32(self.__issue_ts) + pack_uint32(self.__expire) + \
                       pack_uint32(self.__salt) + pack_uint16(len(self.__service))

        for _, service in self.__service.items():
            signing_info += service.pack()

        signature = hmac.new(signing, signing_info, sha256).digest()

        return '007' + base64.b64encode(zlib.compress(pack_string(signature) + signing_info)).decode('utf-8')
