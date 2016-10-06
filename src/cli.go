package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func cli() (*cobra.Command, error) {

	// configuration parameters
	cfgParams, err := getConfigParams()
	if err != nil {
		return nil, err
	}

	// application parameters
	appParams := new(params)

	// set the root command
	rootCmd := new(cobra.Command)

	// overwrites the configuration parameters with the ones specified in the command line (if any)
	rootCmd.Flags().StringVarP(&appParams.serverAddress, "serverAddress", "s", cfgParams.serverAddress, "HTTP API URL (ip:port) or just (:port)")
	rootCmd.Flags().StringVarP(&appParams.natsAddress, "natsAddress", "n", cfgParams.natsAddress, "NATS bus Address (nats://ip:port)")
	rootCmd.Flags().StringVarP(&appParams.logLevel, "logLevel", "l", cfgParams.logLevel, "Log level: NONE, EMERGENCY, ALERT, CRITICAL, ERROR, WARNING, NOTICE, INFO, DEBUG")

	for _, cmd := range cfgParams.validTransfCmd {
		isValidTransfCmd[cmd] = true
	}

	rootCmd.Use = "natstest"
	rootCmd.Short = "NATS Test Component"
	rootCmd.Long = `NATS Test Component`
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		// check values
		err := checkParams(appParams)
		if err != nil {
			return err
		}
		// load the test map from the test configuration files
		err = loadTestMap()
		if err != nil {
			return err
		}
		// initialize the NATS bus
		initNatsBus(appParams.natsAddress)
		// start the http server
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
