package main

import (
	"flag"
	"log"
)

const (
	signalAddressExample = "ws://SIGNAL_ADDDRESS:8080/relay"
)

func main() {
	signalMode := flag.Bool("signal", false, "if true, act as the signal server")
	offerMode := flag.Bool("offerer", false, "if true, act as the offerer")
	answerMode := flag.Bool("answerer", false, "if true, act as the answerer")
	signalAddress := flag.String("signal-address", signalAddressExample, "the URL of the signal server")
	flag.Parse()

	checks := []*bool{signalMode, offerMode, answerMode}
	found := 0
	for _, c := range checks {
		if *c {
			found += 1
		}
	}
	if found != 1 {
		log.Fatalln("specify one of: signal, offerer, answerer")
	}
	if (*offerMode || *answerMode) && (len(*signalAddress) < 1 || *signalAddress == signalAddressExample) {
		log.Fatalln("specify a signal-address")
	}

	if *signalMode {
		startSignalServer()
	} else if *offerMode {
	} else if *answerMode {
	}
}
