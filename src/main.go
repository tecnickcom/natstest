// Copyright (c) 2016 MIRACL UK LTD
// NATS Test Component
package main

func main() {
	rootCmd, err := cli()
	if err != nil {
		critLog.Fatalf("unable to start the service: %v", err)
	}
	// execute the root command and log errors (if any)
	if err = rootCmd.Execute(); err != nil {
		critLog.Fatalf("unable to start the service: %v", err)
	}
}
