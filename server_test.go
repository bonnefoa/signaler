package main

import (
	"crypto/tls"
	"log"
	"net"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var firstLaunch = true
var testAddr = "localhost:10443"

func TestMain(m *testing.M) {
	if firstLaunch {
		log.SetFlags(0)
		go launchWebsocketServer(&testAddr)
		firstLaunch = false
	}
	os.Exit(m.Run())
}

func connect(t *testing.T, addr *string) *websocket.Conn {
	var err error
	for i := 0; i < 5; i++ {
		time.Sleep(200 * time.Millisecond)
		u := url.URL{Scheme: "wss", Host: *addr, Path: "/signaler"}
		d := websocket.Dialer{Subprotocols: []string{"signaler"}}
		d.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
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

func waitCandidateNumber(t *testing.T, expectedNumber int) {
	for i := 0; i < 10 || len(candidates) != expectedNumber; i++ {
		time.Sleep(100 * time.Millisecond)
	}
	if len(candidates) != expectedNumber {
		t.Fatalf("Expected %d candidates, got %d",
			expectedNumber, len(candidates))
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
	sndConn.SetReadDeadline(time.Now().Add(time.Second))
	thirdMsg := recvMsg(t, sndConn)
	if thirdMsg != nil {
		t.Fatalf("Should have timeouted on third msg, got %s", thirdMsg)
	}

	closeConn(t, sndConn)
	waitCandidateNumber(t, 1)
	closeConn(t, firstConn)
	waitCandidateNumber(t, 0)
}
