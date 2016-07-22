package brokerlib

import (
	"fmt"
	"log"

	"github.com/bonnefoa/signaler/conf"
	zmq "github.com/pebbe/zmq4"
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
			_, err = frontend.SendMessage([]string{sender, "", "KO"})
			if err != nil {
				log.Print("Error when sending message:", err)
				continue
			}
		}
		_, err = frontend.SendMessage([]string{sender, "", "OK"})
		if err != nil {
			log.Print("Error when sending message:", err)
			return err
		}

	}
}

func createBrokerSocket(port int) (*zmq.Socket, error) {
	socket, err := zmq.NewSocket(zmq.ROUTER)
	if err != nil {
		log.Print("Error when creating broker socket", err)
		return nil, err
	}
	err = socket.SetRouterHandover(true)
	if err != nil {
		log.Print("Error when setting handover", err)
		return nil, err
	}
	socket.SetRouterMandatory(1)
	if err != nil {
		log.Print("Error when setting router mandatory", err)
		return nil, err
	}
	err = socket.Bind(fmt.Sprintf("tcp://*:%d", port))
	if err != nil {
		log.Printf("Error when binding to port %d: %s", port, err)
		return nil, err
	}
	return socket, nil
}

// LaunchBroker start broker for message routing
func LaunchBroker() error {
	frontend, err := createBrokerSocket(*conf.FrontendPort)
	if err != nil {
		return err
	}
	backend, err := createBrokerSocket(*conf.BackendPort)
	if err != nil {
		return err
	}
	return brokerListen(frontend, backend)
}
