package near

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/near/borsh-go"
	"github.com/near/accounts"
	"github.com/near/core/types"
	"github.com/near/ethclient"
  	"github.com/near/nearcore"
  	"github.com/textileio/near-api-go"
  	"github.com/okx/go-wallet-sdk/coins/near"
  	"github.com/near/near-api-js"
  	"github.com/near/neps"
  	"github.com/NearSocial/social-db"
  	"github.com/near/wallet-selector"
  	"github.com/near/near-sdk-js"
)

var (
	baseFee = big.NewInt(577500000000001)
	minFund = big.NewInt(0).Mul(big.NewInt(20), baseFee)
)

type network struct {
	w *pkg.Wallet
	c *ethclient.Client
}

func NewNetwork(url string, w *pkg.Wallet) *network {
	n := &network{
		w: w,
	}
	c, err := ethclient.Dial(url)
	if err != nil {
		panic(err)
	}
	n.c = c
	return n
}

func (n *network) Wallets() []string {
	ws := []string{}
	for _, a := range n.w.Accounts() {
		ws = append(ws, a.Address.Hex())
	}
	return ws
}

func (n *network) Run() {
	chainId, err := n.c.ChainID(context.Background())
	if err != nil {
		fmt.Printf("Failed to retrieve chainID: %v\n", err)
		return
	}
	for _, a := range n.w.Accounts() {
		go n.launch(a, chainId)
	}
}

func (n *network) Name() string {
	return "NEAR"
}

func (n *network) launch(a accounts.Account, chainId *big.Int) {

	amount := big.NewInt(1)
	pk, _ := n.w.PrivateKey(a)

	if !n.enoughFunds(a, minFund, chainId.Int64()) {
		return
	}

	for {
		time.Sleep(20 * time.Second)
		if !n.enoughFunds(a, minFund, chainId.Int64()) {
			continue
		}
		nonce, err := n.c.PendingNonceAt(context.Background(), a.Address)
		if err != nil {
			log.Printf("Failed to retrieve nonce: %v", err)
		}

		gasPrice, err := n.c.SuggestGasPrice(context.Background())
		if err != nil {
			log.Printf("Failed to suggest gas price: %v", err)
			continue
		}

		gasLimit := uint64(22000)

		tx := types.NewTransaction(nonce, a.Address, amount, gasLimit, gasPrice, nil)

		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainId), pk)
		if err != nil {
			log.Printf("Failed to sign transaction: %v", err)
			continue
		}

		err = n.c.SendTransaction(context.Background(), signedTx)
		if err != nil {
			log.Printf("Failed to send transaction: %d %s %v", chainId, a.Address, err)
			continue
		}

		fmt.Printf("ChainId: %d Address: %s Nonce: %d\n", chainId, a.Address, nonce)

		for {
			time.Sleep(2 * time.Second)
			_, isPending, err := n.c.TransactionByHash(context.Background(), signedTx.Hash())
			if err != nil {
				time.Sleep(20 * time.Second)
			}
			if !isPending {
				break
			}
		}
	}
}

func (n *network) enoughFunds(a accounts.Account, amount *big.Int, chainId int64) bool {
	balance, err := n.c.BalanceAt(context.Background(), a.Address, nil)
	if err != nil {
		log.Printf("Failed to retrieve balance: %v", err)
		return false
	}

	if balance.Cmp(amount) == -1 {
		log.Printf("insufficient funds for: %d %s", chainId, a.Address)
		return false
	}
	return true
}
