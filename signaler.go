package main

import (
	"flag"
	"log"
)

var addr = flag.String("addr", ":10443", "Websocket server address")
var brokerHost = flag.String("broker-host", "localhost", "Host of broker")
var frontendPort = flag.Int("frontend-port", 5558, "Port used for frontend broker")
var backendPort = flag.Int("backed-port", 5559, "Port used for backend broker")
var maxZmqSockets = flag.Int("max-zmq-sockets", 500000, "Maximum of zeromq sockets")

func main() {
	flag.Parse()
	log.SetFlags(0)
	go launchBroker()
	launchWebsocketServer(addr)
}
