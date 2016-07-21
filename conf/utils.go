package conf

import "log"
import zmq "github.com/pebbe/zmq4"

// SetupZmq configure zmq parameters like max sockets
func SetupZmq() error {
	err := zmq.SetMaxSockets(*MaxZmqSockets)
	if err != nil {
		log.Print("Error on zmq setup", err)
		return err
	}
	return nil
}
