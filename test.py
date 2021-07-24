import base64
import json
import time
from ctypes import *

import asyncio
from pyrogram import Client
from pyrogram.raw.functions.channels import GetFullChannel
from pyrogram.raw.functions.phone import JoinGroupCall
from pyrogram.raw.types import Updates, DataJSON


class GoString(Structure):
    _fields_ = [("p", c_char_p), ("n", c_longlong)]


def c_go_string(p: str):
    return GoString(str.encode(p), len(p))


lib = cdll.LoadLibrary("./dist/ntgcalls.so")
app = Client(
    'test',
    api_id=2799555,
    api_hash='47d66bbf0939d0ddf32caf8bad590ed7'
)
app.start()

chat_id = -1001393716813
lib.joinVoiceCall.argtypes = [c_int64, GoString]
lib.joinVoiceCall.restype = c_bool
lib.waitRequestJoin.argtypes = [c_int64]
lib.waitRequestJoin.restype = c_char_p
lib.sendResponseCall.argtypes = [c_int64, GoString]
lib.sendResponseCall.restype = c_bool
lib.initClient()
res = lib.joinVoiceCall(chat_id, c_go_string(""))
if res:
    async def wait_response():
        params = json.loads(lib.waitRequestJoin(chat_id).decode("utf-8"))
        request_call = {
            'ufrag': params['ufrag'],
            'pwd': params['pwd'],
            'fingerprints': [{
                'hash': params['hash'],
                'setup': params['setup'],
                'fingerprint': params['fingerprint'],
            }],
            'ssrc': params['source'],
        }
        chat = await app.resolve_peer(chat_id)
        full_chat = (
            await app.send(
                GetFullChannel(channel=chat),
            )
        ).full_chat.call
        if len(params['invite_hash']) == 0:
            params['invite_hash'] = None
        result: Updates = await app.send(
            JoinGroupCall(
                call=full_chat,
                params=DataJSON(data=json.dumps(request_call)),
                muted=False,
                join_as=await app.resolve_peer(
                    (await app.get_me())['id'],
                ),
                invite_hash=params['invite_hash'],
            ),
        )
        transport = json.loads(result.updates[0].call.params.data)[
            'transport'
        ]
        """file_path = 'bdescription.txt'
        with open(file_path, 'r') as f:
            result = f.read()
        transport = json.loads(base64.b64decode(result))
        print(transport)"""
        lib.sendResponseCall(chat_id, c_go_string(json.dumps({
            'ufrag': transport['ufrag'],
            'pwd': transport['pwd'],
            'fingerprints': transport['fingerprints'],
            'candidates': transport['candidates'],
        })))
    asyncio.get_event_loop().run_until_complete(wait_response())
    time.sleep(1000)  # This allow to keep alive the Go service
else:
    print("Internal Error")
