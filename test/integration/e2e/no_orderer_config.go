/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package e2e

import (
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/chclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

func runWithNoOrdererConfigFixture(t *testing.T) {
	runWithNoOrdererConfig(t, config.FromFile("../"+integration.ConfigChBlockTestFile))
}

// RunWithNoOrdererConfig enables chclient scenarios using config and sdk options provided
func runWithNoOrdererConfig(t *testing.T, configOpt apiconfig.ConfigProvider, sdkOpts ...fabsdk.Option) {

	sdk, err := fabsdk.New(configOpt, sdkOpts...)
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	// ************ Test setup complete ************** //

	//Discovery filter added to test discovery filter behavior
	discoveryFilter := &mockDiscoveryFilter{called: false}
	// Channel client is used to query and execute transactions (Org1 is default org)
	chClient, err := sdk.NewClient(fabsdk.WithUser("User1")).Channel(channelID, fabsdk.WithTargetFilter(discoveryFilter))
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	// Release all channel client resources
	defer chClient.Close()

	response, err := chClient.Query(chclient.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()})
	if err != nil {
		t.Fatalf("Failed to query funds: %s", err)
	}
	value := response.Payload

	//Test if discovery filter is being called
	if !discoveryFilter.called {
		t.Fatalf("discoveryFilter not called")
	}

	eventID := "test([a-zA-Z]+)"

	// Register chaincode event (pass in channel which receives event details when the event is complete)
	notifier := make(chan *chclient.CCEvent)
	rce, err := chClient.RegisterChaincodeEvent(notifier, ccID, eventID)
	if err != nil {
		t.Fatalf("Failed to register cc event: %s", err)
	}

	// Move funds
	response, err = chClient.Execute(chclient.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCTxArgs()})
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	select {
	case ccEvent := <-notifier:
		t.Logf("Received CC event: %s\n", ccEvent)
	case <-time.After(time.Second * 20):
		t.Fatalf("Did NOT receive CC event for eventId(%s)\n", eventID)
	}

	// Unregister chain code event using registration handle
	err = chClient.UnregisterChaincodeEvent(rce)
	if err != nil {
		t.Fatalf("Unregister cc event failed: %s", err)
	}

	// Verify move funds transaction result
	response, err = chClient.Query(chclient.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()})
	if err != nil {
		t.Fatalf("Failed to query funds after transaction: %s", err)
	}

	valueInt, _ := strconv.Atoi(string(value))
	valueAfterInvokeInt, _ := strconv.Atoi(string(response.Payload))
	if valueInt+1 != valueAfterInvokeInt {
		t.Fatalf("Execute failed. Before: %s, after: %s", value, response.Payload)
	}
}

type mockDiscoveryFilter struct {
	called bool
}

// Accept returns true if this peer is to be included in the target list
func (df *mockDiscoveryFilter) Accept(peer apifabclient.Peer) bool {
	df.called = true
	return true
}
