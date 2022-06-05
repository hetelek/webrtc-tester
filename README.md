## Webrtc Tester

Webrtc tester is a simple tool that allows you to connect 2 webrtc clients and send messages over a data channel.

## Usage
1. Start signaling server
```
./webrtc_tester -signal
```

2. Start offerer
```
./webrtc_tester -offerer -signal-address ws://SIGNAL_SERVER:8080/relay
```

3. Start answerer
```
./webrtc_tester -offerer -signal-address ws://SIGNAL_SERVER:8080/relay
```

4. Machines 2 and 3 will attempt to connect

## Compiling
Standard:
```
go build -o webrtc_tester main.go payloads.go relay.go webrtc.go
```


Or you can cross-compile, for example:
```
GOOS=linux GOARCH=amd64 go build -o webrtc_tester main.go payloads.go relay.go webrtc.go
```

## Config.json
`config.json` is passed directly to the webrtc library. Configure as needed.
