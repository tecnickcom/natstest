package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"
)

var emptyParamCases = []string{
	"--serverAddress=",
	"--natsAddress=",
}

func TestCliEmptyParamError(t *testing.T) {
	for _, param := range emptyParamCases {
		os.Args = []string{"natstest", param}
		cmd, err := cli()
		if err != nil {
			t.Error(fmt.Errorf("An error wasn't expected: %v", err))
			return
		}
		if cmdtype := reflect.TypeOf(cmd).String(); cmdtype != "*cobra.Command" {
			t.Error(fmt.Errorf("The expected type is '*cobra.Command', found: '%s'", cmdtype))
			return
		}

		old := os.Stderr // keep backup of the real stdout
		defer func() { os.Stderr = old }()
		os.Stderr = nil

		// execute the main function
		if err := cmd.Execute(); err == nil {
			t.Error(fmt.Errorf("An error was expected"))
		}
	}
}

func TestCliNoConfigError(t *testing.T) {
	os.Args = []string{"natstest", "--serverAddress=:8123", "--natsAddress=nats://127.0.0.1:3334"}
	cmd, err := cli()
	if err != nil {
		t.Error(fmt.Errorf("An error wasn't expected: %v", err))
		return
	}
	if cmdtype := reflect.TypeOf(cmd).String(); cmdtype != "*cobra.Command" {
		t.Error(fmt.Errorf("The expected type is '*cobra.Command', found: '%s'", cmdtype))
		return
	}

	old := os.Stderr // keep backup of the real stdout
	defer func() { os.Stderr = old }()
	os.Stderr = nil

	oldTestMap := testMap
	testMap = nil
	oldCfg := ConfigPath
	ConfigPath[0] = "wrong/path/0/"
	ConfigPath[1] = "wrong/path/1/"
	ConfigPath[2] = "wrong/path/2/"
	ConfigPath[3] = "wrong/path/3/"
	ConfigPath[4] = "wrong/path/4/"
	defer func() {
		ConfigPath = oldCfg
		testMap = oldTestMap
	}()

	// execute the main function
	if err := cmd.Execute(); err == nil {
		t.Error(fmt.Errorf("An error was expected"))
	}
}

func TestCli(t *testing.T) {
	os.Args = []string{"natstest", "--serverAddress=:8123", "--natsAddress=nats://127.0.0.1:4222"}
	cmd, err := cli()
	if err != nil {
		t.Error(fmt.Errorf("An error wasn't expected: %v", err))
		return
	}
	if cmdtype := reflect.TypeOf(cmd).String(); cmdtype != "*cobra.Command" {
		t.Error(fmt.Errorf("The expected type is '*cobra.Command', found: '%s'", cmdtype))
		return
	}

	old := os.Stderr // keep backup of the real stdout
	defer func() { os.Stderr = old }()
	os.Stderr = nil

	// use two separate channels for server and client testing
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()

		oldTestMap := testMap
		testMap = nil
		oldCfg := ConfigPath
		ConfigPath[0] = "../resources/test/etc/natstest/"
		ConfigPath[1] = "wrong/path/1/"
		ConfigPath[2] = "wrong/path/2/"
		ConfigPath[3] = "wrong/path/3/"
		ConfigPath[4] = "wrong/path/4/"
		defer func() {
			ConfigPath = oldCfg
			testMap = oldTestMap
		}()

		// start server
		if err := cmd.Execute(); err != nil {
			t.Error(fmt.Errorf("An error was not expected: %v", err))
		}
	}()
	go func() {
		defer wg.Done()

		// wait for the http server and NATS connection to start
		time.Sleep(2000 * time.Millisecond)

		// test index
		testEndPoint(t, "GET", "/", "", 200)
		// test 404
		testEndPoint(t, "GET", "/INVALID", "", 404)
		// test 405
		testEndPoint(t, "PATCH", "/", "", 405)
		// test status
		testEndPoint(t, "GET", "/status", "", 200)
		// test unknown test (wrong test name)
		testEndPoint(t, "GET", "/test/MISSING", "", 404)
		// test internal test config
		testEndPoint(t, "GET", "/test/@cli", "", 200)
		testEndPoint(t, "GET", "/test/one", "", 200)
		// test all, including a faulty json config test
		testEndPoint(t, "GET", "/test/all", "", 200)

		// test busy mode
		setBusy(true)
		defer setBusy(false)
		testEndPoint(t, "GET", "/status", "", 409)
		testEndPoint(t, "GET", "/test/@cli", "", 409)
		testEndPoint(t, "GET", "/reload", "", 409)
		testEndPoint(t, "DELETE", "/delete/@beta", "", 409)
		setBusy(false)

		// test new test
		testEndPoint(t, "PUT", "/new/alpha", "", 417)

		jsonraw := `[
	{
		"Topic" : "@.put.test",
		"Request" : {
			"integer" : 123,
			"name" : "some string",
			"previousReturnedValue" : "alpha",
			"array" : [
				{
					"key1" : "value2",
					"key2" : "beta"
				},
				{
					"key1" : "value2",
					"key2" : "value2 test string"
				}
			],
			"submap" : {
				"key1" : "gamma",
				"key2" : "delta"
			}
		},
		"Response" : {
			"integer" : "~re:[0-9]+",
			"name" : "~re:[a-z ]+",
			"previousReturnedValue" : "~pv:0.Request.previousReturnedValue",
			"array" : [
				{
					"key1" : "value2",
					"key2" : "beta"
				},
				{
					"key1" : "value2",
					"key2" : "value2 test string"
				}
			],
			"submap" : {
				"key1" : "gamma",
				"key2" : "delta"
			}
		}
	}
]`
		testEndPoint(t, "PUT", "/new/@alpha", jsonraw, 200)
		// test overwrite existing test
		testEndPoint(t, "PUT", "/new/@alpha", jsonraw, 200)
		// test new entry
		testEndPoint(t, "GET", "/test/@alpha", "", 200)

		// reset and reload test config files
		testEndPoint(t, "GET", "/reload", "", 200)

		// check the reload status
		testEndPoint(t, "GET", "/test/@alpha", "", 404)

		// test add/delete
		testEndPoint(t, "PUT", "/new/@beta", jsonraw, 200)
		testEndPoint(t, "DELETE", "/delete/@beta", "", 200)
		testEndPoint(t, "DELETE", "/delete/@beta", "", 404)

		// test "all" with an error file
		// error config
		jsonerr := `[
	{
		"Topic" : "@.put.error.test",
		"Request" : {
			"integer" : 123
		},
		"Response" : {
			"integer" : "~re:[a-z]+"
		}
	}
]`
		testEndPoint(t, "PUT", "/new/error", jsonerr, 417)
		testEndPoint(t, "GET", "/test/all", "", 417)

		// try to connect to the wrong nats bus address
		initNatsBus("nats://127.0.0.1:4333")
		testEndPoint(t, "GET", "/status", "", 503)
		initNatsBus("nats://127.0.0.1:4222")

		// test reload error
		oldTestMap := testMap
		testMap = nil
		oldCfg := ConfigPath
		ConfigPath[0] = "wrong/path/0/"
		ConfigPath[1] = "wrong/path/1/"
		ConfigPath[2] = "wrong/path/2/"
		ConfigPath[3] = "wrong/path/3/"
		ConfigPath[4] = "wrong/path/4/"
		defer func() {
			ConfigPath = oldCfg
			testMap = oldTestMap
		}()
		testEndPoint(t, "GET", "/reload", "", 500)

		// close the server goroutine
		wg.Done()
	}()
	wg.Wait()
}

// return true if the input is a JSON
func isJSON(s []byte) bool {
	var js map[string]interface{}
	return json.Unmarshal(s, &js) == nil
}

func testEndPoint(t *testing.T, method string, path string, data string, code int) {
	var payload = []byte(data)
	req, err := http.NewRequest(method, fmt.Sprintf("http://127.0.0.1:8123%s", path), bytes.NewBuffer(payload))
	if err != nil {
		t.Error(fmt.Errorf("An error was not expected: %v", err))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Error(fmt.Errorf("An error was not expected: %v", err))
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != code {
		t.Error(fmt.Errorf("The expected status code is %d, found %d", code, resp.StatusCode))
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(fmt.Errorf("An error was not expected: %v", err))
		return
	}
	if !isJSON(body) {
		t.Error(fmt.Errorf("The body is not a JSON"))
	}
}
