package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"testing"
)

func getMainOutput() (string, error) {
	old := os.Stdout // keep backup of the real stdout
	defer func() { os.Stdout = old }()
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w

	// execute the main function
	main()

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// back to normal state
	w.Close()
	out := <-outC

	return out, nil
}

func TestMainVersion(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"natstest", "version"}
	out, err := getMainOutput()
	if err != nil {
		t.Error(fmt.Errorf("Unexpected error: %v", err))
	}
	match, err := regexp.MatchString("^[\\d]+\\.[\\d]+\\.[\\d]+[\\s]*$", out)
	if err != nil {
		t.Error(fmt.Errorf("Unexpected error: %v", err))
	}
	if !match {
		t.Error(fmt.Errorf("The version has an invalid format"))
	}
}
