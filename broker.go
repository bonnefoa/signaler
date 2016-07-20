package main

import (
	"log"

	zmq "github.com/pebbe/zmq3"
)

func brokerListen(frontend *zmq.Socket, backend *zmq.Socket) error {
	defer frontend.Close()
	defer backend.Close()
	for {
		msg, err := frontend.RecvMessage(0)
		if err != nil {
			log.Print("Error when receiving message:", err)
			return err
		}
		log.Printf("Broker received %q", msg)
		sender := msg[0]
		dst := msg[2]
		data := msg[3]

		workerMsg := make([]string, 2, 2)
		workerMsg[0] = dst
		workerMsg[1] = data
		log.Printf("Send worker msg %q", workerMsg)
		_, err = backend.SendMessage(workerMsg)
		if err != nil {
			log.Print("Error when sending message:", err)
			return err
		}

		reqMsg := make([]string, 3, 3)
		reqMsg[0] = sender
		reqMsg[1] = ""
		reqMsg[2] = "OK"
		_, err = frontend.SendMessage(reqMsg)
		if err != nil {
			log.Print("Error when sending message:", err)
			return err
		}

	}
}

func launchBroker() error {
	frontend, err := zmq.NewSocket(zmq.ROUTER)
	if err != nil {
		log.Printf("Error when creating broker socket")
		return err
	}
	backend, err := zmq.NewSocket(zmq.ROUTER)
	if err != nil {
		log.Printf("Error when creating broker socket")
		return err
	}
	frontend.Bind("tcp://*:5558")
	backend.Bind("tcp://*:5559")
	return brokerListen(frontend, backend)
}