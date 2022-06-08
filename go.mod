module github.com/hetelek/webrtc-tester

go 1.16

require (
	github.com/gorilla/websocket v1.5.0
	github.com/pion/stun v0.3.5
	github.com/pion/webrtc/v3 v3.1.41
)

replace (
	github.com/pion/ice/v2 v2.2.6 => github.com/hetelek/ice/v2 v2.3.2
	github.com/pion/turn/v2 v2.0.8 => github.com/hetelek/turn/v2 v2.1.0
)
