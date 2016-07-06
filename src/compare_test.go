package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestAreMatching(t *testing.T) {

	err := areMatching("test", "test")
	if err != nil {
		t.Error(fmt.Errorf("error while processing templates: %v", err))
	}

	err = areMatching("alpha", "beta")
	if err == nil {
		t.Error(fmt.Errorf("an error was expected"))
	}

	err = areMatching("~re:[a-z]+", "test")
	if err != nil {
		t.Error(fmt.Errorf("error while processing templates: %v", err))
	}

	err = areMatching("~re:[0-9]+", 123)
	if err != nil {
		t.Error(fmt.Errorf("error while processing templates: %v", err))
	}

	err = areMatching("~re:[0-9]+", "test")
	if err == nil {
		t.Error(fmt.Errorf("an error was expected"))
	}

	testCache = testMap["internal"]

	err = areMatching(testMap["internal"], testMap["internal"])
	if err != nil {
		t.Error(fmt.Errorf("error while processing templates: %v", err))
	}

	err = areMatching(testMap["internal"], nil)
	if err == nil {
		t.Error(fmt.Errorf("an error was expected"))
	}

	type Vertex2 struct {
		X int
		Y int
	}

	type Vertex3 struct {
		X int
		Y int
		Z int
	}

	v2 := Vertex2{3, 5}
	v3 := Vertex3{3, 5, 7}
	v4 := Vertex3{3, 11, 7}

	// pointers

	err = areMatching(&v2, &v2)
	if err != nil {
		t.Error(fmt.Errorf("error while processing templates: %v", err))
	}

	err = areMatching(&v3, &v3)
	if err != nil {
		t.Error(fmt.Errorf("error while processing templates: %v", err))
	}

	// struct

	err = areMatching(v2, v2)
	if err != nil {
		t.Error(fmt.Errorf("error while processing templates: %v", err))
	}

	err = areMatching(v2, v3)
	if err != nil {
		t.Error(fmt.Errorf("error while processing templates: %v", err))
	}

	err = areMatching(v3, v2)
	if err == nil {
		t.Error(fmt.Errorf("an error was expected"))
	}

	err = areMatching(v3, v4)
	if err == nil {
		t.Error(fmt.Errorf("an error was expected"))
	}

	// slice

	s2 := []int{3, 5}
	s3 := []int{3, 5, 7}
	s4 := []int{3, 11, 7}

	err = areMatching(s2, s2)
	if err != nil {
		t.Error(fmt.Errorf("error while processing templates: %v", err))
	}

	err = areMatching(s3, s2)
	if err == nil {
		t.Error(fmt.Errorf("an error was expected"))
	}

	err = areMatching(s3, s4)
	if err == nil {
		t.Error(fmt.Errorf("an error was expected"))
	}

	// map

	m2 := map[string]int{"a": 3, "b": 5}
	m3 := map[string]int{"a": 3, "b": 5, "c": 7}
	m4 := map[string]int{"a": 3, "b": 11, "c": 7}

	err = areMatching(m2, m2)
	if err != nil {
		t.Error(fmt.Errorf("error while processing templates: %v", err))
	}

	err = areMatching(m3, m2)
	if err == nil {
		t.Error(fmt.Errorf("an error was expected"))
	}

	err = areMatching(m3, m4)
	if err == nil {
		t.Error(fmt.Errorf("an error was expected"))
	}

	// interfaces

	err = areMatching(reflect.ValueOf(m3).Interface(), reflect.ValueOf(m2).Interface())
	if err == nil {
		t.Error(fmt.Errorf("an error was expected"))
	}

	var v interface{}
	err = areMatching(v, v)
	if err == nil {
		t.Error(fmt.Errorf("an error was expected"))
	}
}
