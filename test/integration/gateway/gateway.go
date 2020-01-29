/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
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

	nw, err := gw.GetNetwork(channelID)
	if err != nil {
		t.Fatalf("Failed to get network: %s", err)
	}

	name := nw.GetName()
	if name != channelID {
		t.Fatalf("Incorrect network name: %s", name)
	}

	contract := nw.GetContract(ccID)

	name = contract.GetName()
	if name != ccID {
		t.Fatalf("Incorrect contract name: %s", name)
	}

	runContract(contract, t)
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

	nw, err := gw.GetNetwork(channelID)
	if err != nil {
		t.Fatalf("Failed to get network: %s", err)
	}

	name := nw.GetName()
	if name != channelID {
		t.Fatalf("Incorrect network name: %s", name)
	}

	contract := nw.GetContract(ccID)

	name = contract.GetName()
	if name != ccID {
		t.Fatalf("Incorrect contract name: %s", name)
	}

	runContract(contract, t)
}

func runContract(contract gateway.Contract, t *testing.T) {
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
