package chat

import (
	"flag"
	"log"
)

func main() {
	var port = flag.Int("port", 0, "tcp port")
	var client = flag.Bool("client", false, "run client")
	var server = flag.Bool("server", false, "run client")

	if *server {
		runClient(*port)
	}
	if *client {
		runClient(*port)
	}
	log.Printf("--server or --client required"
	return
}


func runServer() {

}

func runClient() {

}