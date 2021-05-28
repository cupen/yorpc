package chat

import (
	"flag"
	"log"
	"net/http"

	"github.com/cupen/yorpc"
)

func main() {
	var addr = flag.String("address", "127.0.0.1:7788", "network address")
	var client = flag.Bool("client", false, "run client")
	var server = flag.Bool("server", false, "run client")
	if *server {
		runClient(*addr)
	}
	if *client {
		runClient(*addr)
	}
	log.Printf("--server or --client required")
	return
}

func runServer(address string) {
	opts := &yorpc.Options{}
	http.HandleFunc("/", func(w ResponseWriter, r *Request) {
		s := yorpc.NewServer(opts)
		err := s.Start(r, w)
		if err != nil {
			log.Printf("server start: %v", err)
		}
	})
	http.ListenAndServe(address, nil)
}

func runClient(address string) {
	c := yorpc.NewClient(address)
	c.Call(1, []byte("hello"))
}
