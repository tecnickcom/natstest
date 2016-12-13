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
	log            *LogData   // Log level: EMERGENCY, ALERT, CRITICAL, ERROR, WARNING, NOTICE, INFO, DEBUG.
	stats          *StatsData // StatsD configuration, it is used to collect usage metrics
	serverAddress  string     // HTTP API URL (ip:port) or just (:port)
	natsAddress    string     // NATS bus Address (nats://ip:port)
	validTransfCmd []string   // list of valid transformation commands
}

var configDir string
var appParams = new(params)

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

	viper.SetDefault("log.level", LogLevel)
	viper.SetDefault("log.network", LogNetwork)
	viper.SetDefault("log.address", LogAddress)

	viper.SetDefault("stats.prefix", StatsPrefix)
	viper.SetDefault("stats.network", StatsNetwork)
	viper.SetDefault("stats.address", StatsAddress)
	viper.SetDefault("stats.flush_period", StatsFlushPeriod)

	viper.SetDefault("serverAddress", ServerAddress)
	viper.SetDefault("natsAddress", NatsAddress)
	viper.SetDefault("validTransfCmd", ValidTransfCmd)

	// name of the configuration file without extension
	viper.SetConfigName("config")

	// configuration type
	viper.SetConfigType("json")

	if configDir != "" {
		viper.AddConfigPath(configDir)
	}

	// add local configuration paths
	for _, cpath := range ConfigPath {
		viper.AddConfigPath(cpath)
	}

	// Find and read the local configuration file (if any)
	viper.ReadInConfig()

	// read configuration parameters
	cfg = getViperParams()

	// support environment variables for the remote configuration
	viper.AutomaticEnv()
	viper.SetEnvPrefix(ProgramName) // will be uppercased automatically
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

	viper.SetDefault("log.level", cfg.log.Level)
	viper.SetDefault("log.network", cfg.log.Network)
	viper.SetDefault("log.address", cfg.log.Address)

	viper.SetDefault("stats.prefix", cfg.stats.Prefix)
	viper.SetDefault("stats.network", cfg.stats.Network)
	viper.SetDefault("stats.address", cfg.stats.Address)
	viper.SetDefault("stats.flush_period", cfg.stats.FlushPeriod)

	viper.SetDefault("serverAddress", cfg.serverAddress)
	viper.SetDefault("natsAddress", cfg.natsAddress)
	viper.SetDefault("validTransfCmd", cfg.validTransfCmd)

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
	return getViperParams(), nil
}

// getViperParams reads the config params via Viper
func getViperParams() params {
	return params{

		log: &LogData{
			Level:   viper.GetString("log.level"),
			Network: viper.GetString("log.network"),
			Address: viper.GetString("log.address"),
		},

		stats: &StatsData{
			Prefix:      viper.GetString("stats.prefix"),
			Network:     viper.GetString("stats.network"),
			Address:     viper.GetString("stats.address"),
			FlushPeriod: viper.GetInt("stats.flush_period"),
		},

		serverAddress:  viper.GetString("serverAddress"),
		natsAddress:    viper.GetString("natsAddress"),
		validTransfCmd: viper.GetStringSlice("validTransfCmd"),
	}
}

// checkParams cheks if the configuration parameters are valid
func checkParams(prm *params) error {
	// Log
	if prm.log.Level == "" {
		return errors.New("log Level is empty")
	}
	err := prm.log.setLog()
	if err != nil {
		return err
	}

	// StatsD
	if prm.stats.Prefix == "" {
		return errors.New("The stats Prefix is empty")
	}
	if prm.stats.Network != "udp" && prm.stats.Network != "tcp" {
		return errors.New("The stats Network must be udp or tcp")
	}
	if prm.stats.FlushPeriod < 0 {
		return errors.New("The stats FlushPeriod must be >= 0")
	}

	if prm.serverAddress == "" {
		return errors.New("serverAddress is empty")
	}
	if prm.natsAddress == "" {
		return errors.New("natsAddress is empty")
	}

	return nil
}
