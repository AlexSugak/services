package cli

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/skycoin/skycoin/src/api/webrpc"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/wallet"
	gcli "github.com/urfave/cli"
)

const (
	// default number of addresses to generate
	defaultN = 10
	// default wait time (in secs)
	defaultWait = 10
)

var (
	ErrSeedNotSpecified = errors.New("seed value is requried but was not specified")
	ErrNotEnoughCoins   = errors.New("could not find an address with at leat 1 SKY")
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
	addresses := cipherAddrToString(wallet.GenerateAddresses(uint64(N)))
	fmt.Println("done")

	sigs := *SigsFromContext(c)
	rpc := *RPCClientFromContext(c)

	// start by finding an address with at least 1 SKY
	from, amount, err := findAddrWitBalance(&rpc, addresses, 1e6)
	if err != nil {
		return err
	}

	fmt.Println("Running tests...")
	t := Split
	var to []string

transactions:
	for {
		if t == Split {
			// split by sending coins from one address to two random addresses
			to = getRandomAddresses(addresses, 2)
		} else if t == Merge {
			// merge by sending coins from two addresses to one random address
			to = getRandomAddresses(addresses, 1)
		}

		_, err := SendTransaction(&rpc, wallet, t, from, to, amount)
		if err != nil {
			return err
		}
		// TODO: print transaction details

		// use TO addresses as FROM for the next iteration
		from = to

		select {
		case <-time.After(time.Second * time.Duration(Wait)):
			continue
		case <-sigs:
			fmt.Println("quiting")
			break transactions
		}
	}

	if Cleanup {
		fmt.Println("cleaning up")
		to = []string{
			addresses[0],
		}
		// cleanup by sending all coints to the first generated address
		_, err := SendTransaction(&rpc, wallet, Merge, addresses, to, amount)
		if err != nil {
			return err
		}
		// TODO: print transaction details
	}

	// TODO: print test summary

	return nil
}

func createWallet(name string, seed string) (*wallet.Wallet, error) {
	return wallet.NewWallet(name, wallet.Options{
		Coin:  wallet.CoinTypeSkycoin,
		Label: name,
		Seed:  seed,
	})
}

// findAddrWitBalance tryes to find an address with at least minimum specified amount of coins on balance
func findAddrWitBalance(client *webrpc.Client, addrs []string, minBalance uint64) ([]string, uint64, error) {
	var (
		amount uint64
		addr   string
	)
	unspent, err := client.GetUnspentOutputs(addrs)
	if err != nil {
		return nil, amount, err
	}

	spendable, err := unspent.Outputs.SpendableOutputs().ToUxArray()
	if err != nil {
		return nil, amount, err
	} else if len(spendable) == 0 {
		return nil, amount, ErrNotEnoughCoins
	}

	for i := range spendable {
		if spendable[i].Body.Coins > amount {
			amount = spendable[i].Body.Coins
			if amount >= minBalance {
				addr = spendable[i].Body.Address.String()
				break
			}
		}
	}

	if amount < minBalance {
		return nil, amount, ErrNotEnoughCoins
	}

	return []string{addr}, amount, nil
}

func getRandomAddresses(addrs []string, n int) []string {
	res := make([]string, n)
	len := len(addrs)
	i := 0
	for i < n {
		j := rand.Intn(len)
		// if result does not contain new random address, add it and increment index
		if f := index(res, addrs[j]); f == -1 {
			res[i] = addrs[j]
			i++
		}
	}

	return res
}

func index(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}

	return -1
}

func cipherAddrToString(addr []cipher.Address) []string {
	res := make([]string, len(addr))
	for i, addr := range addr {
		res[i] = addr.String()
	}

	return res
}
