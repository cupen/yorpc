namespace yorpc {
  type JsonHandler = (obj: any) => any
  type FlatBufferHandler = (obj:any) => any

  export class HandlerHub extends Map<number, MsgHandler> {
    /**
     * @rpc.bind(101)
     * function(session, data: Uint8Array): Uint8Array {
     *    .....
     * }
     * @param msgId 
     */
    public bind(msgId: number, encoding: Encoding = Json) {
      return function (target, propertyKey: string, descriptor: PropertyDescriptor) {
        console.log("bind2(): called");
        console.log(target);
        console.log(propertyKey);
        console.log(descriptor);

        let f = descriptor.value;
        descriptor.value = function(bytes: ArrayBuffer): any {
          let obj = encoding.decode(bytes)
          return f(obj)
        }
      }
    }

    public bind_json(msgId: number, encoding: Encoding = Json) {
      return function (f: Function): Function {
        this[msgId] = f;
        return function () {
          f();
        }
      };
    }

    public bind_flatbuffer(msgId: number, encoding: Encoding) {
      return function (f: Function): Function {
        this[msgId] = f;
        return function () {
          f();
        }
      };
    }
  }
}