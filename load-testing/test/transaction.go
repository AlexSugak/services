package test

import (
	"fmt"
	"github.com/skycoin/skycoin/src/api/cli"
	"github.com/skycoin/skycoin/src/api/webrpc"
	"github.com/skycoin/skycoin/src/visor"
	"github.com/skycoin/skycoin/src/wallet"
	"time"
)

type TransactionType int

type TransactionDetails struct {
	Id       string
	Type     TransactionType
	From     []string
	To       []cli.SendAmount
	Start    time.Time
	End      time.Time
	Duration time.Duration
	Status   *visor.TransactionStatus
}

const (
	Split TransactionType = iota + 1
	Merge
)

func SendTransaction(client *webrpc.Client, wallet *wallet.Wallet, txType TransactionType, from []string, to []string, amount uint64) (TransactionDetails, error) {

	cliTo := toSendAmount(to, amount)
	start := time.Now()

	tx, err := cli.CreateRawTx(client, wallet, from, from[0], cliTo)
	if err != nil {
		return nil, err
	}

	txId, err = cli.InjectTransaction(tx)
	if err != nil {
		return nil, err
	}

	status, err = waitForStatus(client, txId)
	if err != nil {
		return nil, err
	}

	end := time.Now()
	duration = time.Since(start)

	return TransactionDetails{
		Id:       txId,
		Type:     txType,
		From:     from,
		To:       cliTo,
		Start:    start,
		End:      end,
		Duration: duration,
		Status:   status,
	}, nil

}

func toSendAmount(addr []string, amount uint64) cli.SendAmount {
	to := make([]cli.SendAmount, len(addr))
	for i := range addr {
		to[i] = cli.SendAmount{addr[i], amount / uint64(len(addr))}
	}

	return to
}

func waitForStatus(client *webrpc.Client, txId string) (*visor.TransactionStatus, error) {
	var (
		tx  *webrpc.TxnResult
		err error
	)

	for {
		if tx, err = cllient.GetTransactionByID(txId); err != nil {
			return err
		}

		return &tx.Transaction.Status, nil
		<-time.After(time.Second * 1)
	}
}
