package main

import (
	"fmt"
	"testing"
)

func TestLogInvalidLevel(t *testing.T) {
	err := Log(10, "test")
	if err == nil {
		t.Error(fmt.Errorf("An error was expected"))
	}
}
