package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

var numMsgSend int64

var upgrader = websocket.Upgrader{
	Subprotocols: []string{"signaler"},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func listenLoop(conn *websocket.Conn, cnd *candidate) error {
	log.Print("Start listening loop", cnd)
	for {
		messageType, msg, err := conn.ReadMessage()
		log.Printf("Received message %s, for %s", string(msg), cnd.String())
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
		if _, ok := candidates[dstID]; !ok {
			log.Printf("Unknown dest: %s", dstID)
			continue
		}
		candidates[dstID].conn.WriteMessage(messageType, msg)
		atomic.AddInt64(&numMsgSend, 1)
		log.Printf("Sending msg %s to %s", msg, dstID)
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
	log.Printf("Add candidate %s", cnd.ID)
	candidates[cnd.ID] = cnd
	return &cnd, nil
}

func cleanupConnection(conn *websocket.Conn, ID string) {
	log.Printf("Cleaning connection %s", ID)
	delete(candidates, ID)
	conn.Close()
}

func signalerEndpoint(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrader:", err)
		return
	}
	cnd, err := handshake(conn)
	if err != nil {
		log.Print("handshake:", err)
		conn.Close()
		return
	}
	defer cleanupConnection(conn, cnd.ID)
	err = listenLoop(conn, cnd)
	if err != nil {
		log.Print("listenLoop:", err)
		return
	}
}

func launchWebsocketServer(addr *string) {
	http.HandleFunc("/signaler", signalerEndpoint)
	log.Printf("Listening on %s", *addr)
	err := http.ListenAndServeTLS(*addr, "cert.pem", "key.pem", nil)
	log.Fatal(err)
}
