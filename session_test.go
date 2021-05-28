package yorpc

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

type TestClient struct {
	ws      *websocket.Conn
	handler func(int, []byte)
}

func (this *TestClient) Start() {
	if this.ws == nil {
		panic(fmt.Errorf("Invalid websocket."))
	}

	msgType, msgBody, err := this.ws.ReadMessage()
	if err != nil {
		panic(err)
	}
	this.handler(msgType, msgBody)
}

var opts = Options{MaxConn: 1000, HeartBeat: 1}
var msgDatas = make(chan []byte)
var handlers = map[uint16]MsgHandler{
	101: func(session Session, msgData []byte) []byte {
		msgDatas <- msgData
		session.SendMsg(2000, []byte("resp"))
		return nil
	},

	102: func(session Session, msgData []byte) []byte {
		msgDatas <- msgData
		return msgData
	},
}
var isStarting = false

func startServerForTest(listen string, path string) {
	if isStarting {
		return
	}
	isStarting = true
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		session := NewRpcSession("1", "2", handlers)
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			panic(err)
		}
		session.Connect2(ws)
		session.Start(opts)
	})
	go http.ListenAndServe(listen, nil)
	time.Sleep(10 * time.Millisecond)
}

func newClient(url string) (*RpcSession, error) {
	ws, _, err := websocket.DefaultDialer.Dial(url, http.Header{})
	if err != nil {
		return nil, err
	}
	c := NewRpcSession("abc", "def", handlers)
	c.Connect2(ws)
	return c, nil
}

func TestRPC(t *testing.T) {
	assert := assert.New(t)
	// ch := make(chan int)

	addr := "127.0.0.1:55555"
	startServerForTest(addr, "/testcase1")

	// client
	url := fmt.Sprintf("ws://%s/testcase1", addr)
	c, err := newClient(url)
	go c.Start(opts)
	assert.NoError(err)
	{
		msgDatas_get := func() []byte {
			select {
			case v := <-msgDatas:
				return v
			case <-time.After(1 * time.Second):
				return nil
			}
		}

		c.SendMsg(101, []byte("hello yorpc1!!!"))
		assert.Equal("hello yorpc1!!!", string(msgDatas_get()))

		c.Call(102, []byte("hello yorpc2!!!"), func(respData []byte) {
			t.Logf("respData: %s", string(respData))
			msgDatas <- []byte("resp")
		})
		assert.Equal("hello yorpc2!!!", string(msgDatas_get()))
		assert.Equal("resp", string(msgDatas_get()))
	}
}

func TestKeepAlive(t *testing.T) {
	assert := assert.New(t)

	s := NewRpcSession("1", "2", nil)
	s.KeepAlive(time.Second)
	now := time.Now()
	assert.True(s.IsAlive(now))

	now = now.Add(time.Second)
	assert.False(s.IsAlive(now))
}

func TestStartStop(t *testing.T) {
	assert := assert.New(t)

	addr := "127.0.0.1:55555"
	startServerForTest(addr, "/testcase1")

	// client
	url := fmt.Sprintf("ws://%s/testcase1", addr)
	c, err := newClient(url)
	assert.NoError(err)
	c.Stop()

	ch := make(chan error)
	go func() {
		ch <- c.Start(opts)
	}()
	err = <-ch
	assert.Error(err)
}
