namespace yorpc {
  export interface Encoding {
    encode(obj: Object): Uint8Array
    decode(bytes: ArrayBuffer): Object
  }

  class _Json implements Encoding {
    public encode(obj: Object): Uint8Array {
      let bytes = JSON.stringify(obj);
      return str2bytes(bytes);
    }

    public decode(data: ArrayBuffer): Object {
      let text = bytes2str(data);
      return JSON.parse(text);
    }
  }
  export let Json = new _Json();

  export class FlatBuffer implements Encoding {
    constructor(fbBuilder: any) {
    }

    public encode(obj: Object): Uint8Array {
      throw new Error("Not Implements yet");
    }

    public decode(data: ArrayBuffer): Object {
      throw new Error("Not Implements yet");
    }
    private fbBuilder;
  }

}