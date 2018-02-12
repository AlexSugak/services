package test

import (
	"errors"
	"fmt"
	"github.com/skycoin/skycoin/src/wallet"
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
	ErrNotEnoughCoins   = errors.New("Could not find an address with at leat 1 SKY")
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

	fmt.Printf("Starting random transactions load test, with %d test addresses generated with \"%s\" seed and %d seconds wait time between transactions\n", N, Seed, Wait)

	fmt.Println("Creating test wallet and generating addresses...")
	wallet, err := createWallet("load-test-wallet", Seed)
	if err != nil {
		return err
	}
	addresses := wallet.GenerateAddresses(uint64(N))
	fmt.Println("done")

	sigs := *SigsFromContext(c)
	rpc := *RpcClientFromContext(c)

	var (
		to   []string
		from []string
		t    = Split
	)

	from = findAddrWithMaxBalance(rpc)

	fmt.Println("Running tests...")
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

func createWallet(name string, seed string) (*wallet.Wallet, error) {
	return wallet.NewWallet(name, wallet.Options{
		Coin:  wallet.CoinTypeSkycoin,
		Label: name,
		Seed:  seed,
	})
}

func findAddrWithMaxBalance(client *webrpc.Client) ([]string, error) {
	return nil, ErrNotEnoughCoins
}
