package main

import (
	"fmt"
	"time"

	"github.com/nats-io/nats"
	log "github.com/sirupsen/logrus"
)

var busTimeout = time.Duration(BusTimeout) * time.Second
var natsOpts = nats.DefaultOptions
var natsConn *nats.Conn

// connect to the NATS bus
func initNatsBus(addr string) {
	log.WithFields(log.Fields{
		"nats": addr,
	}).Info("initializing NATS bus")
	natsOpts.Servers = []string{addr}
}

// connect to the NATS bus
func openNatsBus() (err error) {
	log.WithFields(log.Fields{
		"nats": natsOpts.Servers,
	}).Info("opening NATS bus connection")
	natsConn, err = natsOpts.Connect()
	if err != nil {
		return fmt.Errorf("Can't connect to the NATS message queue %v", err)
	}
	return nil
}

// close the NATS bus
func closeNatsBus() {
	log.WithFields(log.Fields{
		"nats": natsOpts.Servers,
	}).Info("closing NATS bus connection")
	err := natsConn.Flush()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Flush NATS bus connection")
	}
	natsConn.Close()
}

// send a message to the specified topic and get the raw answer
func sendBusRequest(topic string, request []byte) ([]byte, error) {
	if topic[0] == '@' {
		// echo the request for internal testing
		return request, nil
	}
	msg, err := natsConn.Request(topic, request, busTimeout)
	if err != nil {
		if err == nats.ErrTimeout {
			err = fmt.Errorf("request timeout: %v", err)
		}
		return nil, err
	}
	return msg.Data, nil
}
