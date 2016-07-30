package main

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"regexp"
)

// check if the messages are matching
func areMatching(expected interface{}, actual interface{}) (err error) {
	err = checkMatch(reflect.ValueOf(expected), reflect.ValueOf(actual))
	if err != nil {
		return getFormattedDiffError(err.Error(), expected, actual)
	}
	return nil
}

// checkMatch is a recursive function to check if the fields defined
// in "expected" are defined and have the same value in "actual"
func checkMatch(expected reflect.Value, actual reflect.Value) (err error) {

	if (expected.Kind() != actual.Kind()) && (expected.Kind() != reflect.String) {
		return getFormattedDiffError("the types are different", expected, actual)
	}

	switch expected.Kind() {

	case reflect.Invalid:
		return getFormattedDiffError("invalid kind", expected, actual)

	case reflect.Ptr:
		fallthrough

	case reflect.Interface:
		return checkMatch(expected.Elem(), actual.Elem())

	case reflect.Struct:
		return processCompareStruct(expected, actual)

	case reflect.Slice:
		return processCompareSlice(expected, actual)

	case reflect.Map:
		return processCompareMap(expected, actual)

	default:
		return processCompareDefault(expected, actual)

	}
}

// processCompareStruct process the Struct case
func processCompareStruct(expected reflect.Value, actual reflect.Value) (err error) {
	if expected.NumField() > actual.NumField() {
		return getFormattedDiffError("missing struct fields", expected, actual)
	}
	for i := 0; i < expected.NumField(); i++ {
		err = checkMatch(expected.Field(i), actual.Field(i))
		if err != nil {
			return err
		}
	}
	return nil
}

// processCompareSlice process the Slice case
func processCompareSlice(expected reflect.Value, actual reflect.Value) (err error) {
	if expected.Len() > actual.Len() {
		return getFormattedDiffError("missing slice items", expected, actual)
	}
	for i := 0; i < expected.Len(); i++ {
		err = checkMatch(expected.Index(i), actual.Index(i))
		if err != nil {
			return err
		}
	}
	return nil
}

// processCompareMap process the Map case
func processCompareMap(expected reflect.Value, actual reflect.Value) (err error) {
	for _, key := range expected.MapKeys() {
		err = checkMatch(expected.MapIndex(key), actual.MapIndex(key))
		if err != nil {
			return err
		}
	}
	return nil
}

// processCompareDefault process the Default case
func processCompareDefault(expected reflect.Value, actual reflect.Value) (err error) {
	if expected.Interface() == actual.Interface() {
		return nil
	}
	if expected.Kind() != reflect.String {
		return getFormattedDiffError("values are different", expected, actual)
	}
	// check if the expected value is a regular expression
	value := expected.Interface().(string)
	if len(value) == 0 || value[0] != '~' || value[0:int(math.Min(float64(len(value)), float64(4)))] != "~re:" {
		// the value is not a regular expression
		return getFormattedDiffError("values are different", expected, actual)
	}
	// compare using a regular expression
	sv := fmt.Sprintf("%v", actual.Interface())
	match, err := regexp.MatchString(value[4:], sv)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	if !match {
		return getFormattedDiffError("the regular expression do not match", expected, actual)
	}
	return nil
}

// getFormattedDiffError returns a json string containing the expected and actual object
func getFormattedDiffError(message string, expected interface{}, actual interface{}) error {
	type CompareError struct {
		Error    string      `json:"error"`    // error message
		Expected interface{} `json:"expected"` // expected value
		Actual   interface{} `json:"actual"`   // actual value
	}
	errStruct := &CompareError{
		Error:    message,
		Expected: expected,
		Actual:   actual,
	}
	errmsg, err := json.Marshal(errStruct)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return fmt.Errorf("%s", string(errmsg))
}
