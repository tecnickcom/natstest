package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/viper"
)

func getTestCfgParams() *params {
	return &params{
		log: &LogData{
			Level:   "INFO",
			Network: "",
			Address: "",
		},
		stats: &StatsData{
			Prefix:      "natstest-test",
			Network:     "udp",
			Address:     ":8125",
			FlushPeriod: 100,
		},
		serverAddress:  ":8081",
		natsAddress:    "nats://127.0.0.1:4222",
		validTransfCmd: []string{"/bin/cat", "/bin/echo"},
	}
}

func TestCheckParams(t *testing.T) {
	err := checkParams(getTestCfgParams())
	if err != nil {
		t.Error(fmt.Errorf("No errors are expected: %v", err))
	}
}

func TestCheckParamsErrorsLogLevelEmpty(t *testing.T) {
	cfg := getTestCfgParams()
	cfg.log.Level = ""
	err := checkParams(cfg)
	if err == nil {
		t.Error(fmt.Errorf("An error was expected because logLevel is empty"))
	}
}

func TestCheckParamsErrorsLogLevelInvalid(t *testing.T) {
	cfg := getTestCfgParams()
	cfg.log.Level = "INVALID"
	err := checkParams(cfg)
	if err == nil {
		t.Error(fmt.Errorf("An error was expected because logLevel is invalid"))
	}
}

func TestCheckParamsErrorsStatsPrefix(t *testing.T) {
	cfg := getTestCfgParams()
	cfg.stats.Prefix = ""
	err := checkParams(cfg)
	if err == nil {
		t.Error(fmt.Errorf("An error was expected because the stats Prefix is empty"))
	}
}

func TestCheckParamsErrorsStatsNetwork(t *testing.T) {
	cfg := getTestCfgParams()
	cfg.stats.Network = ""
	err := checkParams(cfg)
	if err == nil {
		t.Error(fmt.Errorf("An error was expected because the stats Network is empty"))
	}
}

func TestCheckParamsErrorsStatsFlushPeriod(t *testing.T) {
	cfg := getTestCfgParams()
	cfg.stats.FlushPeriod = -1
	err := checkParams(cfg)
	if err == nil {
		t.Error(fmt.Errorf("An error was expected because the stats FlushPeriod is negative"))
	}
}

func TestCheckParamsErrorsServerAddress(t *testing.T) {
	cfg := getTestCfgParams()
	cfg.serverAddress = ""
	err := checkParams(cfg)
	if err == nil {
		t.Error(fmt.Errorf("An error was expected because serverAddress is empty"))
	}
}

func TestCheckParamsErrorsNatsAddress(t *testing.T) {
	cfg := getTestCfgParams()
	cfg.natsAddress = ""
	err := checkParams(cfg)
	if err == nil {
		t.Error(fmt.Errorf("An error was expected because natsAddress is empty"))
	}
}

func TestGetConfigParams(t *testing.T) {
	prm, err := getConfigParams()
	if err != nil {
		t.Error(fmt.Errorf("An error was not expected: %v", err))
	}
	if prm.serverAddress != ":8081" {
		t.Error(fmt.Errorf("Found different server address than expected, found %s", prm.serverAddress))
	}
	if prm.log.Level != "DEBUG" {
		t.Error(fmt.Errorf("Found different logLevel than expected, found %s", prm.log.Level))
	}
}

func TestGetLocalConfigParams(t *testing.T) {

	// test environment variables
	defer unsetRemoteConfigEnv()
	os.Setenv("NATSTEST_REMOTECONFIGPROVIDER", "consul")
	os.Setenv("NATSTEST_REMOTECONFIGENDPOINT", "127.0.0.1:98765")
	os.Setenv("NATSTEST_REMOTECONFIGPATH", "/config/natstest")
	os.Setenv("NATSTEST_REMOTECONFIGSECRETKEYRING", "")

	prm, rprm := getLocalConfigParams()

	if prm.serverAddress != ":8081" {
		t.Error(fmt.Errorf("Found different server address than expected, found %s", prm.serverAddress))
	}
	if prm.log.Level != "DEBUG" {
		t.Error(fmt.Errorf("Found different logLevel than expected, found %s", prm.log.Level))
	}
	if rprm.remoteConfigProvider != "consul" {
		t.Error(fmt.Errorf("Found different remoteConfigProvider than expected, found %s", rprm.remoteConfigProvider))
	}
	if rprm.remoteConfigEndpoint != "127.0.0.1:98765" {
		t.Error(fmt.Errorf("Found different remoteConfigEndpoint than expected, found %s", rprm.remoteConfigEndpoint))
	}
	if rprm.remoteConfigPath != "/config/natstest" {
		t.Error(fmt.Errorf("Found different remoteConfigPath than expected, found %s", rprm.remoteConfigPath))
	}
	if rprm.remoteConfigSecretKeyring != "" {
		t.Error(fmt.Errorf("Found different remoteConfigSecretKeyring than expected, found %s", rprm.remoteConfigSecretKeyring))
	}

	_, err := getRemoteConfigParams(prm, rprm)
	if err == nil {
		t.Error(fmt.Errorf("A remote configuration error was expected"))
	}

	rprm.remoteConfigSecretKeyring = "/etc/natstest/cfgkey.gpg"
	_, err = getRemoteConfigParams(prm, rprm)
	if err == nil {
		t.Error(fmt.Errorf("A remote configuration error was expected"))
	}
}

// Test real Consul provider
// To activate this define the environmental variable NATSTEST_LIVECONSUL
func TestGetConfigParamsRemote(t *testing.T) {

	enable := os.Getenv("NATSTEST_LIVECONSUL")
	if enable == "" {
		return
	}

	// test environment variables
	defer unsetRemoteConfigEnv()
	os.Setenv("NATSTEST_REMOTECONFIGPROVIDER", "consul")
	os.Setenv("NATSTEST_REMOTECONFIGENDPOINT", "127.0.0.1:8500")
	os.Setenv("NATSTEST_REMOTECONFIGPATH", "/config/natstest")
	os.Setenv("NATSTEST_REMOTECONFIGSECRETKEYRING", "")

	// load a specific config file just for testing
	oldCfg := ConfigPath
	viper.Reset()
	for k := range ConfigPath {
		ConfigPath[k] = "wrong/path/"
	}
	defer func() { ConfigPath = oldCfg }()

	prm, err := getConfigParams()
	if err != nil {
		t.Error(fmt.Errorf("An error was not expected: %v", err))
	}
	if prm.serverAddress != ":8123" {
		t.Error(fmt.Errorf("Found different server address than expected, found %s", prm.serverAddress))
	}
	if prm.log.Level != "debug" {
		t.Error(fmt.Errorf("Found different logLevel than expected, found %s", prm.log.Level))
	}
}

func TestCliWrongConfigError(t *testing.T) {

	// test environment variables
	defer unsetRemoteConfigEnv()
	os.Setenv("NATSTEST_REMOTECONFIGPROVIDER", "consul")
	os.Setenv("NATSTEST_REMOTECONFIGENDPOINT", "127.0.0.1:999999")
	os.Setenv("NATSTEST_REMOTECONFIGPATH", "/config/wrong")
	os.Setenv("NATSTEST_REMOTECONFIGSECRETKEYRING", "")

	// load a specific config file just for testing
	oldCfg := ConfigPath
	viper.Reset()
	for k := range ConfigPath {
		ConfigPath[k] = "wrong/path/"
	}
	defer func() { ConfigPath = oldCfg }()

	_, err := cli()
	if err == nil {
		t.Error(fmt.Errorf("An error was expected"))
		return
	}
}

// unsetRemoteConfigEnv clear the environmental variables used to set the remote configuration
func unsetRemoteConfigEnv() {
	os.Setenv("NATSTEST_REMOTECONFIGPROVIDER", "")
	os.Setenv("NATSTEST_REMOTECONFIGENDPOINT", "")
	os.Setenv("NATSTEST_REMOTECONFIGPATH", "")
	os.Setenv("NATSTEST_REMOTECONFIGSECRETKEYRING", "")
}
