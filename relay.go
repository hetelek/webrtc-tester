package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type RelayConnection struct {
	conn   *websocket.Conn
	other  *RelayConnection
	closed bool
	lock   sync.Mutex
}

func (rc *RelayConnection) Close() {
	rc.lock.Lock()
	defer rc.lock.Unlock()
	if !rc.closed {
		rc.conn.Close()
		rc.closed = true
		if rc.other != nil {
			rc.other.Close()
		}
	}
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	pendingConn *RelayConnection
)

func relay(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	relayConn := &RelayConnection{
		conn: c,
	}
	if pendingConn != nil && !pendingConn.closed {
		log.Println("starting relay")
		relayConn.other = pendingConn
		pendingConn.other = relayConn
		pendingConn = nil
	} else {
		log.Println("client connected, waiting for peer")
		pendingConn = relayConn
	}

	startPump := func(read *RelayConnection) {
		defer read.Close()
		for {
			mt, message, err := read.conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}

			if read.other == nil {
				continue
			}

			err = read.other.conn.WriteMessage(mt, message)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}
	go startPump(relayConn)
}

func startSignalServer() {
	http.HandleFunc("/relay", relay)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
