/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

func Example() {
	// A wallet existing in the 'wallet' folder
	wallet, err := NewFileSystemWallet("wallet")
	if err != nil {
		fmt.Printf("Failed to create wallet: %s\n", err)
		os.Exit(1)
	}

	// Path to the network config (CCP) file
	ccpPath := filepath.Join(
		"..",
		"connection-org1.yaml",
	)

	// Connect to the gateway peer(s) using the network config and identity in the wallet
	gw, err := Connect(
		WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		fmt.Printf("Failed to connect to gateway: %s\n", err)
		os.Exit(1)
	}
	defer gw.Close()

	// Get the network channel 'mychannel'
	network, err := gw.GetNetwork("mychannel")
	if err != nil {
		fmt.Printf("Failed to get network: %s\n", err)
		os.Exit(1)
	}

	// Get the smart contract 'fabcar'
	contract := network.GetContract("fabcar")

	// Submit a transaction in that contract to the ledger
	result, err := contract.SubmitTransaction("createCar", "CAR10", "VW", "Polo", "Grey", "Mary")
	if err != nil {
		fmt.Printf("Failed to submit transaction: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(string(result))
}

func ExampleWithConfig() {
	// Path to the network config (CCP) file
	ccpPath := filepath.Join(
		"..",
		"connection-org1.yaml",
	)

	// Connect to the gateway peer(s) using the network config
	gw, err := Connect(
		WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		WithUser("admin"),
	)
	if err != nil {
		fmt.Printf("Failed to connect to gateway: %s\n", err)
		os.Exit(1)
	}
	defer gw.Close()
}

func ExampleWithSDK() {
	sdk, err := fabsdk.New(config.FromFile("testdata/connection-tls.json"))
	if err != nil {
		fmt.Printf("Failed to create SDK: %s", err)
	}

	gw, err := Connect(
		WithSDK(sdk),
		WithUser("user1"),
	)
	if err != nil {
		fmt.Printf("Failed to create gateway: %s", err)
	}
	defer gw.Close()
}

func ExampleWithIdentity() {
	// A wallet existing in the 'wallet' folder
	wallet, err := NewFileSystemWallet("wallet")
	if err != nil {
		fmt.Printf("Failed to create wallet: %s\n", err)
		os.Exit(1)
	}

	// Connect to the gateway peer(s) using an identity from this wallet
	gw, err := Connect(
		WithConfig(config.FromFile("testdata/connection-tls.json")),
		WithIdentity(wallet, "appUser"),
	)

	if err != nil {
		fmt.Printf("Failed to connect to gateway: %s\n", err)
		os.Exit(1)
	}
	defer gw.Close()
}

func ExampleWithUser() {
	// Connect to the gateway peer(s) using an identity defined in the network config
	gw, err := Connect(
		WithConfig(config.FromFile("testdata/connection-tls.json")),
		WithUser("user1"),
	)

	if err != nil {
		fmt.Printf("Failed to connect to gateway: %s\n", err)
		os.Exit(1)
	}
	defer gw.Close()
}

func ExampleWithTimeout() {
	// Path to the network config (CCP) file
	ccpPath := filepath.Join(
		"..",
		"connection-org1.yaml",
	)

	// Connect to the gateway peer(s) using the network config
	gw, err := Connect(
		WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		WithUser("admin"),
		WithTimeout(300*time.Second),
	)
	if err != nil {
		fmt.Printf("Failed to connect to gateway: %s\n", err)
		os.Exit(1)
	}
	defer gw.Close()

}

func ExampleNewInMemoryWallet() {
	wallet := NewInMemoryWallet()

	fmt.Println(wallet)
}

func ExampleNewFileSystemWallet() {
	walletPath := filepath.Join("..", "wallet")

	wallet, err := NewFileSystemWallet(walletPath)
	if err != nil {
		fmt.Printf("Failed to create wallet: %s\n", err)
		return
	}

	fmt.Println(wallet)
}

func myGateway() *Gateway {
	// Path to the network config (CCP) file
	ccpPath := filepath.Join(
		"..",
		"connection-org1.yaml",
	)

	// Connect to the gateway peer(s) using the network config
	gw, _ := Connect(
		WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		WithUser("admin"),
	)
	return gw
}

func myNetwork() *Network {
	gw := myGateway()
	nw, _ := gw.GetNetwork("mychannel")
	return nw
}

func myContract() *Contract {
	nw := myNetwork()
	return nw.GetContract("fabcar")
}

func runContract(c *Contract) {

}

func ExampleContract_CreateTransaction() {
	contract := myContract()

	txn, err := contract.CreateTransaction("createCar")
	if err != nil {
		fmt.Printf("Failed to create transaction: %s\n", err)
		return
	}

	result, err := txn.Submit("CAR10", "VW", "Polo", "Grey", "Mary")
	if err != nil {
		fmt.Printf("Failed to submit transaction: %s\n", err)
		return
	}

	fmt.Println(string(result))
}

func ExampleContract_EvaluateTransaction() {
	contract := myContract()

	result, err := contract.EvaluateTransaction("queryCar", "CAR01")
	if err != nil {
		fmt.Printf("Failed to evaluate transaction: %s\n", err)
		return
	}

	fmt.Println(string(result))
}

func ExampleContract_SubmitTransaction() {
	contract := myContract()

	result, err := contract.SubmitTransaction("createCar", "CAR10", "VW", "Polo", "Grey", "Mary")
	if err != nil {
		fmt.Printf("Failed to submit transaction: %s\n", err)
		return
	}

	fmt.Println(string(result))
}

func ExampleContract_RegisterEvent() {
	contract := myContract()

	eventID := "test([a-zA-Z]+)"

	reg, notifier, err := contract.RegisterEvent(eventID)
	if err != nil {
		fmt.Printf("Failed to register contract event: %s", err)
		return
	}
	defer contract.Unregister(reg)

	result, err := contract.SubmitTransaction("createCar", "CAR10", "VW", "Polo", "Grey", "Mary")
	if err != nil {
		fmt.Printf("Failed to submit transaction: %s\n", err)
		return
	}

	fmt.Println(string(result))

	var ccEvent *fab.CCEvent
	select {
	case ccEvent = <-notifier:
		fmt.Printf("Received CC event: %#v\n", ccEvent)
	case <-time.After(time.Second * 20):
		fmt.Printf("Did NOT receive CC event for eventId(%s)\n", eventID)
	}
}

func ExampleTransaction_Evaluate() {
	contract := myContract()

	txn, err := contract.CreateTransaction("queryCar")
	if err != nil {
		fmt.Printf("Failed to create transaction: %s\n", err)
		return
	}

	result, err := txn.Evaluate("CAR01")
	if err != nil {
		fmt.Printf("Failed to evaluate transaction: %s\n", err)
		return
	}

	fmt.Println(string(result))
}

func ExampleTransaction_Submit() {
	contract := myContract()

	txn, err := contract.CreateTransaction("createCar")
	if err != nil {
		fmt.Printf("Failed to create transaction: %s\n", err)
		return
	}

	result, err := txn.Submit("CAR10", "VW", "Polo", "Grey", "Mary")
	if err != nil {
		fmt.Printf("Failed to submit transaction: %s\n", err)
		return
	}

	fmt.Println(string(result))
}

func ExampleTransaction_RegisterCommitEvent() {
	contract := myContract()

	txn, err := contract.CreateTransaction("createCar")
	if err != nil {
		fmt.Printf("Failed to create transaction: %s\n", err)
		return
	}

	notifier := txn.RegisterCommitEvent()

	result, err := txn.Submit("CAR10", "VW", "Polo", "Grey", "Mary")
	if err != nil {
		fmt.Printf("Failed to submit transaction: %s\n", err)
		return
	}

	var cEvent *fab.TxStatusEvent
	select {
	case cEvent = <-notifier:
		fmt.Printf("Received commit event: %#v\n", cEvent)
	case <-time.After(time.Second * 20):
		fmt.Printf("Did NOT receive commit event\n")
	}

	fmt.Println(string(result))
}

func ExampleWithEndorsingPeers() {
	contract := myContract()

	txn, err := contract.CreateTransaction(
		"createCar",
		WithEndorsingPeers("peer1.org1.example.com:8051", "peer1.org2.example.com:10051"),
	)
	if err != nil {
		fmt.Printf("Failed to create transaction: %s\n", err)
		return
	}

	result, err := txn.Submit("CAR10", "VW", "Polo", "Grey", "Mary")
	if err != nil {
		fmt.Printf("Failed to submit transaction: %s\n", err)
		return
	}

	fmt.Println(string(result))
}

func ExampleWithTransient() {
	contract := myContract()

	transient := make(map[string][]byte)
	transient["price"] = []byte("8500")

	txn, err := contract.CreateTransaction(
		"changeCarOwner",
		WithTransient(transient),
	)
	if err != nil {
		fmt.Printf("Failed to create transaction: %s\n", err)
		return
	}

	result, err := txn.Submit("CAR10", "Archie")
	if err != nil {
		fmt.Printf("Failed to submit transaction: %s\n", err)
		return
	}

	fmt.Println(string(result))
}

func ExampleNetwork_GetContract() {
	network := myNetwork()

	contract := network.GetContract("fabcar")

	fmt.Println(contract.Name())
}

func ExampleNetwork_RegisterBlockEvent() {
	network := myNetwork()

	reg, notifier, err := network.RegisterBlockEvent()
	if err != nil {
		fmt.Printf("Failed to register block event: %s", err)
		return
	}
	defer network.Unregister(reg)

	contract := network.GetContract("fabcar")

	runContract(contract) // submit transactions

	var bEvent *fab.BlockEvent
	select {
	case bEvent = <-notifier:
		fmt.Printf("Received block event: %#v\n", bEvent)
	case <-time.After(time.Second * 20):
		fmt.Printf("Did NOT receive block event\n")
	}
}

func ExampleNetwork_RegisterFilteredBlockEvent() {
	network := myNetwork()

	reg, notifier, err := network.RegisterFilteredBlockEvent()
	if err != nil {
		fmt.Printf("Failed to register filtered block event: %s", err)
		return
	}
	defer network.Unregister(reg)

	contract := network.GetContract("fabcar")

	runContract(contract) // submit transactions

	var bEvent *fab.FilteredBlockEvent
	select {
	case bEvent = <-notifier:
		fmt.Printf("Received block event: %#v\n", bEvent)
	case <-time.After(time.Second * 20):
		fmt.Printf("Did NOT receive block event\n")
	}
}

func ExampleConnect() {
	// A wallet existing in the 'wallet' folder
	wallet, err := NewFileSystemWallet("wallet")
	if err != nil {
		fmt.Printf("Failed to create wallet: %s\n", err)
		os.Exit(1)
	}

	// Path to the network config (CCP) file
	ccpPath := filepath.Join(
		"..",
		"connection-org1.yaml",
	)

	// Connect to the gateway peer(s) using the network config and identity in the wallet
	gw, err := Connect(
		WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		fmt.Printf("Failed to connect to gateway: %s\n", err)
		os.Exit(1)
	}
	defer gw.Close()
}

func ExampleGateway_GetNetwork() {
	gw := myGateway()

	network, err := gw.GetNetwork("fabcar")
	if err != nil {
		fmt.Printf("Failed to get network: %s\n", err)
		return
	}

	fmt.Println(network.Name())
}

func ExampleWallet_Get() {
	// A wallet existing in the 'wallet' folder
	wallet, err := NewFileSystemWallet("wallet")
	if err != nil {
		fmt.Printf("Failed to create wallet: %s\n", err)
		return
	}

	id, err := wallet.Get("appUser")

	fmt.Println(id)
}

func ExampleWallet_List() {
	// A wallet existing in the 'wallet' folder
	wallet, err := NewFileSystemWallet("wallet")
	if err != nil {
		fmt.Printf("Failed to create wallet: %s\n", err)
		return
	}

	labels, err := wallet.List()

	fmt.Println(labels)
}

func ExampleWallet_Put() {
	// A new transient wallet
	wallet := NewInMemoryWallet()
	wallet.Put("testUser", NewX509Identity("Org1MSP", "--Cert PEM--", "--Key PEM--"))
}

func ExampleWallet_Remove() {
	// A wallet existing in the 'wallet' folder
	wallet, err := NewFileSystemWallet("wallet")
	if err != nil {
		fmt.Printf("Failed to create wallet: %s\n", err)
		return
	}

	wallet.Remove("appUser")
}

func ExampleNewX509Identity() {
	// create new X.509 identity
	id := NewX509Identity("Org1MSP", "--Cert PEM--", "--Key PEM--")

	// put it in a wallet
	wallet := NewInMemoryWallet()
	wallet.Put("testUser", id)
}
