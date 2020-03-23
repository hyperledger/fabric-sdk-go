/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

const (
	channelID = "mychannel"
)

var (
	ccID = "example_cc_e2e" + metadata.TestRunID
)

// RunWithConfig the basic gateway integration test
func RunWithConfig(t *testing.T) {
	configPath := integration.GetConfigPath("config_e2e.yaml")

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(configPath)),
		gateway.WithUser("User1"),
	)

	if err != nil {
		t.Fatalf("Failed to create new Gateway: %s", err)
	}
	defer gw.Close()

	testGateway(gw, t)
}

// RunWithSDK the sdk compatibility gateway integration test
func RunWithSDK(t *testing.T) {
	configPath := integration.GetConfigPath("config_e2e.yaml")

	sdk, err := fabsdk.New(config.FromFile(configPath))

	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	gw, err := gateway.Connect(
		gateway.WithSDK(sdk),
		gateway.WithUser("User1"),
	)

	if err != nil {
		t.Fatalf("Failed to create new Gateway: %s", err)
	}
	defer gw.Close()

	testGateway(gw, t)
}

// RunWithWallet gateway/wallet integration test
func RunWithWallet(t *testing.T) {
	wallet := gateway.NewInMemoryWallet()
	err := populateWallet(wallet)
	if err != nil {
		t.Fatalf("Failed to populate wallet: %s", err)
	}

	configPath := integration.GetConfigPath("config_e2e.yaml")

	sdk, err := fabsdk.New(config.FromFile(configPath))

	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	gw, err := gateway.Connect(
		gateway.WithSDK(sdk),
		gateway.WithIdentity(wallet, "User1"),
	)

	if err != nil {
		t.Fatalf("Failed to create new Gateway: %s", err)
	}
	defer gw.Close()

	testGateway(gw, t)
}

// RunWithTransient tests sending transient data
func RunWithTransient(t *testing.T) {
	configPath := integration.GetConfigPath("config_e2e.yaml")

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(configPath)),
		gateway.WithUser("User1"),
	)
	if err != nil {
		t.Fatalf("Failed to create new Gateway: %s", err)
	}
	defer gw.Close()

	nw, err := gw.GetNetwork(channelID)
	if err != nil {
		t.Fatalf("Failed to get network: %s", err)
	}

	contract := nw.GetContract(ccID)
	testTransientData(contract, t)
}

func testGateway(gw *gateway.Gateway, t *testing.T) {
	nw, err := gw.GetNetwork(channelID)
	if err != nil {
		t.Fatalf("Failed to get network: %s", err)
	}

	name := nw.Name()
	if name != channelID {
		t.Fatalf("Incorrect network name: %s", name)
	}

	contract := nw.GetContract(ccID)

	name = contract.Name()
	if name != ccID {
		t.Fatalf("Incorrect contract name: %s", name)
	}

	runContract(contract, t)
}

func runContract(contract *gateway.Contract, t *testing.T) {
	response, err := contract.EvaluateTransaction("invoke", "query", "b")

	if err != nil {
		t.Fatalf("Failed to query funds: %s", err)
	}

	value, _ := strconv.Atoi(string(response))

	_, err = contract.SubmitTransaction("invoke", "move", "a", "b", "1")

	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	time.Sleep(10 * time.Second)

	response, err = contract.EvaluateTransaction("invoke", "query", "b")

	if err != nil {
		t.Fatalf("Failed to query funds: %s", err)
	}

	newvalue, _ := strconv.Atoi(string(response))

	if newvalue != value+1 {
		t.Fatalf("Incorrect response: %s", response)
	}
}

func testTransientData(contract *gateway.Contract, t *testing.T) {
	transient := make(map[string][]byte)
	transient["result"] = []byte("8500")

	txn, err := contract.CreateTransaction("invoke", gateway.WithTransient(transient))
	if err != nil {
		t.Fatalf("Failed to create transaction: %s", err)
	}

	result, err := txn.Submit("move", "a", "b", "1")
	if err != nil {
		t.Fatalf("Failed to submit transaction: %s", err)
	}

	if string(result) != "8500" {
		t.Fatalf("Incorrect result: %s", string(result))
	}
}

func populateWallet(wallet *gateway.Wallet) error {
	credPath := filepath.Join(
		metadata.GetProjectPath(),
		metadata.CryptoConfigPath,
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "User1@org1.example.com-cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return errors.New("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	err = wallet.Put("User1", identity)
	if err != nil {
		return err
	}
	return nil
}
