/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package e2e

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

// runWithNoOrdererConfig enables chclient scenarios using config and client options provided
func runWithNoOrdererConfig(t *testing.T, configOpt core.ConfigProvider, sdkOpts ...fabsdk.Option) {
	setupAndRun(t, false, configOpt, noOrdererE2ETest, sdkOpts...)
}

func noOrdererE2ETest(t *testing.T, sdk *fabsdk.FabricSDK) {
	//prepare channel client context using client context
	clientChannelContext := sdk.ChannelContext(channelID, fabsdk.WithUser("User1"), fabsdk.WithOrg(orgName))

	// Channel client is used to query and execute transactions (Org1 is default org)
	client, err := channel.New(clientChannelContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	value := queryCCUsingTargetFilter(t, client)

	// Move and verify funds
	ccEvent := moveFunds(t, client)
	verifyFundsIsMoved(t, client, value, ccEvent)
}

func queryCCUsingTargetFilter(t *testing.T, client *channel.Client) []byte {
	//TODO : discovery filter should be fixed
	discoveryFilter := &mockDiscoveryFilter{called: false}

	response, err := client.Query(
		channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCDefaultQueryArgs()},
		channel.WithTargetFilter(discoveryFilter),
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		t.Fatalf("Failed to query funds: %s", err)
	}

	//Test if discovery filter is being called
	if !discoveryFilter.called {
		t.Fatal("discoveryFilter not called")
	}

	return response.Payload
}

type mockDiscoveryFilter struct {
	called bool
}

// Accept returns true if this peer is to be included in the target list
func (df *mockDiscoveryFilter) Accept(peer fab.Peer) bool {
	df.called = true
	return true
}
