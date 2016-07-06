package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"testing"
)

func getMainOutput() string {
	old := os.Stdout // keep backup of the real stdout
	defer func() { os.Stdout = old }()
	r, w, _ := os.Pipe()
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

	return out
}

func TestMainVersion(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"natstest", "version"}
	out := getMainOutput()
	match, _ := regexp.MatchString("^[\\d]+\\.[\\d]+\\.[\\d]+[\\s]*$", out)
	if !match {
		t.Error(fmt.Errorf("The version has an invalid format"))
	}
}
