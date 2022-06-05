package main

import (
	"flag"
)

func main() {
	signalServer := flag.Bool("signal", false, "if true, act as the signal server")
	flag.Parse()

	if *signalServer {
		startSignalServer()
	}
}
