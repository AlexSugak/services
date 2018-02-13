package cli

import (
	"time"

	scli "github.com/skycoin/skycoin/src/api/cli"
	"github.com/skycoin/skycoin/src/api/webrpc"
	"github.com/skycoin/skycoin/src/visor"
	"github.com/skycoin/skycoin/src/wallet"
)

// TransactionType represents enum SPLIT | MERGE
type TransactionType int

// TransactionDetails - short summary of a confirmed transaction
type TransactionDetails struct {
	ID       string
	Type     TransactionType
	From     []string
	To       []scli.SendAmount
	Start    time.Time
	End      time.Time
	Duration time.Duration
	Status   *visor.TransactionStatus
}

const (
	// Split transaction type means sending SKY from one address to multiple other addresses
	Split TransactionType = iota + 1
	// Merge transaction type means sending SKY from multiple addresses to one
	Merge
)

// SendTransaction creates a new raw transaction, injects it and waits for a confirmation status
func SendTransaction(rpc *webrpc.Client, wallet *wallet.Wallet, txType TransactionType, from []string, to []string, amount uint64) (TransactionDetails, error) {
	var t TransactionDetails
	cliTo := toSendAmount(to, amount)
	start := time.Now()

	tx, err := scli.CreateRawTx(rpc, wallet, from, from[0], cliTo)
	if err != nil {
		return t, err
	}

	txID, err := rpc.InjectTransaction(tx)
	if err != nil {
		return t, err
	}

	status, err := waitForStatus(rpc, txID)
	if err != nil {
		return t, err
	}

	end := time.Now()
	duration := time.Since(start)

	t = TransactionDetails{
		ID:       txID,
		Type:     txType,
		From:     from,
		To:       cliTo,
		Start:    start,
		End:      end,
		Duration: duration,
		Status:   status,
	}

	return t, nil

}

func toSendAmount(addr []string, amount uint64) []scli.SendAmount {
	to := make([]scli.SendAmount, len(addr))
	for i := range addr {
		to[i] = scli.SendAmount{
			Addr:  addr[i],
			Coins: amount / uint64(len(addr)),
		}
	}

	return to
}

func waitForStatus(rpc *webrpc.Client, txID string) (*visor.TransactionStatus, error) {
	var (
		tx  *webrpc.TxnResult
		err error
	)

	for {
		if tx, err = rpc.GetTransactionByID(txID); err != nil {
			return nil, err
		}

		if tx.Transaction.Status.Confirmed {
			return &tx.Transaction.Status, nil
		}

		<-time.After(time.Second * 1)
	}
}
