package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
)

// TestEntry defines a single entry in the test configuration file
type TestEntry struct {
	Topic    string      `json:"Topic"`    // topic name
	Request  interface{} `json:"Request"`  // raw message to be sent (input)
	Response interface{} `json:"Response"` // expected response message (output)
}

// TestEntries is a list of test entries
type TestEntries []TestEntry

// testMap contains the sequence of messages to send and the expected responses
var testMap map[string]TestEntries

// testCache contains the sequence of processed messages for the current test
var testCache TestEntries

// testNames contains the test names that can be used as entry points
var testNames []string

// return a list of configuration test files for each type
func loadTestMap() error {
	testMap = make(map[string]TestEntries)
	testNames = make([]string, 0)
	// extract the topic name
	re := regexp.MustCompile(`test_([@a-zA-Z0-9]+)\.json$`)
	// for each configuration directory
	for _, cpath := range ConfigPath {
		// find the test files
		files, _ := filepath.Glob(cpath + "/test_*.json")
		for _, file := range files {
			// extract the test name
			key := re.FindStringSubmatch(file)
			if _, exist := testMap[key[1]]; !exist {
				// store the test config file (local files have priority)
				raw, err := ioutil.ReadFile(file) // #nosec
				if err != nil {
					return fmt.Errorf("unable to read the configuration file: %v", err)
				}
				err = loadRawJSONTest(raw, key[1])
				if err != nil {
					return fmt.Errorf("unable to decode the test file %s: %v", file, err)
				}
			}
		}
	}
	if len(testMap) == 0 {
		return fmt.Errorf("unable to find valid test configuration files")
	}
	return nil
}

// load the test from a JSON string
func loadRawJSONTest(raw []byte, name string) (err error) {
	var testData TestEntries
	err = json.Unmarshal(raw, &testData)
	if err != nil {
		return err
	}
	_, replace := testMap[name]
	testMap[name] = testData
	if !replace {
		testNames = append(testNames, name)
	}
	return nil
}

// execute the specified test
func execTest(test TestEntries) (err error) {
	var request []byte
	var response []byte
	var resp interface{}
	var expresp interface{}

	testCache = make(TestEntries, len(test))

	err = openNatsBus()
	if err != nil {
		return err
	}
	defer closeNatsBus()

	for item, msg := range test {

		// prepare the request
		msg.Request, err = replaceTemplates(msg.Request)
		if err != nil {
			return fmt.Errorf("%s [%d]: unable to process templates on request message %v - %v", msg.Topic, item, msg.Request, err)
		}

		// save the processed message
		testCache[item].Topic = msg.Topic
		testCache[item].Request = msg.Request

		// encode the request
		request, err = json.Marshal(msg.Request)
		if err != nil {
			return fmt.Errorf("%s [%d]: unable to encode request message %v %v", msg.Topic, item, msg.Request, err)
		}

		// send the request message and get the response
		response, err = sendBusRequest(msg.Topic, request)
		if err != nil {
			return fmt.Errorf("%s [%d]: unable to send request message %v %v", msg.Topic, item, msg.Request, err)
		}

		// decode the response message
		err = json.Unmarshal(response, &resp)
		if err != nil {
			return fmt.Errorf("%s [%d]: unable to decode the response message: %v", msg.Topic, item, err)
		}

		// save the response message value for templates
		testCache[item].Response = resp

		// replace templates
		expresp, err = replaceTemplates(msg.Response)
		if err != nil {
			return fmt.Errorf("%s [%d]: unable to process templates on response message %v - %v", msg.Topic, item, resp, err)
		}

		// compare the expected and actual messages
		err = areMatching(expresp, resp)
		if err != nil {
			return fmt.Errorf("%s [%d]: the messages are different: %v", msg.Topic, item, err)
		}
	}
	return nil
}
