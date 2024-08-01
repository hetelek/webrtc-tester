package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/stun"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	pendingConn *RelayConnection
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
		relayConn.other.other = relayConn
		pendingConn = nil
		go relayConn.conn.WriteMessage(websocket.BinaryMessage, startPayload)
		go relayConn.other.conn.WriteMessage(websocket.BinaryMessage, startPayload)
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

func getPublicIP(networkType string) string {
	c, err := stun.Dial(networkType, "stun.l.google.com:19302")
	if err != nil {
		return ""
	}
	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

	var publicIP string
	wg := sync.WaitGroup{}
	wg.Add(1)
	err = c.Do(message, func(res stun.Event) {
		defer wg.Done()
		if res.Error != nil {
			return
		}
		var xorAddr stun.XORMappedAddress
		if err := xorAddr.GetFrom(res.Message); err != nil {
			return
		}
		publicIP = xorAddr.IP.String()
	})
	if err != nil {
		return ""
	}
	wg.Wait()
	return publicIP
}

func printExampleCommands(port int) {
	const (
		signalServerFormat           = "ws://%s:%d/relay"
		exampleOffererCommandFormat  = "./%s -offerer -signal-address \"%s\""
		exampleAnswererCommandFormat = "./%s -answerer -signal-address \"%s\""
	)

	binaryName := filepath.Base(os.Args[0])
	publicIPv4 := getPublicIP("udp4")
	if len(publicIPv4) > 0 {
		address := fmt.Sprintf(signalServerFormat, publicIPv4, port)
		offerExample := fmt.Sprintf(exampleOffererCommandFormat, binaryName, address)
		answerExample := fmt.Sprintf(exampleAnswererCommandFormat, binaryName, address)
		fmt.Printf("example (ipv4):\n%s\n%s\n", offerExample, answerExample)
		fmt.Println()
	}

	publicIPv6 := getPublicIP("udp6")
	if len(publicIPv6) > 0 {
		publicIPv6 = fmt.Sprintf("[%s]", publicIPv6)
		address := fmt.Sprintf(signalServerFormat, publicIPv6, port)
		offerExample := fmt.Sprintf(exampleOffererCommandFormat, binaryName, address)
		answerExample := fmt.Sprintf(exampleAnswererCommandFormat, binaryName, address)
		fmt.Printf("example (ipv6):\n%s\n%s\n", offerExample, answerExample)
		fmt.Println()
	}
}

func startSignalServer(port int) {
	http.HandleFunc("/relay", relay)
	log.Printf("hosting signal server at :%d", port)
	go printExampleCommands(port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
