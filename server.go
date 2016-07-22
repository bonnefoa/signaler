package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"
	zmq "github.com/pebbe/zmq4"
)

var openedConnections int32

var upgrader = websocket.Upgrader{
	Subprotocols: []string{"signaler"},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func cleanConnection(cnd *candidate, clientSocket *zmq.Socket,
	workerSocket *zmq.Socket) {
	log.Printf("Cleaning connection %s", cnd.ID)
	quitMsg := make([]string, 2, 2)
	quitMsg[1] = string("Q")
	quitMsg[0] = cnd.ID
	_, err := clientSocket.SendMessage(quitMsg)
	if err != nil {
		log.Printf("Could not send quit message to %s", cnd.String())
	}
	clientSocket.Close()
	atomic.AddInt32(&openedConnections, -1)
}

func listenLoop(conn *websocket.Conn, cnd *candidate) error {
	clientSocket, err := createClientSocket(cnd)
	if err != nil {
		return err
	}
	workerSocket, err := createWorkerSocket(cnd)
	if err != nil {
		return err
	}
	go workerLoop(conn, workerSocket, cnd)
	defer cleanConnection(cnd, clientSocket, workerSocket)
	atomic.AddInt32(&openedConnections, 1)

	log.Print("Start listening loop", cnd.String())
	for {
		messageType, msg, err := conn.ReadMessage()
		log.Printf("[Websocket %s]: Received %q", cnd.String(), msg)
		if err != nil {
			return err
		}
		if messageType == websocket.CloseMessage {
			log.Print("Got a close message")
			return nil
		}
		var res map[string]string
		json.Unmarshal(msg, &res)
		if _, ok := res["dest"]; !ok {
			return fmt.Errorf("No dest in message")
		}
		dstID := res["dest"]

		req := make([]string, 2, 2)
		req[1] = string(msg)
		req[0] = dstID

		_, err = clientSocket.SendMessage(req)
		if err != nil {
			log.Printf("Error sending message to broker")
			return err
		}
		data, err := clientSocket.RecvMessage(0)
		if err != nil {
			log.Printf("Error receiving message to broker")
			return err
		}
		log.Printf("[Zmq %s] Received %q", cnd.ID, data)
	}
}

func handshake(conn *websocket.Conn) (*candidate, error) {
	messageType, candidateMarshalled, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	if messageType != websocket.TextMessage {
		return nil, fmt.Errorf("Message type should be text, got %d",
			messageType)
	}
	cnd := candidate{conn: conn}
	json.Unmarshal(candidateMarshalled, &cnd)
	if cnd.ID == "" {
		return nil, fmt.Errorf("Illegal handshake message %s",
			string(candidateMarshalled))
	}
	return &cnd, nil
}

func signalerHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrader:", err)
		return
	}
	defer conn.Close()
	cnd, err := handshake(conn)
	if err != nil {
		log.Print("handshake:", err)
		conn.Close()
		return
	}
	err = listenLoop(conn, cnd)
	if err != nil {
		log.Print("listenLoop:", err)
		return
	}
}

func launchWebsocketServer(addr *string) {
	http.HandleFunc("/signaler", signalerHandler)
	log.Printf("Listening on %s", *addr)
	err := http.ListenAndServe(*addr, nil)
	log.Fatal(err)
}
