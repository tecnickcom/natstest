package main

import (
	"fmt"
	"io"
)

func mockJSONMarshalError(v interface{}) ([]byte, error) {
	return nil, fmt.Errorf("SIMULATED json.Marshal ERROR")
}

func mockSendJSONEncode(w io.Writer, v interface{}) error {
	return fmt.Errorf("SIMULATED sendJSONEncode ERROR")
}

func mockIoutilReadAll(r io.Reader) ([]byte, error) {
	return nil, fmt.Errorf("SIMULATED ioutil.ReadAll ERROR")
}
