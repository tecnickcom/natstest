package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

// validTransfCommandMap is a map containing the list of valid transformation commands
var isValidTransfCmd = make(map[string]bool)

// jsonOpenMark is the string identifying the beginning of a JSON string
const jsonStartMark = "#@~"

// jsonOpenMark is the string identifying the end of a JSON string
const jsonEndMark = "~@#"

// replaceTemplates replace all templates with the corresponding value
func replaceTemplates(obj interface{}) (interface{}, error) {
	// wrap original in a reflect.Value
	original := reflect.ValueOf(obj)

	if original.Kind() == reflect.Invalid {
		return nil, fmt.Errorf("the request data is invalid")
	}

	// create a copy
	copy := reflect.New(original.Type()).Elem()
	// replace templates
	processTemplates(copy, original)
	// encode the copy interface as json
	jsoncopy, err := json.Marshal(copy.Interface())

	// unescape marked json strings
	search := regexp.MustCompile("(?U)\"" + jsonStartMark + ".*" + jsonEndMark + "\"")
	jsoncopy = search.ReplaceAllFunc(jsoncopy, func(str []byte) []byte {
		str = str[(len(jsonStartMark) + 1):(len(str) - len(jsonEndMark) - 1)]
		ustr, err := strconv.Unquote("\"" + string(str) + "\"")
		if err != nil {
			return []byte("\"\"")
		}
		return []byte(ustr)
	})

	// decode json
	var ret interface{}
	err = json.Unmarshal(jsoncopy, &ret)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode JSON: %s", jsoncopy)
	}

	return ret, nil
}

// processTemplates find and replace individual templates
// NOTE: some of this code based on https://gist.github.com/hvoecking/10772475 (MIT LICENSE)
func processTemplates(copy, original reflect.Value) {
	switch original.Kind() {
	// The first cases handle nested structures and process them recursively

	// invalid kind
	case reflect.Invalid:
		return

	// If it is a pointer we need to unwrap and call once again
	case reflect.Ptr:
		processTemplatePtr(copy, original)

	// If it is an interface (which is very similar to a pointer), do basically the
	// same as for the pointer. Though a pointer is not the same as an interface so
	// note that we have to call Elem() after creating a new object because otherwise
	// we would end up with an actual pointer
	case reflect.Interface:
		processTemplateInterface(copy, original)

	// If it is a struct we process each field
	case reflect.Struct:
		processTemplateStruct(copy, original)

	// If it is a slice we create a new slice and process each element
	case reflect.Slice:
		processTemplateSlice(copy, original)

	// If it is a map we create a new map and process each value
	case reflect.Map:
		processTemplateMap(copy, original)

	// Otherwise we cannot traverse anywhere so this finishes the recursion

	// If it is a string process, check if it is a template
	case reflect.String:
		processTemplateString(copy, original)

	// And everything else will simply be taken from the original
	default:
		copy.Set(original)
	}
}

// processTemplatePtr process the Ptr case
func processTemplatePtr(copy, original reflect.Value) {
	// To get the actual value of the original we have to call Elem()
	// At the same time this unwraps the pointer so we don't end up in
	// an infinite recursion
	originalValue := original.Elem()
	// Check if the pointer is nil
	if !originalValue.IsValid() {
		return
	}
	// Allocate a new object and set the pointer to it
	copy.Set(reflect.New(originalValue.Type()))
	// Unwrap the newly created pointer
	processTemplates(copy.Elem(), originalValue)
}

// processTemplateInterface process the Interface case
func processTemplateInterface(copy, original reflect.Value) {
	// Get rid of the wrapping interface
	originalValue := original.Elem()
	// Check if the pointer is nil
	if !originalValue.IsValid() {
		return
	}
	// Create a new object. Now new gives us a pointer, but we want the value it
	// points to, so we have to call Elem() to unwrap it
	copyValue := reflect.New(originalValue.Type()).Elem()
	processTemplates(copyValue, originalValue)
	copy.Set(copyValue)
}

// processTemplateStruct process the Struct case
func processTemplateStruct(copy, original reflect.Value) {
	for i := 0; i < original.NumField(); i++ {
		processTemplates(copy.Field(i), original.Field(i))
	}
}

// processTemplateSlice process the Slice case
func processTemplateSlice(copy, original reflect.Value) {
	copy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
	for i := 0; i < original.Len(); i++ {
		processTemplates(copy.Index(i), original.Index(i))
	}
}

// processTemplateMap process the Map case
func processTemplateMap(copy, original reflect.Value) {
	copy.Set(reflect.MakeMap(original.Type()))
	for _, key := range original.MapKeys() {
		originalValue := original.MapIndex(key)
		// New gives us a pointer, but again we want the value
		copyValue := reflect.New(originalValue.Type()).Elem()
		processTemplates(copyValue, originalValue)
		copy.SetMapIndex(key, copyValue)
	}
}

// processTemplateString process the String case
func processTemplateString(copy, original reflect.Value) {
	value := original.Interface().(string)
	tmark := value[0:int(math.Min(float64(len(value)), float64(4)))] // template marker
	if tmark == "~ts:" {
		// replace the template with the current time
		t := time.Now().UTC()
		// extract time format
		timeFormat := value[4:]
		if timeFormat != "" {
			copy.SetString(t.Format(timeFormat))
		} else {
			// unix timestamp in seconds
			jenc, err := json.Marshal(int32(t.Unix()))
			if err == nil {
				copy.SetString(jsonStartMark + string(jenc) + jsonEndMark)
			}
		}
	} else if tmark == "~pv:" {
		// replace the template with the real value
		newval := getFieldValue(value[4:], testCache).Interface()
		if reflect.TypeOf(newval).Kind() == reflect.String {
			// the replacement value is also a string
			copy.SetString(newval.(string))
		} else {
			// encode the replacement value as JSON string (to be decoded later)
			jenc, err := json.Marshal(newval)
			if err == nil {
				copy.SetString(jsonStartMark + string(jenc) + jsonEndMark)
			}
		}
	} else {
		// this is not a template; copy the value
		copy.Set(original)
	}
}

// getFieldValue returns the data value specified by the path
func getFieldValue(path string, data interface{}) reflect.Value {
	cache := reflect.ValueOf(data)

	// separate the template from the transformation statement
	parts := strings.SplitN(path, ">", 2)

	// extract the path keys
	keys := strings.Split(parts[0], ".")
	for _, key := range keys {
		if cache.Kind() == reflect.Interface || cache.Kind() == reflect.Ptr {
			cache = cache.Elem()
		}
		if cache.Kind() == reflect.Map {
			cache = cache.MapIndex(reflect.ValueOf(key))
		} else if cache.Kind() == reflect.Struct {
			cache = cache.FieldByName(key)
		} else if cache.Kind() == reflect.Slice {
			idx, _ := strconv.Atoi(key)
			cache = cache.Index(idx)
		}
	}
	// process the transformation (if any)
	if len(parts) == 2 {
		val, err := execTransfCmd(parts[1], cache)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("unable to execute the trasformation command")
		}
		return val
	}
	return cache
}

// execTransfCmd execute the specified command template
func execTransfCmd(template string, value reflect.Value) (reflect.Value, error) {
	var strvalue string
	if reflect.TypeOf(value.Interface()).Kind() == reflect.String {
		strvalue = value.Interface().(string)
	} else {
		// encode the object as JSON string
		jsonval, err := json.Marshal(value.Interface())
		if err != nil {
			return value, fmt.Errorf("unable to json-encode the value: %#v -- [%v]", value.Interface(), err)
		}
		strvalue = string(jsonval)
	}
	parts := strings.Fields(template)
	if len(parts) == 1 {
		return value, fmt.Errorf("the command is missing arguments: %v", template)
	}
	if !isValidTransfCmd[parts[0]] {
		return value, fmt.Errorf("the following command is not valid: %v", parts[0])
	}
	args := parts[1:]
	// search and replace the input argument value (%v)
	for i := range args {
		if args[i] == "%v" {
			args[i] = strvalue
			break
		}
	}
	out, err := exec.Command(parts[0], args...).Output()
	// #nosec
	if err != nil {
		return value, fmt.Errorf("unable to run the command: %v -- [%v]", template, err)
	}
	return reflect.ValueOf(strings.Trim(string(out), "\n")), nil
}
