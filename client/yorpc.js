class Client {
    constructor(ws) {
        this.conn = ws
    }

    call(msgId, bytes, timeout=5) {
        this.conn.send(bytes)
    }

    send(msgId, bytes) {
        this.conn.send(bytes)
    }
}

module.exports = {
    Client: Client
}


if (require.main === module) {
    const WebSocket = require("ws")
    var ws = new WebSocket("ws://127.0.0.1:55555/case2") 
    var c = new Client(ws)

    ws.on("open", function(){
        ws.send("hello")
    })
    ws.onmessage = function(event) {
        console.log(event)
    }
}


