package test

import (
	"errors"
	"fmt"
	gcli "github.com/urfave/cli"
	"time"
)

const (
	// default number of addresses to generate
	defaultN = 10
	// default wait time (in secs)
	defaultWait = 10
)

var (
	ErrSeedNotSpecified = errors.New("Seed value is requried but was not specified")
)

func randomTxsCmd() gcli.Command {
	name := "random"
	return gcli.Command{
		Name:  name,
		Usage: "Creates N random addresses and executes a series of random MERGE or SPLIT transactions, optionally cleaning up the result by sending all remaining coins to one address",
		Flags: []gcli.Flag{
			gcli.StringFlag{
				Name:  "seed, s",
				Usage: "seed to generate skycoin addresses",
			},
			gcli.IntFlag{
				Name:  "n",
				Usage: "number of test addresses to generate",
				Value: defaultN,
			},
			gcli.IntFlag{
				Name:  "wait, w",
				Usage: "time, in secs, to wait between load test iterations",
				Value: defaultWait,
			},
			gcli.BoolFlag{
				Name:  "cleanup, c",
				Usage: "forces cleanup of the load test results by creating a new transaction to send all coins from generated addresses to the first generated address",
			},
		},
		Action:       randomTxs,
		OnUsageError: onCommandUsageError(name),
	}
}

func randomTxs(c *gcli.Context) error {
	//rpc := RpcClientFromContext(c)

	N := c.Int("n")
	if N == 0 {
		N = defaultN
	}
	Cleanup := c.Bool("c")
	Wait := c.Int("w")
	if Wait == 0 {
		Wait = defaultWait
	}
	Seed := c.String("s")
	if Seed == "" {
		return ErrSeedNotSpecified
	}

	fmt.Printf("Starting random transactions load test, with %d test addresses generated with %s seed and %d seconds wait time between transactions\n", N, Seed, Wait)

	sigs := *SigsFromContext(c)

transactions:
	for {
		fmt.Println("starting job at", time.Now())
		time.Sleep(time.Second * 2)
		fmt.Println("done job at", time.Now())

		select {
		case <-time.After(time.Second * time.Duration(Wait)):
			fmt.Println("tick at", time.Now())
			continue
		case <-sigs:
			fmt.Println("quiting")
			break transactions
		}
	}

	if Cleanup {
		fmt.Println("cleaning up")
	}

	return nil
}
