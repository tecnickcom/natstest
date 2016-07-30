package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestLoadTestMapError1(t *testing.T) {
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

	err := loadTestMap()
	if err == nil {
		t.Error(fmt.Errorf("an error was expected"))
	}
}

func loadTestMapErrorTesting(t *testing.T, mode os.FileMode) {
	tmpdir := "../target/tmp/"
	os.RemoveAll(tmpdir)
	os.MkdirAll(tmpdir, 0700)
	defer os.RemoveAll(tmpdir)
	data := []byte("[{")
	file := tmpdir + "test_@error.json"
	err := ioutil.WriteFile(file, data, 0644)
	if err != nil {
		t.Error(fmt.Errorf("unable to write the file: %s -- %v", file, err))
	}
	err = os.Chmod(file, mode)
	if err != nil {
		t.Error(fmt.Errorf("unable to set the mode %v to %s -- %v", mode, file, err))
	}
	oldTestMap := testMap
	testMap = nil
	oldCfg := ConfigPath
	ConfigPath[0] = "../target/tmp/"
	ConfigPath[1] = "wrong/path/1/"
	ConfigPath[2] = "wrong/path/2/"
	ConfigPath[3] = "wrong/path/3/"
	ConfigPath[4] = "wrong/path/4/"
	defer func() {
		ConfigPath = oldCfg
		testMap = oldTestMap
	}()

	err = loadTestMap()
	if err == nil {
		t.Error(fmt.Errorf("an error was expected"))
	}
}

func TestLoadTestMapErrorA(t *testing.T) {
	loadTestMapErrorTesting(t, 0644)
}

func TestLoadTestMapErrorB(t *testing.T) {
	loadTestMapErrorTesting(t, 0200)
}
