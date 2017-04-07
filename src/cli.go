package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func cli() (*cobra.Command, error) {

	var logLevel string
	var serverAddress string
	var natsAddress string

	rootCmd := new(cobra.Command)
	rootCmd.Flags().StringVarP(&configDir, "configDir", "c", "", "Configuration directory to be added on top of the search list")
	rootCmd.Flags().StringVarP(&logLevel, "logLevel", "o", "*", "Log level: EMERGENCY, ALERT, CRITICAL, ERROR, WARNING, NOTICE, INFO, DEBUG")
	rootCmd.Flags().StringVarP(&serverAddress, "serverAddress", "s", "*", "HTTP API URL (ip:port) or just (:port)")
	rootCmd.Flags().StringVarP(&natsAddress, "natsAddress", "n", "*", "NATS bus Address (nats://ip:port)")
	err := rootCmd.ParseFlags(os.Args)
	if err != nil {
		return nil, err
	}

	rootCmd.Use = "natstest"
	rootCmd.Short = "NATS Test Component"
	rootCmd.Long = `NATS Test Component`
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {

		// configuration parameters
		cfgParams, err := getConfigParams()
		if err != nil {
			return err
		}
		appParams = &cfgParams
		if logLevel != "*" {
			appParams.log.Level = logLevel
		}
		if serverAddress != "*" {
			appParams.serverAddress = serverAddress
		}
		if natsAddress != "*" {
			appParams.natsAddress = natsAddress
		}

		for _, cmd := range cfgParams.validTransfCmd {
			isValidTransfCmd[cmd] = true
		}

		// check values
		err = checkParams(appParams)
		if err != nil {
			return err
		}

		// initialize StatsD client
		err = initStats(appParams.stats)
		if err == nil {
			defer stats.Close()
		}

		// load the test map from the test configuration files
		err = loadTestMap()
		if err != nil {
			return err
		}

		// initialize the NATS bus
		initNatsBus(appParams.natsAddress)

		// start the HTTP server
		return startServer(appParams.serverAddress)
	}

	// sub-command to print the version
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "print this program version",
		Long:  `print this program version`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(ProgramVersion)
		},
	}
	rootCmd.AddCommand(versionCmd)

	return rootCmd, nil
}
