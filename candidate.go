package main

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type candidate struct {
	ID   string
	conn *websocket.Conn
}

func (cnd candidate) String() string {
	return fmt.Sprintf("Id:%s\r\n", cnd.ID)
}

type candidateMap map[string]candidate

var candidates = make(candidateMap)
