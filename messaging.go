package main

import (
	"fmt"
	"log"
	"sync/atomic"

	"github.com/bonnefoa/signaler/conf"
	"github.com/gorilla/websocket"
	zmq "github.com/pebbe/zmq4"
)

// Socket responsible for sending message to relay to broker
func createClientSocket(cnd *candidate) (*zmq.Socket, error) {
	clientSocket, err := zmq.NewSocket(zmq.REQ)
	clientSocket.SetIdentity(cnd.ID)
	if err != nil {
		log.Print("Could not create REQ socket", err)
		return nil, err
	}
	err = clientSocket.Connect(fmt.Sprintf("tcp://%s:%d", *conf.BrokerHost,
		*conf.FrontendPort))
	if err != nil {
		log.Printf("Could not connect to broker")
		return nil, err
	}
	return clientSocket, nil
}

// Socket responsible for relaying message from broker to websocket
func createWorkerSocket(cnd *candidate) (*zmq.Socket, error) {
	workerSocket, err := zmq.NewSocket(zmq.DEALER)
	workerSocket.SetIdentity(cnd.ID)
	if err != nil {
		log.Printf("Could not create DEALER socket")
		return nil, err
	}
	err = workerSocket.Connect(fmt.Sprintf("tcp://%s:%d", *conf.BrokerHost,
		*conf.BackendPort))
	if err != nil {
		log.Printf("Could not connect to broker")
		return nil, err
	}
	return workerSocket, nil
}

func workerLoop(conn *websocket.Conn, socket *zmq.Socket, cnd *candidate) {
	defer socket.Close()
	identity, err := socket.GetIdentity()
	if err != nil {
		log.Print("Could not get identity", err)
		return
	}
	log.Printf("Start worker loop for %s", identity)
	for {
		msg, err := socket.RecvBytes(0)
		log.Printf("[Zmq %s] received %q", identity, msg)
		if err != nil {
			log.Printf("Error when receiving message for candidate %s",
				cnd.String())
			return
		}
		if len(msg) == 1 && msg[0] == 'Q' {
			log.Printf("[Worker %s] Exiting", identity)
			return
		}
		log.Printf("[Websocket %s]: Send %q", cnd.String(), msg)
		conn.WriteMessage(websocket.TextMessage, msg)
		atomic.AddUint64(&numMsgSend, 1)
	}
}
