package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGetFieldValueArray(t *testing.T) {
	val := getFieldValue("0.Response.array", testMap["@internal"])
	result := fmt.Sprintf("%#v", val)
	expected := "[]interface {}{map[string]interface {}{\"key1\":\"value2\", \"key2\":\"beta\"}, map[string]interface {}{\"key1\":\"value2\", \"key2\":\"value2 test string\"}}"
	if result != expected {
		t.Error(fmt.Errorf("Found different value than expected: %v", result))
	}
}

func TestGetFieldValueString(t *testing.T) {
	val := getFieldValue("0.Request.array.1.key2", testMap["@internal"])
	if val.Interface().(string) != "value2 test string" {
		t.Error(fmt.Errorf("Found different value than expected: %v", val.Interface()))
	}
}

func TestGetFieldValueNum(t *testing.T) {
	val := getFieldValue("1.Response.integer", testMap["@internal"])
	if val.Interface().(float64) != 123 {
		t.Error(fmt.Errorf("Found different value than expected: %v", val.Interface()))
	}
}

func TestGetFieldValueErr(t *testing.T) {
	val := getFieldValue("0.Request.name>/wrong_cmd", testMap["@internal"])
	if val.Interface().(string) != "some string" {
		t.Error(fmt.Errorf("Found different value than expected: %v", val.Interface()))
	}
}

func TestReplaceTemplates(t *testing.T) {

	testCache = testMap["@internal"]

	res0, err := replaceTemplates(testMap["@internal"][0].Response)
	if err != nil {
		t.Error(fmt.Errorf("error while processing templates: %v", err))
	}

	testCache[0].Response = res0
	testCache[0].Response.(map[string]interface{})["integer"] = 123
	testCache[0].Response.(map[string]interface{})["name"] = "some string"

	if res0.(map[string]interface{})["previousReturnedValue"].(string) != "alpha" {
		t.Error(fmt.Errorf("Found different value than expected 'alpha'"))
	}

	if res0.(map[string]interface{})["array"].([]interface{})[0].(map[string]interface{})["key2"].(string) != "beta" {
		t.Error(fmt.Errorf("Found different value than expected 'beta'"))
	}

	if res0.(map[string]interface{})["submap"].(map[string]interface{})["key2"].(string) != "delta" {
		t.Error(fmt.Errorf("Found different value than expected '200'"))
	}

	res1, err := replaceTemplates(testMap["@internal"][1].Request)
	if err != nil {
		t.Error(fmt.Errorf("error while processing templates: %v", err))
	}

	testCache[0].Request = res1

	if res1.(map[string]interface{})["previousReturnedValue"].(string) != "some string" {
		t.Error(fmt.Errorf("Found different value than expected 'alpha'"))
	}

	if res1.(map[string]interface{})["array"].([]interface{})[0].(map[string]interface{})["key2"].(string) != "beta" {
		t.Error(fmt.Errorf("Found different value than expected 'beta'"))
	}
}

func TestReplaceTemplatesErrors(t *testing.T) {
	_, err := replaceTemplates(nil)
	if err == nil {
		t.Error(fmt.Errorf("an error was expected"))
	}

	type Vertex struct {
		X int
		Y int
	}

	// pointer
	v1 := Vertex{3, 5}
	_, err = replaceTemplates(&v1)
	if err != nil {
		t.Error(fmt.Errorf("error while processing templates: %v", err))
	}

	// nil pointer
	var n interface{}
	_, err = replaceTemplates(&n)
	if err != nil {
		t.Error(fmt.Errorf("error while processing templates: %v", err))
	}
}

func TestPocessTemplatesErrors(t *testing.T) {
	a := reflect.ValueOf(nil)
	processTemplates(a, a)
}

func TestExecCmdTemplate(t *testing.T) {
	ret, err := execTransfCmd("/bin/echo -n $%v", reflect.ValueOf("ciao ciao"))
	if err != nil {
		t.Error(fmt.Errorf("an error was not expected: %v", err))
	}
	if ret.Interface().(string) != "$ciao ciao" {
		t.Error(fmt.Errorf("a different return value was expected"))
	}
}

func TestExecCmdTemplateErr(t *testing.T) {
	ret, err := execTransfCmd("/bin/echo %v", reflect.ValueOf(""))
	if err == nil {
		t.Error(fmt.Errorf("an error was expected"))
	}

	ret, err = execTransfCmd("/bin/cat -l %v", reflect.ValueOf("dog"))
	if err == nil {
		t.Error(fmt.Errorf("an error was expected"))
	}

	ret, err = execTransfCmd("/missing_param", reflect.ValueOf("ciao ciao"))
	if err == nil {
		t.Error(fmt.Errorf("an error was expected"))
	}
	if ret.Interface().(string) != "ciao ciao" {
		t.Error(fmt.Errorf("a different return value was expected"))
	}
}
