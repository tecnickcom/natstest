package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

// busy is a mutex variable to force only one process at the time
var busy = false
var startTime = time.Now()

// setBusy set the busy value
func setBusy(val bool) {
	busy = val
}

// return a list of available routes
func index(rw http.ResponseWriter, hr *http.Request, ps httprouter.Params) {
	type info struct {
		Busy     bool     `json:"busy"`     // true if a test is in progress
		Duration float64  `json:"duration"` // elapsed time since last test start in seconds
		Entries  Routes   `json:"routes"`   // available routes (http entry points)
		Tests    []string `json:"tests"`    // available test names
	}
	sendResponse(rw, hr, ps, http.StatusOK, info{
		Busy:     busy,
		Duration: time.Since(startTime).Seconds(),
		Entries:  routes,
		Tests:    testNames,
	})
}

// returns the status of the service
func status(rw http.ResponseWriter, hr *http.Request, ps httprouter.Params) {
	if busy {
		sendBusyResponse(rw, hr, ps)
		return
	}

	type info struct {
		Busy           bool    `json:"busy"`     // true if a test is in progress
		Duration       float64 `json:"duration"` // elapsed time since last test start in seconds
		NatsConnection bool    `json:"nats"`     // true if the nats connection is working
		Message        string  `json:"message"`  // error message
	}
	status := http.StatusOK
	natsConnection := true
	message := "The service is healthy"
	err := openNatsBus()
	if err != nil {
		status = http.StatusServiceUnavailable
		natsConnection = false
		message = fmt.Sprintf("Unable to connect to the NATS bus: %v", natsOpts.Servers)
	} else {
		closeNatsBus()
	}
	sendResponse(rw, hr, ps, status, info{
		Busy:           busy,
		Duration:       time.Since(startTime).Seconds(),
		NatsConnection: natsConnection,
		Message:        message,
	})
}

// send a "BUSY" response
func sendBusyResponse(rw http.ResponseWriter, hr *http.Request, ps httprouter.Params) {
	type infoBusy struct {
		Busy     bool    `json:"busy"`     // true if a test is in progress
		Duration float64 `json:"duration"` // elapsed time since last test start in seconds
		Message  string  `json:"message"`  // error message
	}
	sendResponse(rw, hr, ps, http.StatusConflict, infoBusy{
		Busy:     busy,
		Duration: time.Since(startTime).Seconds(),
		Message:  "Another test is already in progress, please wait ...",
	})
}

// test all components
func test(rw http.ResponseWriter, hr *http.Request, ps httprouter.Params) {

	if busy {
		sendBusyResponse(rw, hr, ps)
		return
	}
	setBusy(true)
	defer setBusy(false)

	startTime = time.Now()
	testCounter := 0

	var err error

	test := ps.ByName("name")
	if test == "all" {
		// execute all tests
		for key, testData := range testMap {
			if key[0] != '@' { // exclude internal tests
				testCounter++
				err = execTest(testData)
				if err != nil {
					// stop as soon a one test fails
					break
				}
			}
		}
	} else if _, exist := testMap[test]; exist {
		// execute the selected test
		testCounter++
		err = execTest(testMap[test])
	} else {
		// test not found
		sendResponse(rw, hr, ps, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}
	if err != nil {
		sendResponse(rw, hr, ps, http.StatusExpectationFailed, err.Error())
		return
	}

	type info struct {
		Tests    int     `json:"tests"`    // number of executed tests
		Duration float64 `json:"duration"` // full test duration in seconds
		Message  string  `json:"message"`  // message
	}
	sendResponse(rw, hr, ps, http.StatusOK, info{
		Tests:    testCounter,
		Duration: time.Since(startTime).Seconds(),
		Message:  "All tests completed successfully",
	})
}

// reload and reset all tests from configuration files
func reload(rw http.ResponseWriter, hr *http.Request, ps httprouter.Params) {
	if busy {
		sendBusyResponse(rw, hr, ps)
		return
	}
	err := loadTestMap()
	if err != nil {
		sendResponse(rw, hr, ps, http.StatusInternalServerError, err.Error())
		return
	}
	sendResponse(rw, hr, ps, http.StatusOK, "the test configuration files were successfully reloaded")
}

// load and execute the test sent via PUT
func newtest(rw http.ResponseWriter, hr *http.Request, ps httprouter.Params) {
	body, err := ioutil.ReadAll(hr.Body)
	if err != nil {
		sendResponse(rw, hr, ps, http.StatusInternalServerError, err.Error())
		return
	}
	err = loadRawJSONTest(body, ps.ByName("name"))
	if err != nil {
		sendResponse(rw, hr, ps, http.StatusExpectationFailed, err.Error())
		return
	}
	test(rw, hr, ps)
}

// remove the specified test
func deltest(rw http.ResponseWriter, hr *http.Request, ps httprouter.Params) {
	if busy {
		sendBusyResponse(rw, hr, ps)
		return
	}
	name := ps.ByName("name")
	for item, value := range testNames {
		if value == name {
			testNames = append(testNames[:item], testNames[item+1:]...)
			delete(testMap, name)
			sendResponse(rw, hr, ps, http.StatusOK, fmt.Sprintf("the test %s has been successfully removed", name))
			return
		}
	}
	sendResponse(rw, hr, ps, http.StatusNotFound, fmt.Sprintf("unable to find the test %s", name))
}
