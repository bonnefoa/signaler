package conf

import "flag"

// BrokerHost flag for Broker host
var BrokerHost = flag.String("broker-host", "localhost", "Host of broker")

// FrontendPort flag for Frontend port
var FrontendPort = flag.Int("frontend-port", 5558, "Port used for frontend broker")

// BackendPort flag for Backend port
var BackendPort = flag.Int("backed-port", 5559, "Port used for backend broker")

// MaxZmqSockets flag for Maximum zmq socket
var MaxZmqSockets = flag.Int("max-zmq-sockets", 500000, "Maximum of zeromq sockets")
