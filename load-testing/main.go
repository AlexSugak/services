package main

import (
	"fmt"
	"github.com/skycoin/services/load-testing/cli"
	"os"
)

func main() {
	config, err := test.LoadConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	app := test.NewApp(config)

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
