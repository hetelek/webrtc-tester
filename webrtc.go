package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/gorilla/websocket"
	webrtc "github.com/pion/webrtc/v3"
)

type WebrtcConn struct {
	offerer bool
	pc      *webrtc.PeerConnection
	ws      *websocket.Conn
}

func startWebrtc(signalAddress string, offerer bool) {
	// read config.json
	configFile, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	defer configFile.Close()
	byteValue, err := ioutil.ReadAll(configFile)
	if err != nil {
		panic(err)
	}
	var config webrtc.Configuration
	json.Unmarshal(byteValue, &config)
	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	conn, _, err := websocket.DefaultDialer.Dial(signalAddress, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	webrtcConn := WebrtcConn{
		offerer: offerer,
		pc:      pc,
		ws:      conn,
	}
	if !offerer {
		webrtcConn.pc.OnDataChannel(func(dc *webrtc.DataChannel) {
			log.Println("recieved data channel")
			dc.OnMessage(func(msg webrtc.DataChannelMessage) {
				log.Printf("received msg: %s", string(msg.Data))
				textMsg := "world"
				log.Printf("sending msg: %s\n", textMsg)
				dc.SendText(textMsg)
			})
		})
	}
	webrtcConn.pc.OnConnectionStateChange(func(pcs webrtc.PeerConnectionState) {
		log.Printf("connection state changed: %s\n", pcs.String())
		switch pcs {
		case webrtc.PeerConnectionStateClosed:
			fallthrough
		case webrtc.PeerConnectionStateDisconnected:
			fallthrough
		case webrtc.PeerConnectionStateFailed:
			os.Exit(0)
		case webrtc.PeerConnectionStateConnected:
			webrtcConn.ws.Close()
		}
	})
	webrtcConn.pc.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}
		jsonCandidate := i.ToJSON()
		log.Printf("sending candidate: %s", jsonCandidate.Candidate)
		webrtcConn.send(jsonCandidate)
	})
	for {
		_, message, err := webrtcConn.ws.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			webrtcConn.ws.Close()
			log.Println("websocket disconnected")
			break
		}
		webrtcConn.processMessage(message)
	}
	select {}
}

func (conn *WebrtcConn) send(obj interface{}) {
	payload, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	err = conn.ws.WriteMessage(websocket.BinaryMessage, payload)
	if err != nil {
		panic(err)
	}
}

func (conn *WebrtcConn) processMessage(payload []byte) {
	if bytes.Compare(payload, startPayload) == 0 {
		if conn.offerer {
			dc, err := conn.pc.CreateDataChannel("dc", nil)
			if err != nil {
				panic(err)
			}
			dc.OnOpen(func() {
				go func() {
					for {
						textMsg := "hello"
						log.Printf("sending msg: %s\n", textMsg)
						dc.SendText(textMsg)
						time.Sleep(5 * time.Second)
					}
				}()
			})
			dc.OnMessage(func(msg webrtc.DataChannelMessage) {
				log.Printf("received msg: %s\n", string(msg.Data))
			})

			offer, err := conn.pc.CreateOffer(nil)
			if err != nil {
				panic(err)
			}
			err = conn.pc.SetLocalDescription(offer)
			if err != nil {
				panic(err)
			}
			log.Println("sending offer")
			conn.send(offer)
		}
	} else {
		var (
			sdp       webrtc.SessionDescription
			candidate webrtc.ICECandidateInit
		)

		switch {
		case json.Unmarshal(payload, &sdp) == nil && sdp.SDP != "":
			if sdp.Type == webrtc.SDPTypeOffer && conn.offerer {
				panic("received offer as offerer. Ensure one client is answerer and one is offerer")
			} else if sdp.Type == webrtc.SDPTypeAnswer && !conn.offerer {
				panic("received answer as answerer. Ensure one client is answerer and one is offerer")
			}

			log.Println("received remote SDP")
			err := conn.pc.SetRemoteDescription(sdp)
			if err != nil {
				panic(err)
			}
			if !conn.offerer {
				answer, err := conn.pc.CreateAnswer(nil)
				if err != nil {
					panic(err)
				}
				err = conn.pc.SetLocalDescription(answer)
				if err != nil {
					panic(err)
				}
				log.Println("sending answer")
				conn.send(answer)
			}
		case json.Unmarshal(payload, &candidate) == nil && candidate.Candidate != "":
			if err := conn.pc.AddICECandidate(candidate); err != nil {
				panic(err)
			}
		default:
			panic("Unknown message")
		}
	}
}
