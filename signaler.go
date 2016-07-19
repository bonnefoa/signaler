package main

import (
	"flag"
	"log"
)

var addr = flag.String("addr", ":10443", "Websocket server address")

func main() {
	flag.Parse()
	log.SetFlags(0)
	launchWebsocketServer(addr)
}
