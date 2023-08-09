namespace yorpc {
  export type MsgHandler = (session: any, msgId: number, args: Uint8Array) => Uint8Array;
  export type MsgCallback = (args: Uint8Array) => void;
  export interface MsgSender {
    send(ArrayBuffer): void;
  }

  export class Session {
    constructor(sender: MsgSender, handler: MsgHandler | null) {
      this.callbacks = new Map<number, MsgCallback>();
      this.callSeqNum = 0;
      this.sender = sender;
      this.handler = handler;
    }

    public call(msgId: number, data: any, callback = null) {
      if (typeof data == 'string') {
        data = str2bytes(data);
      }
      let buf;
      if (isNull(callback)) {
        let s = short2bytes(msgId);
        buf = new Uint8Array(2 + data.length);
        buf[0] = s[0];
        buf[1] = s[1];
        buf.set(data, 2);

      } else {
        let callSeqId = this._genNextCallSeqId();
        let s = short2bytes(msgId);
        buf = new Uint8Array(1 + 2 + data.length);
        buf[0] = this._encodeCallFlag(1, callSeqId);
        buf[1] = s[0];
        buf[2] = s[1];
        buf.set(data, 3);
        this.callbacks[callSeqId] = callback;

      }
      this._send(buf.buffer);
    }

    public onMessage(msg) {
      if (typeof msg == 'string') {
        console.log(msg);

      } else if (msg instanceof Uint8Array) {
        console.debug("Uint8Array", msg);
        this._onMessageData(msg);

      } else {
        this._onMessageData(msg.data);

      }
    }

    public _onMessageData(msg: Uint8Array) {
      let callFlagDecoded = this._decodeCallFlag(msg[0]);
      let isReq = callFlagDecoded[0];
      let callSeqId = callFlagDecoded[1];
      let msgId = 0;
      let msgData: Uint8Array;
      if (callSeqId > 0) {
        if (isReq) {
          // request with callback
          let s = msg.slice(1, 3);
          msgId = bytes2short(s);
          msgData = msg.slice(3, msg.length);

        } else {
          // callback
          msgData = msg.slice(1, msg.length);
          let callback = this.callbacks[callSeqId];
          if (isNull(callback)) {
            console.log(`ERROR: callback was null. callSeqId=${callSeqId}`);
            return;
          }
          this.callbacks.delete(callSeqId);
          // this.callbacks[callSeqId] = null;
          callback(msgData);
          return;

        }
      } else {
        // request without callback
        let s = msg.slice(1, 3);
        msgId = bytes2short(s);
        msgData = msg.slice(3, msg.length);
      }
      if (this.handler == null) {
        console.log(`ERROR: handler was null. msgid=${msgId}`);
        return;
      }

      let rs;
      try {
        rs = this.handler(this, msgId, msgData);
      } catch (err) {
        console.log(err);
      } finally {
        if (callSeqId > 0) {
          this._returnMsg(callSeqId, rs);
        }
      }
    }

    _genNextCallSeqId(): number {
      this.callSeqNum++;
      this.callSeqNum = (this.callSeqNum % 128);
      if (this.callSeqNum == 0) {
        this.callSeqNum++;
      }
      return this.callSeqNum;
    }

    _encodeCallFlag(isReq, callSeqId): number {
      // TODO: assert typeo of isReq
      return (isReq << 7) + (callSeqId & 0x7f);
    }

    _decodeCallFlag(v): [boolean, number] {
      // TODO: assert typeo of isReq
      let isReq = (v >> 7) == 1;
      let callSeqId = (v & 0x7f);
      return [isReq, callSeqId];
    }

    public _returnMsg(callSeqId, data) {
      let buf = new Uint8Array(1 + data.length);
      buf[0] = this._encodeCallFlag(0, callSeqId);
      buf.set(data, 1);
      this._send(buf.buffer);
    }

    /**
     * 原始接口 
     * @param data 
     */
    public _send(data: ArrayBuffer): void {
      this.sender.send(data);
    }

    private handler: MsgHandler | null
    private callbacks: Map<number, MsgCallback>
    private callSeqNum = 0
    private sender: MsgSender
  }
}

// @See https://github.com/google/closure-library/blob/master/closure/goog/crypt/crypt.js#L117
function str2bytes(text: string): Uint8Array {
  let out = [];
  let p = 0;
  for (let i = 0; i < text.length; i++) {
    let c = text.charCodeAt(i);
    if (c < 128) {
      out[p++] = c;
    } else if (c < 2048) {
      out[p++] = (c >> 6) | 192;
      out[p++] = (c & 63) | 128;
    } else if (
      ((c & 0xFC00) == 0xD800) && (i + 1) < text.length &&
      ((text.charCodeAt(i + 1) & 0xFC00) == 0xDC00)) {
      // Surrogate Pair
      c = 0x10000 + ((c & 0x03FF) << 10) + (text.charCodeAt(++i) & 0x03FF);
      out[p++] = (c >> 18) | 240;
      out[p++] = ((c >> 12) & 63) | 128;
      out[p++] = ((c >> 6) & 63) | 128;
      out[p++] = (c & 63) | 128;
    } else {
      out[p++] = (c >> 12) | 224;
      out[p++] = ((c >> 6) & 63) | 128;
      out[p++] = (c & 63) | 128;
    }
  }
  return new Uint8Array(out);
}


function bytes2str(bytes) {
  // TODO(user): Use native implementations if/when available
  let out = [], pos = 0, c = 0;
  while (pos < bytes.length) {
    let c1 = bytes[pos++];
    if (c1 < 128) {
      out[c++] = String.fromCharCode(c1);
    } else if (c1 > 191 && c1 < 224) {
      let c2 = bytes[pos++];
      out[c++] = String.fromCharCode((c1 & 31) << 6 | c2 & 63);
    } else if (c1 > 239 && c1 < 365) {
      // Surrogate Pair
      let c2 = bytes[pos++];
      var c3 = bytes[pos++];
      let c4 = bytes[pos++];
      let u = ((c1 & 7) << 18 | (c2 & 63) << 12 | (c3 & 63) << 6 | c4 & 63) -
          0x10000;
      out[c++] = String.fromCharCode(0xD800 + (u >> 10));
      out[c++] = String.fromCharCode(0xDC00 + (u & 1023));
    } else {
      let c2 = bytes[pos++];
      let c3 = bytes[pos++];
      out[c++] =
          String.fromCharCode((c1 & 15) << 12 | (c2 & 63) << 6 | c3 & 63);
    }
  }
  return out.join('');
};

function short2bytes(v, bigEndian = false): Uint8Array {
  let arr = new Uint8Array(2);
  if (bigEndian) {
    arr[0] = (v & 0xff00) >> 8;
    arr[1] = (v & 0x00ff);
  } else {
    arr[0] = (v & 0x00ff);
    arr[1] = (v & 0xff00) >> 8;
  }
  return arr;
}

function bytes2short(arr, bigEndian = false): number {
  if (bigEndian) {
    return arr[1] | arr[0] << 8;
  }
  return arr[0] | arr[1] << 8;
}

function isNull(v): boolean {
  return v === undefined || v === null;
}