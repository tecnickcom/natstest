package main

import (
	"errors"

	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

// params struct contains the application parameters
type remoteConfigParams struct {
	remoteConfigProvider      string // remote configuration source ("consul", "etcd")
	remoteConfigEndpoint      string // remote configuration URL (ip:port)
	remoteConfigPath          string // remote configuration path where to search fo the configuration file ("/config/natstest")
	remoteConfigSecretKeyring string // path to the openpgp secret keyring used to decript the remote configuration data ("/etc/natstest/configkey.gpg")
}

// isEmpty returns true if all the fields are empty strings
func (rcfg remoteConfigParams) isEmpty() bool {
	return rcfg.remoteConfigProvider == "" && rcfg.remoteConfigEndpoint == "" && rcfg.remoteConfigPath == "" && rcfg.remoteConfigSecretKeyring == ""
}

// params struct contains the application parameters
type params struct {
	serverAddress  string   // HTTP API URL (ip:port) or just (:port)
	natsAddress    string   // NATS bus Address (nats://ip:port)
	validTransfCmd []string // list of valid transformation commands
	logLevel       string   // Log level: NONE, EMERGENCY, ALERT, CRITICAL, ERROR, WARNING, NOTICE, INFO, DEBUG
}

// getConfigParams returns the configuration parameters
func getConfigParams() (params, error) {
	cfg, rcfg := getLocalConfigParams()
	return getRemoteConfigParams(cfg, rcfg)
}

// getLocalConfigParams returns the local configuration parameters
func getLocalConfigParams() (cfg params, rcfg remoteConfigParams) {

	viper.Reset()

	// set default remote configuration values
	viper.SetDefault("remoteConfigProvider", RemoteConfigProvider)
	viper.SetDefault("remoteConfigEndpoint", RemoteConfigEndpoint)
	viper.SetDefault("remoteConfigPath", RemoteConfigPath)
	viper.SetDefault("remoteConfigSecretKeyring", RemoteConfigSecretKeyring)
	// set default configuration values
	viper.SetDefault("serverAddress", ServerAddress)
	viper.SetDefault("natsAddress", NatsAddress)
	viper.SetDefault("validTransfCmd", ValidTransfCmd)
	viper.SetDefault("logLevel", LogLevel)

	// name of the configuration file without extension
	viper.SetConfigName("config")

	// configuration type
	viper.SetConfigType("json")

	// add local configuration paths
	for _, cpath := range ConfigPath {
		viper.AddConfigPath(cpath)
	}

	// Find and read the local configuration file (if any)
	viper.ReadInConfig()

	// read configuration parameters
	cfg = params{
		serverAddress:  viper.GetString("serverAddress"),
		natsAddress:    viper.GetString("natsAddress"),
		validTransfCmd: viper.GetStringSlice("validTransfCmd"),
		logLevel:       viper.GetString("logLevel"),
	}

	// support environment variables for the remote configuration
	viper.AutomaticEnv()
	viper.SetEnvPrefix(ServiceName) // will be uppercased automatically
	viper.BindEnv("remoteConfigProvider")
	viper.BindEnv("remoteConfigEndpoint")
	viper.BindEnv("remoteConfigPath")
	viper.BindEnv("remoteConfigSecretKeyring")

	rcfg = remoteConfigParams{
		remoteConfigProvider:      viper.GetString("remoteConfigProvider"),
		remoteConfigEndpoint:      viper.GetString("remoteConfigEndpoint"),
		remoteConfigPath:          viper.GetString("remoteConfigPath"),
		remoteConfigSecretKeyring: viper.GetString("remoteConfigSecretKeyring"),
	}

	return cfg, rcfg
}

// getRemoteConfigParams returns the remote configuration parameters
func getRemoteConfigParams(cfg params, rcfg remoteConfigParams) (params, error) {

	if rcfg.isEmpty() {
		return cfg, nil
	}

	viper.Reset()

	// set default configuration values
	viper.SetDefault("serverAddress", cfg.serverAddress)
	viper.SetDefault("natsAddress", cfg.natsAddress)
	viper.SetDefault("validTransfCmd", cfg.validTransfCmd)
	viper.SetDefault("logLevel", cfg.logLevel)

	// configuration type
	viper.SetConfigType("json")

	// add remote configuration provider
	var err error
	if rcfg.remoteConfigSecretKeyring == "" {
		err = viper.AddRemoteProvider(rcfg.remoteConfigProvider, rcfg.remoteConfigEndpoint, rcfg.remoteConfigPath)
	} else {
		err = viper.AddSecureRemoteProvider(rcfg.remoteConfigProvider, rcfg.remoteConfigEndpoint, rcfg.remoteConfigPath, rcfg.remoteConfigSecretKeyring)
	}
	if err == nil {
		// try to read the remote configuration (if any)
		err = viper.ReadRemoteConfig()
	}
	if err != nil {
		return cfg, err
	}

	// read configuration parameters
	return params{
			serverAddress:  viper.GetString("serverAddress"),
			natsAddress:    viper.GetString("natsAddress"),
			validTransfCmd: viper.GetStringSlice("validTransfCmd"),
			logLevel:       viper.GetString("logLevel"),
		},
		nil
}

// checkParams cheks if the configuration parameters are valid
func checkParams(appParams *params) error {
	if appParams.serverAddress == "" {
		return errors.New("serverAddress is empty")
	}
	if appParams.natsAddress == "" {
		return errors.New("natsAddress is empty")
	}
	if appParams.logLevel == "" {
		return errors.New("logLevel is empty")
	}
	levelNum, ok := logLevelCodes[appParams.logLevel]
	if !ok {
		return errors.New("The logLevel must be one of the following: NONE, EMERGENCY, ALERT, CRITICAL, ERROR, WARNING, NOTICE, INFO, DEBUG")
	}
	logLevelCode = levelNum
	return nil
}
