package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type env struct {
	NearTestnetRPC string `json:"near_testnet_rpc"`
	AxelarTestnetRPC string `json:"axelar_testnet_rpc"`
	Mnemonic string `json:"mnemonic"`
}

func main() {
	ns := []network{}
	file, err := os.Open("env.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	e := &env{}
	if err := json.Unmarshal(data, e); err != nil {
		panic(err)
	}

	w := pkg.NewWallet(e.Mnemonic)

	nearTestnet := near.NewNetwork(e.NearTestnetRPC, w)
	axelarTestnet := axelar.NewNetwork(e.AxelarTestnetRPC, w)
	ns = append(ns, nearTestnet, axelarTestnet)

	for _, n := range ns {
		fmt.Println("Fund These Addresses: ")
		fmt.Println()
		fmt.Printf("%s: \n", n.Name())
		fmt.Printf("  {\n")

		for _, a := range n.Wallets() {
			fmt.Printf("\t%s\n", a)
		}
		fmt.Printf("  }\n\n")
	}

	fmt.Print("Enter >>> ")

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	for _, n := range ns {
		go n.Run()
	}
	select {}
}
