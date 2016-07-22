package main

import (
	"flag"
	"log"

	"github.com/bonnefoa/signaler/brokerlib"
	"github.com/bonnefoa/signaler/conf"
)

func main() {
	flag.Parse()
	log.SetFlags(0)
	conf.SetupZmq()
	brokerlib.LaunchBroker()
}
