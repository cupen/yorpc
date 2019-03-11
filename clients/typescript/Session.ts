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
  var out = [], pos = 0, c = 0;
  while (pos < bytes.length) {
    var c1 = bytes[pos++];
    if (c1 < 128) {
      out[c++] = String.fromCharCode(c1);
    } else if (c1 > 191 && c1 < 224) {
      var c2 = bytes[pos++];
      out[c++] = String.fromCharCode((c1 & 31) << 6 | c2 & 63);
    } else if (c1 > 239 && c1 < 365) {
      // Surrogate Pair
      var c2 = bytes[pos++];
      var c3 = bytes[pos++];
      var c4 = bytes[pos++];
      var u = ((c1 & 7) << 18 | (c2 & 63) << 12 | (c3 & 63) << 6 | c4 & 63) -
          0x10000;
      out[c++] = String.fromCharCode(0xD800 + (u >> 10));
      out[c++] = String.fromCharCode(0xDC00 + (u & 1023));
    } else {
      var c2 = bytes[pos++];
      var c3 = bytes[pos++];
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

namespace yorpc {
  export type MsgHandler = (session: any, args: Uint8Array) => Uint8Array;
  export type MsgCallback = (args: Uint8Array) => void;
  export interface MsgSender {
    send(ArrayBuffer): void;
  }

  export class Session {
    constructor(sender: MsgSender, handlers: HandlerHub = null) {
      this.callbacks = new Map<number, MsgCallback>();
      this.callSeqNum = 0;
      this.sender = sender;
      this.handlers = handlers;
    }

    getNextCallSeqId(): number {
      this.callSeqNum++;
      this.callSeqNum = (this.callSeqNum % 128);
      if (this.callSeqNum == 0) {
        this.callSeqNum++;
      }
      return this.callSeqNum;
    }

    encodeCallFlag(isReq, callSeqId): number {
      // TODO: assert typeo of isReq
      return (isReq << 7) + (callSeqId & 0x7f);
    }

    decodeCallFlag(v): [boolean, number] {
      // TODO: assert typeo of isReq
      let isReq = (v >> 7) == 1;
      let callSeqId = (v & 0x7f);
      return [isReq, callSeqId];
    }

    public onMessage(msg) {
      if (typeof msg == 'string') {
        console.log(msg);

      } else if (msg instanceof Uint8Array) {
        console.log("Uint8Array");
        console.log(msg);
        this.onMessageData(msg);

      } else {
        this.onMessageData(msg.data);

      }
    }

    public onMessageData(msg: Uint8Array) {
      let callFlagDecoded = this.decodeCallFlag(msg[0]);
      let isReq = callFlagDecoded[0];
      let callSeqId = callFlagDecoded[1];
      let msgId = 0;
      let msgData = null;
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
          this.callbacks[callSeqId] = null;
          callback(msgData);
          return;

        }
      } else {
        // request without callback
        let s = msg.slice(1, 3);
        msgId = bytes2short(s);
        msgData = msg.slice(3, msg.length);
      }
      let h = this.handlers.get(msgId);
      if (isNull(h)) {
        console.log(`ERROR: handler was null. msgid=${msgId}`);
        return;
      }

      let rs = null;
      try {
        rs = h(this, msgData);
      } catch (err) {
        console.log(err);
      } finally {
        if (callSeqId > 0) {
          this.returnMsg(callSeqId, rs);
        }
      }
    }

    public sendMsg(msgId: number, data: any, callback = null) {
      if (typeof data == 'string') {
        data = str2bytes(data);
      }
      console.log(callback);
      let buf = null;
      if (isNull(callback)) {
        let s = short2bytes(msgId);
        buf = new Uint8Array(2 + data.length);
        buf[0] = s[0];
        buf[1] = s[1];
        buf.set(data, 2);

      } else {
        let callSeqId = this.getNextCallSeqId();
        let s = short2bytes(msgId);
        buf = new Uint8Array(1 + 2 + data.length);
        buf[0] = this.encodeCallFlag(1, callSeqId);
        buf[1] = s[0];
        buf[2] = s[1];
        buf.set(data, 3);
        this.callbacks[callSeqId] = callback;

      }
      this.send(buf.buffer);
    }

    public returnMsg(callSeqId, data) {
      let buf = new Uint8Array(1 + data.length);
      buf[0] = this.encodeCallFlag(0, callSeqId);
      buf.set(data, 1);
      this.send(buf.buffer);
    }

    /**
     * @rpc.bind(101)
     * function(session, data: Uint8Array): Uint8Array {
     *    .....
     * }
     * @param msgId 
     */
    public bind(msgId: number) {
      return function (f: Function): Function {
        this.handlers[msgId] = f;
        return f;
      };
    }

    public send(data: ArrayBuffer): void {
      this.sender.send(data);
    }

    private handlers: HandlerHub
    private callbacks: Map<number, MsgCallback>
    private callSeqNum = 0
    private sender: MsgSender
  }

  interface Peer {

  }
}