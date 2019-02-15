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

func (this *TestClient) Stop() {

}

func TestRPC(t *testing.T) {
	assert := assert.New(t)

	msgCallSeqNums := make(chan uint8)
	msgDatas := make(chan []byte)

	handlers := map[uint16]MsgHandler{
		101: func(session Session, callSeqId uint8, msgData []byte) {
			t.Logf("101 callId:%d body:%s", callSeqId, string(msgData))
			msgCallSeqNums <- callSeqId
			msgDatas <- msgData
			session.SendMsg(2000, []byte("resp"))
		},

		102: func(session Session, callSeqId uint8, msgData []byte) {
			t.Logf("102 callId:%d body:%s", callSeqId, string(msgData))
			msgCallSeqNums <- callSeqId
			msgDatas <- msgData

			session.ReturnMsg(callSeqId, msgData)
		},
	}

	// ch := make(chan int)
	opts := Options{MaxConn: 1000, HeartBeat: 1}
	server := NewWebSocketServer(opts)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		session := NewRpcSession("1", handlers)
		server.ServeHttp(w, r, session)
	})

	addr := "127.0.0.1:55555"
	go http.ListenAndServe(addr, nil)

	// c, err := Connect("ws://" + addr + "/case2")
	c := NewRpcSession("", handlers)
	go c.StartWithUrl("ws://" + addr)

	msgCallSeqNums_get := func() uint8 {
		select {
		case v := <-msgCallSeqNums:
			return v
		case <-time.After(3 * time.Second):
			return 0
		}
	}

	msgDatas_get := func() []byte {
		select {
		case v := <-msgDatas:
			return v
		case <-time.After(3 * time.Second):
			return nil
		}
	}

	c.SendMsg(101, []byte("hello yorpc1!!!"))
	assert.Equal(uint8(0), msgCallSeqNums_get())
	assert.Equal("hello yorpc1!!!", string(msgDatas_get()))

	//
	ch := make(chan int)
	c.Call(102, []byte("hello yorpc2!!!"), func(respData []byte) {
		t.Logf("respData: %s", string(respData))
		ch <- 1
	})
	assert.Equal(uint8(1), msgCallSeqNums_get())
	assert.Equal("hello yorpc2!!!", string(msgDatas_get()))
	<-ch
}
