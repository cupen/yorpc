const DIR_Up = 0;
const DIR_Down = 1;
const DIR_Left = 2;
const DIR_Right = 3;

class Keys {
    constructor() {
        this.up = false;
        this.down = false;
        this.left = false;
        this.right = false;
    }
    isValid() {
        return this.up || this.down || this.left || this.right;
    }
}

class Position {
    constructor(x = 0, y = 0) {
        this.x = x;
        this.y = y;
    }
    set(x, y) {
        this.x = x;
        this.y = y;
    }
}

class Player {
    constructor(id, name) {
        this.name = name;
        this.id = id;
        this.avatar = null;
        this.pos = new Position();
        this.dirty = false;
        this.dir = DIR_Up;
        this.speed = 4;
        this.keys = new Keys();
        this.sprite = null;
    }
    bindAvatar(avatar, sprite) {
        this.avatar = avatar;
        this.sprite = sprite;
    }
    bindNetwork(conn) {
        this.conn = conn;
        return this.conn.start((event) => {
            let msg = JSON.parse(event.data);
            this.onMessage(msg.type, msg.data)
        })
    }
    onMessage(name, data) {
        switch (name) {
            case "start":
                this.onStart();
                break;
            case "join":
        }
    }
    move(x, y) {
        this.pos.set(x, y);
        this.dirty = true;
    }
    asSprite() {
        return this.sprite;
    }
    onKeyEvent(event, keyCode) {
        if (event === "keyup") {
            console.log(`key-up:  ${keyCode}`)
            switch (keyCode) {
                case 87:
                    this.keys.up = false;
                    break;
                case 83:
                    this.keys.down = false;
                    break;
                case 65:
                    this.keys.left = false;
                    break;
                case 68:
                    this.keys.right = false;
                    break;
            }
        } else if (event === "keydown") {
            console.log(`key-down: ${keyCode}`)
            switch (keyCode) {
                case 87:
                    this.keys.up = true;
                    break;
                case 83:
                    this.keys.down = true;
                    break;
                case 65:
                    this.keys.left = true;
                    break;
                case 68:
                    this.keys.right = true;
                    break;
            }
        }
    }
    onStart() {
        console.info("Game start");
    }
    onNetwork(conn) {
        conn.send("move", {
            "id": this.id,
            "x": this.pos.x,
            "y": this.pos.y
        });
    }
    onUpdate() {
        // console.log("onUpdate")
        if (this.keys.isValid()) {
            if (this.keys.up) {
                this.pos.y -= this.speed;
            }
            if (this.keys.down) {
                this.pos.y += this.speed;
            }
            if (this.keys.left) {
                this.pos.x -= this.speed;
            }
            if (this.keys.right) {
                this.pos.x += this.speed;
            }

            this.dirty = true
        }
        if (this.dirty) {
            let offsetX =  this.sprite.x - this.pos.x;
            let offsetY =  this.sprite.y - this.pos.y;
            if (offsetX < 0) {
                this.sprite.x += this.speed;
            } else if (offsetX > 0) {
                this.sprite.x -= this.speed;
            }
            if (offsetY < 0) {
                this.sprite.y += this.speed;
            } else if (offsetY > 0) {
                this.sprite.y -= this.speed;
            }
            // this.sprite.x = this.pos.x;
            // this.sprite.y = this.pos.y;
            // this.dirty = true
            // console.log(`pos-> x: ${this.sprite.x} y: ${this.sprite.y}`)
        }
    }
}

class Connection {
    constructor(url) {
        this.url = url;
        this.ws = null;
        // this.connect();
    }

    _connect(callback, retry = 3) {
        return new Promise((resolve, reject) => {
            this.ws = new WebSocket(this.url);
            this.ws.addEventListener("open", (event) => {
                console.info("Connection opened");
                this.onOpen();
                resolve();
            });
            this.ws.addEventListener("message", callback);
            this.ws.addEventListener("close", (event) => {
                console.info("Connection closed");
                this.ws = null;
                if (retry <= 0) {
                    reject();
                    return;
                }
                setTimeout(() => {
                    this._connect(callback, retry = retry - 1);
                }, 1000);
            });
        })
    }
    onOpen() {
        this.timer = setInterval(() => {
            // console.log(`keepalive: ${this.ws}`);
            this.keepAlive();
        }, 3 * 1000);
    }

    start(callback) {
        return this._connect(callback)
    }

    send(name, data) {
        this._send(name, data);
    }

    keepAlive() {
        this._send("ping", {});
    }

    stop() {
        if (this.timer) {
            clearInterval(this.timer);
            this.timer = null;
        }
        if (this.ws) {
            try {
                this.ws.close();
            } catch (error) {
                console.error(error);
            }
            this.ws = null;
        }
    }

    _send(name, data) {
        if (name !== "ping") {
            console.info(`send: ${name} -> ${JSON.stringify(data)}`)
        }
        this.ws.send(JSON.stringify({
            "type": name,
            "data": data,
        }));
    }
}