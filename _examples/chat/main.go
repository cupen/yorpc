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
)

func main() {
	var addr = flag.String("address", "127.0.0.1:7788", "network address")
	var client = flag.Bool("client", false, "run client")
	var server = flag.Bool("server", false, "run server")
	var nickName = flag.String("nickname", "bot-1", "your nickname(only for client)")
	flag.Parse()
	if *server {
		runServer(*addr)
		return
	}
	if *client {
		runClient(*addr, *nickName)
		return
	}
	log.Printf("--server or --client required")
	return
}

func runServer(address string) {
	opts := &yorpc.Options{}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := yorpc.NewWebsocketByHTTP(r, w)
		if err != nil {
			panic(err)
		}
		id := r.URL.Query().Get("nickname")
		if id == "" {
			panic(fmt.Errorf("missing nickname"))
		}
		p := chatroom.NewPlayerSession(chatroom.NewPlayerImpl(id))
		s := yorpc.NewServer(conn, p, opts)
		if err := s.Run(); err != nil {
			log.Printf("server stopped: %v", err)
		}
	})
	log.Printf("server running: %s", address)
	http.ListenAndServe(address, nil)
}

func runClient(address string, nickName string) {
	url := fmt.Sprintf("ws://%s?nickname=%s", address, nickName)
	c, err := yorpc.NewClient(url)
	if err != nil {
		panic(err)
	}
	c.Start()
	// defer c.Stop()
	i := 0
	for {
		ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
		i++
		resp, err := c.Call(ctx, 1, []byte(fmt.Sprintf("hello! I'm %s. %d", nickName, i)))
		if err != nil {
			log.Panicf("call failed: %v  resp=%v", err, resp)
		}
		time.Sleep(1 * time.Second)
	}
}
