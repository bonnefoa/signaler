package main

import (
	"flag"
	"log"
)

var websocketAddr = flag.String("addr", ":10443", "Websocket server address")

func main() {
	flag.Parse()
	log.SetFlags(0)
	go launchBroker()
	launchWebsocketServer(addr)
}
