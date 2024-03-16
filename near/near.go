package near

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"time"

	"github.com/near/borsh-go"
	"github.com/near/near/accounts"
	"github.com/near/near/core/types"
	"github.com/near/near/ethclient"
	"github.com/near/nearcore"
	"github.com/near/near-api-go"
	"github.com/okx/go-wallet-sdk/coins/near"
	"github.com/near/near-api-js"
	"github.com/near/neps"
	"github.com/NearSocial/social-db"
	"github.com/near/wallet-selector"
	"github.com/near/near-sdk-js"
)

var (
	classHash, _        = borsh.HexToFelt("0x2794ce20e5f2ff0d40e632cb53845b9f4e526ebd8471983f7dbd355b721d5a")
	ethContract  string = "0x049d36570d4e46f48e99674bd3fcc84644ddd6b96f7c741b1562b82f9e004dc7"
)

type starkAccount struct {
	acc      *accounts.Account
	ks       *accounts.MemKeystore
	pub      *borsh.Felt
	pk       *borsh.Felt
	deployed bool
}

type network struct {
	accounts map[string]starkAccount
	c        *ethclient.Client
}

type faccount struct {
	PubKey   string `json:"pub_key"`
	PrvKey   string `json:"prv_key"`
	Addreess string `json:"address"`
}

func NewNetwork(url string) *network {
	file, err := os.OpenFile("stark-accounts.json", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	accs := []faccount{}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &accs); err != nil {
			panic(err)
		}
	}
	c, err := ethclient.Dial(url)
	if err != nil {
		panic(err)
	}
	n := &network{
		accounts: make(map[string]starkAccount),
		c:        c,
	}

	if len(accs) == 0 {
		for i := 0; i < 10; i++ {
			ks, pub, pk := accounts.GetRandomKeys()
			accountAddressFelt, _ := new(borsh.Felt).SetString("0x1")
			a, err := accounts.NewAccount(c, accountAddressFelt, pub.String(), ks, 2)
			if err != nil {
				continue
			}

			precomputedAddress, err := a.PrecomputeAddress(&borsh.Zero, pub, classHash, []*borsh.Felt{pub})
			if err != nil {
				continue
			}

			a.AccountAddress = precomputedAddress
			n.accounts[precomputedAddress.String()] = starkAccount{
				acc: a,
				ks:  ks,
				pub: pub,
				pk:  pk,
			}
			accs = append(accs, faccount{
				PubKey:   pub.String(),
				PrvKey:   pk.String(),
				Addreess: precomputedAddress.String(),
			})
		}
		b, _ := json.Marshal(accs)
		_, err := file.Write(b)
		if err != nil {
			panic(err)
		}
	} else {
		for _, acc := range accs {
			addressFelt, _ := borsh.HexToFelt(acc.Addreess)
			pubFelt, _ := borsh.HexToFelt(acc.PubKey)
			prvFelt, _ := borsh.HexToFelt(acc.PrvKey)
			ks := accounts.NewMemKeystore()
			fakePrivKeyBI, ok := new(big.Int).SetString(acc.PrvKey, 0)
			if !ok {
				continue
			}
			ks.Put(acc.PubKey, fakePrivKeyBI)
			a, err := accounts.NewAccount(c, addressFelt, acc.PubKey, ks, 2)
			if err != nil {
				continue
			}
			n.accounts[acc.Addreess] = starkAccount{
				acc:      a,
				ks:       ks,
				pub:      pubFelt,
				pk:       prvFelt,
				deployed: true,
			}
		}
	}
	return n
}

func (n *network) Name() string {
	return "STARK"
}

func (n *network) Run() {
	go func() {
		for {
			for a := range n.accounts {
				n.getBalance(a)
				time.Sleep(5 * time.Second)
			}
		}
	}()
	for _, a := range n.accounts {
		if !a.deployed {
			time.Sleep(10 * time.Second)
			go func(a starkAccount) {
				for {
					time.Sleep(10 * time.Second)
					// Replace with actual deployment logic for Near network
					fmt.Println("Deploying account...")
					break
				}
				go launch(a)

			}(a)
		} else {
			go launch(a)
		}
	}
}

func launch(a starkAccount) {

	maxfee := new(borsh.Felt).SetUint64(4783000019481)
	for {
		time.Sleep(10 * time.Second)
		nonce, err := a.acc.PendingNonceAt(context.Background(), a.acc.Address)
		if err != nil {
			continue
		}
		gasPrice, err := a.c.SuggestGasPrice(context.Background())
		if err != nil {
			continue
		}
		gasLimit := uint64(22000)

		tx := types.NewTransaction(nonce, a.acc.Address, big.NewInt(1), gasLimit, gasPrice, nil)

		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainId), pk)
		if err != nil {
			continue
		}

		err = a.c.SendTransaction(context.Background(), signedTx)
		if err != nil {
			continue
		}

		fmt.Printf("Transaction sent: %s\n", signedTx.Hash().Hex())

		for {
			time.Sleep(2 * time.Second)
			_, isPending, err := a.c.TransactionByHash(context.Background(), signedTx.Hash())
			if err != nil {
				time.Sleep(20 * time.Second)
			}
			if !isPending {
				break
			}
		}
	}
}

func (n *network) getBalance(accountAddress string) {
	// Replace with actual balance retrieval logic for Near network
	fmt.Println("Retrieving balance...")
}

func (n *network) Wallets() []string {
	ws := []string{}
	for k := range n.accounts {
		ws = append(ws, k)
	}
	return ws
}
