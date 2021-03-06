package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/spf13/cobra"
)

var stopTestServerChan chan bool

var badParamCases = []string{
	"--serverAddress=",
	"--natsAddress=",
	"--logLevel=",
	"--logLevel=INVALID",
}

func TestCliBadParamError(t *testing.T) {
	for _, param := range badParamCases {
		os.Args = []string{ProgramName, param}
		cmd, err := cli()
		if err != nil {
			t.Error(fmt.Errorf("Unexpected error: %v", err))
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

func TestWrongParamError(t *testing.T) {
	os.Args = []string{ProgramName, "--unknown"}
	_, err := cli()
	if err == nil {
		t.Error(fmt.Errorf("An error was expected"))
		return
	}
}

func TestCliNoConfigError(t *testing.T) {
	os.Args = []string{
		ProgramName,
		"--serverAddress=:8123",
		"--natsAddress=nats://127.0.0.1:3334",
		"--logLevel=DEBUG",
	}
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
	for k := range ConfigPath {
		ConfigPath[k] = "wrong/path/"
	}
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
	os.Args = []string{ProgramName}
	cmd, err := cli()
	if err != nil {
		t.Error(fmt.Errorf("Unexpected error: %v", err))
		return
	}
	if cmdtype := reflect.TypeOf(cmd).String(); cmdtype != "*cobra.Command" {
		t.Error(fmt.Errorf("The expected type is '*cobra.Command', found: '%s'", cmdtype))
		return
	}

	old := os.Stderr // keep backup of the real stdout
	defer func() { os.Stderr = old }()
	os.Stderr = nil

	// add an endpoint to test the panic handler
	routes = append(routes,
		Route{
			"GET",
			"/panic",
			triggerPanic,
			"TRIGGER PANIC",
		})
	defer func() { routes = routes[:len(routes)-1] }()

	// use two separate channels for server and client testing
	var twg sync.WaitGroup
	startTestServer(t, cmd, &twg)
	startTestClient(t)
	twg.Wait()
}

func startTestServer(t *testing.T, cmd *cobra.Command, twg *sync.WaitGroup) {

	stopTestServerChan = make(chan bool)

	twg.Add(1)
	go func() {
		defer twg.Done()

		chp := make(chan error, 1)
		go func() {
			chp <- cmd.Execute()
		}()

		stopped := false
		for {
			select {
			case err, ok := <-chp:
				if ok && !stopped && err != nil {
					stopTestServerChan <- true
					t.Error(fmt.Errorf("An error was not expected: %v", err))
				}
				return
			case <-stopTestServerChan:
				stopped = true
				err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
				if err != nil {
					t.Error(fmt.Errorf("An error was not expected: %v", err))
				}
			}
		}
	}()

	// wait for the server to start
	time.Sleep(500 * time.Millisecond)
}

func startTestClient(t *testing.T) {

	// check if the server is running
	select {
	case stop, ok := <-stopTestServerChan:
		if ok && stop {
			return
		}
	default:
		break
	}

	defer func() { stopTestServerChan <- true }()

	testEndPoint(t, "GET", "/", "", 200)
	testEndPoint(t, "GET", "/status", "", 200)

	// error conditions

	testEndPoint(t, "GET", "/INVALID", "", 404) // NotFound
	testEndPoint(t, "PATCH", "/", "", 405)      // MethodNotAllowed
	testEndPoint(t, "GET", "/panic", "", 500)   // PanicHandler

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

	// invalid comparison command
	jsonerr = `[
	{
		"Topic" : "@.put.error.test",
		"Request" : {
			"value" : "ciao"
		},
		"Response" : {
			"value" : "~xc:/invalid/:ciao"
		}
	}
]`
	testEndPoint(t, "PUT", "/new/error", jsonerr, 417)

	// comparison error
	jsonerr = `[
	{
		"Topic" : "@.put.error.test",
		"Request" : {
			"value" : {"key":"ciao"}
		},
		"Response" : {
			"value" : "~xc:/bin/cat:ciao"
		}
	}
]`
	testEndPoint(t, "PUT", "/new/error", jsonerr, 417)

	// try to connect to the wrong nats bus address
	initNatsBus("nats://127.0.0.1:4333")
	testEndPoint(t, "GET", "/status", "", 503)
	initNatsBus("nats://127.0.0.1:4222")

	// test reload error
	oldTestMap := testMap
	testMap = nil
	oldCfg := ConfigPath
	for k := range ConfigPath {
		ConfigPath[k] = "wrong/path/"
	}
	defer func() {
		ConfigPath = oldCfg
		testMap = oldTestMap
	}()
	testEndPoint(t, "GET", "/reload", "", 500)
}

// triggerPanic triggers a Panic
func triggerPanic(rw http.ResponseWriter, hr *http.Request, ps httprouter.Params) {
	panic("TEST PANIC")
}

// isJSON returns true if the input is JSON
func isJSON(s []byte) bool {
	var js map[string]interface{}
	return json.Unmarshal(s, &js) == nil
}

func testEndPoint(t *testing.T, method string, path string, data string, code int) {
	var payload = []byte(data)
	req, err := http.NewRequest(method, fmt.Sprintf("http://127.0.0.1:8000%s", path), bytes.NewBuffer(payload))
	if err != nil {
		t.Error(fmt.Errorf("An error was not expected: %v", err))
		return
	}
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Error(fmt.Errorf("An error was not expected: %v", err))
		return
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			t.Error(fmt.Errorf("An error was not expected: %v", err))
			return
		}
	}()

	body, err := ioutilReadAll(resp.Body)
	if err != nil {
		t.Error(fmt.Errorf("An error was not expected: %v", err))
		return
	}

	if resp.StatusCode != code {
		t.Error(fmt.Errorf("The expected '%s' status code is %d, found %d", path, code, resp.StatusCode))
		return
	}

	if !isJSON(body) {
		t.Error(fmt.Errorf("The body is not JSON"))
	}
}
