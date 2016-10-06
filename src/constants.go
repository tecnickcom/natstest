package main

// ProgramName contains this program name
const ProgramName = "natstest"

// ProgramVersion contains this program version
// This is automatically populated by the Makefile using the value from the VERSION file
var ProgramVersion = "0.0.0"

// ProgramRelease contains this program release number (or build number)
// This is automatically populated by the Makefile using the value from the RELEASE file
var ProgramRelease = "0"

// ServerAddress is the HTTP API URL (ip:port) or just (:port)
const ServerAddress = ":8081"

// NatsAddress is the default NATS bus address
const NatsAddress = "nats://127.0.0.1:4222"

// BusTimeout is the default NATS bus connection timeout in seconds
const BusTimeout = 1

// ConfigPath list the local paths where to look for configuration files (in order)
var ConfigPath = [...]string{
	"../resources/test/etc/" + ProgramName + "/",
	"./",
	"config/",
	"$HOME/." + ProgramName + "/",
	"/etc/" + ProgramName + "/",
}

// LogLevel defines the default log level: NONE, EMERGENCY, ALERT, CRITICAL, ERROR, WARNING, NOTICE, INFO, DEBUG
const LogLevel = "INFO"

// RemoteConfigProvider is the remote configuration source ("consul", "etcd")
const RemoteConfigProvider = ""

// RemoteConfigEndpoint is the remote configuration URL (ip:port)
const RemoteConfigEndpoint = ""

// RemoteConfigPath is the remote configuration path where to search fo the configuration file ("/config/natstest")
const RemoteConfigPath = ""

// RemoteConfigSecretKeyring is the path to the openpgp secret keyring used to decript the remote configuration data ("/etc/natstest/configkey.gpg")
const RemoteConfigSecretKeyring = ""

// ValidTransfCmd contains the default list of valid transformation commands to be used in test configuration templates
var ValidTransfCmd = []string{
	"/bin/cat",
	"/bin/echo",
}
