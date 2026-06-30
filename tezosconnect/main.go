package main

import (
	"fmt"
	"os"

	"github.com/hyperledger-firefly/tezosconnect/cmd"
)

func main() {
	cmd.InitConfig()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
