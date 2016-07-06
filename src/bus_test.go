package main

import (
	"fmt"
	"testing"
)

func TestOpenNatsBusError(t *testing.T) {
	initNatsBus("nats://127.0.0.1:3333")
	err := openNatsBus()
	if err == nil {
		closeNatsBus()
		t.Error(fmt.Errorf("an error was expected"))
	}
}

func TestSendBusRequestError(t *testing.T) {
	initNatsBus("nats://127.0.0.1:4222")
	err := openNatsBus()
	if err != nil {
		t.Error(fmt.Errorf("Error connecting to the NATS bus"))
	}
	defer closeNatsBus()
	_, err = sendBusRequest("topic", []byte("ABC"))
	if err == nil {
		t.Error(fmt.Errorf("an error was expected"))
	}
}
