/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package e2e

import (
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

// runWithNoOrdererConfig enables chclient scenarios using config and sdk options provided
func runWithNoOrdererConfig(t *testing.T, configOpt core.ConfigProvider, sdkOpts ...fabsdk.Option) {

	if integration.IsLocal() {
		//If it is a local test then add entity mapping to config backend to parse URLs
		configOpt = integration.AddLocalEntityMapping(configOpt, integration.LocalOrdererPeersConfig)
	}

	sdk, err := fabsdk.New(configOpt, sdkOpts...)
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}
	defer sdk.Close()

	// Delete all private keys from the crypto suite store
	// and users from the user store at the end
	integration.CleanupUserData(t, sdk)
	defer integration.CleanupUserData(t, sdk)

	// ************ Test setup complete ************** //

	//TODO : discovery filter should be fixed
	discoveryFilter := &mockDiscoveryFilter{called: false}

	//prepare channel client context using client context
	clientChannelContext := sdk.ChannelContext(channelID, fabsdk.WithUser("User1"), fabsdk.WithOrg(orgName))

	// Channel client is used to query and execute transactions (Org1 is default org)
	client, err := channel.New(clientChannelContext)

	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	response, err := client.Query(
		channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()},
		channel.WithTargetFilter(discoveryFilter))
	if err != nil {
		t.Fatalf("Failed to query funds: %s", err)
	}
	value := response.Payload

	//Test if discovery filter is being called
	if !discoveryFilter.called {
		t.Fatal("discoveryFilter not called")
	}

	eventID := "test([a-zA-Z]+)"

	// Register chaincode event (pass in channel which receives event details when the event is complete)
	reg, notifier, err := client.RegisterChaincodeEvent(ccID, eventID)
	if err != nil {
		t.Fatalf("Failed to register cc event: %s", err)
	}
	defer client.UnregisterChaincodeEvent(reg)

	// Move funds
	moveFunds(response, client, t, notifier, eventID, value)
}

func moveFunds(response channel.Response, client *channel.Client, t *testing.T, notifier <-chan *fab.CCEvent, eventID string, value []byte) {
	response, err := client.Execute(channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCTxArgs()})
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}
	select {
	case ccEvent := <-notifier:
		t.Logf("Received CC event: %#v\n", ccEvent)
	case <-time.After(time.Second * 20):
		t.Fatalf("Did NOT receive CC event for eventId(%s)\n", eventID)
	}
	// Verify move funds transaction result
	response, err = client.Query(channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()})
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
func (df *mockDiscoveryFilter) Accept(peer fab.Peer) bool {
	df.called = true
	return true
}
