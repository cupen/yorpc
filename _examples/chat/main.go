package main

import (
	"chat/chatroom"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cupen/yorpc"
	"github.com/cupen/yorpc/connection/websocket"
)

var (
	server   = flag.String("server", "127.0.0.1:7788", "run server on a network address")
	client   = flag.Bool("client", false, "run client")
	nickName = flag.String("nickname", "", "your nickname(only for client)")
)

func main() {
	flag.Parse()
	if *client {
		if *nickName == "" {
			log.Printf("missing --nickname")
			return
		}
		runClient(*server, *nickName)
		return
	}
	runServer(*server)
	return
}

func runServer(address string) {
	opts := &yorpc.Options{}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.NewServer(r, w)
		if err != nil {
			panic(err)
		}
		id := r.URL.Query().Get("nickname")
		if id == "" {
			panic(fmt.Errorf("missing nickname"))
		}
		p := chatroom.NewChatroomSession(chatroom.NewPlayerImpl(id))
		s := yorpc.NewSession(conn, p, opts)
		if err := s.Run(); err != nil {
			log.Printf("server stopped: %v", err)
		}
	})
	log.Printf("server running: %s", address)
	http.ListenAndServe(address, nil)
}

func runClient(addr string, nickName string) {
	url := fmt.Sprintf("ws://%s?nickname=%s", addr, nickName)
	conn, err := websocket.NewClient(url, 3*time.Second)
	if err != nil {
		panic(err)
	}
	c := yorpc.NewClient(conn, 3*time.Second)
	// defer c.Stop()
	i := 0
	for {
		ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
		i++
		msg := chatroom.Message{
			UserID: "userId",
			Text:   fmt.Sprintf("%d 你好啊,我是 %s", i, nickName),
			Int64:  int64(i),
		}
		data, err := msg.Marshal()
		if err != nil {
			panic(err)
		}
		resp, err := c.Call(ctx, 1001, data)
		if err != nil {
			log.Panicf("call failed: %v  resp=%v", err, resp)
		}
		time.Sleep(1 * time.Second)
	}
}
