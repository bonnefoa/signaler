package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var numConnections = flag.Int("numConnections", 2000,
	"Number of test connections")

var testAddr = "localhost:10443"

func TestMain(m *testing.M) {
	if !flag.Parsed() {
		flag.Parse()
		log.SetFlags(0)
		if !testing.Verbose() {
			log.SetOutput(ioutil.Discard)
		}
		go launchWebsocketServer(&testAddr)
	}
	atomic.StoreUint64(&numMsgSend, 0)
	os.Exit(m.Run())
}

func connect(t *testing.T, addr *string) *websocket.Conn {
	var err error
	for i := 0; i < 5; i++ {
		time.Sleep(200 * time.Millisecond)
		u := url.URL{Scheme: "ws", Host: *addr, Path: "/signaler"}
		d := websocket.Dialer{Subprotocols: []string{"signaler"}}
		c, _, err := d.Dial(u.String(), nil)
		if err == nil {
			return c
		}
	}
	t.Fatal("Could not connect", err)
	return nil
}

func sendHandshake(t *testing.T, cnd candidate, conn *websocket.Conn) {
	err := conn.WriteJSON(cnd)
	if err != nil {
		t.Fatal("write:", err)
	}
}

func sendMessage(t *testing.T, dst string, conn *websocket.Conn,
	key string) {
	msg := make(map[string]string)
	msg[key] = "toto"
	msg["dest"] = dst
	err := conn.WriteJSON(msg)
	if err != nil {
		t.Fatal("write:", err)
	}
}

func closeConn(t *testing.T, conn *websocket.Conn) {
	err := conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		t.Fatal("write close:", err)
	}
	conn.Close()
}

func recvMsg(t *testing.T, conn *websocket.Conn) map[string][]byte {
	res := make(map[string][]byte)
	err := conn.ReadJSON(&res)
	if err, ok := err.(net.Error); ok && err.Timeout() {
		return nil
	}
	if err != nil {
		t.Fatal("read:", err)
	}
	return res
}

func getNumCandidates() int {
	candidatesMutex.Lock()
	num := len(candidates)
	candidatesMutex.Unlock()
	return num
}

func waitCandidateNumber(t *testing.T, expectedNumber int) {
	for i := 0; i < 10 || getNumCandidates() != expectedNumber; i++ {
		time.Sleep(500 * time.Millisecond)
		fmt.Printf("Expected %d, got %d\n", expectedNumber, getNumCandidates())
	}
	if getNumCandidates() != expectedNumber {
		t.Fatalf("Expected %d candidates, got %d",
			expectedNumber, getNumCandidates())
	}
}

func TestSimpleHandshake(t *testing.T) {
	conn := connect(t, &testAddr)
	cnd := candidate{ID: "first"}
	sendHandshake(t, cnd, conn)
	waitCandidateNumber(t, 1)
	conn.Close() // Test brutal close
	waitCandidateNumber(t, 0)
}

func TestSimpleCommunication(t *testing.T) {
	firstConn := connect(t, &testAddr)
	sndConn := connect(t, &testAddr)

	firstID := "firstClient"
	sndID := "secondClient"
	firstCnd := candidate{ID: firstID}
	sndCnd := candidate{ID: sndID}

	sendHandshake(t, firstCnd, firstConn)
	sendHandshake(t, sndCnd, sndConn)

	waitCandidateNumber(t, 2)
	sendMessage(t, sndID, firstConn, "candidate")
	firstMsg := recvMsg(t, sndConn)
	if _, ok := firstMsg["candidate"]; !ok {
		t.Fatalf("Expected candidate in map %s", firstMsg)
	}

	sendMessage(t, sndID, firstConn, "sdp")
	sndMsg := recvMsg(t, sndConn)
	if _, ok := sndMsg["sdp"]; !ok {
		t.Fatalf("Expected sdp in map %s", sndMsg)
	}

	sendMessage(t, "unknown", firstConn, "sdp")
	currentMsg := atomic.LoadUint64(&numMsgSend)
	if currentMsg != 2 {
		t.Fatalf("Only 2 messages should have been send, got %d", currentMsg)
	}

	closeConn(t, sndConn)
	waitCandidateNumber(t, 1)
	closeConn(t, firstConn)
	waitCandidateNumber(t, 0)
}

func openConnection(conn **websocket.Conn, id string, t *testing.T) {
	*conn = connect(t, &testAddr)
	sendHandshake(t, candidate{ID: id}, *conn)
}

func openWorker(t *testing.T, connections []*websocket.Conn, jobs <-chan int) {
	for i := range jobs {
		openConnection(&connections[i], strconv.Itoa(i), t)
	}
}

func TestLotsOfConnections(t *testing.T) {
	connections := make([]*websocket.Conn, *numConnections, *numConnections)

	jobs := make(chan int, *numConnections)

	for i := 0; i < 800; i++ {
		go openWorker(t, connections, jobs)
	}

	for i := 0; i < *numConnections; i++ {
		jobs <- i
	}
	close(jobs)
	waitCandidateNumber(t, *numConnections)
	for i := 0; i < *numConnections; i++ {
		closeConn(t, connections[i])
	}
	waitCandidateNumber(t, 0)
}
