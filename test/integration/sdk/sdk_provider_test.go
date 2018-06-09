/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/test/integration"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/chpvdr"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/dynamicselection"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/stretchr/testify/require"
)

func TestDynamicSelection(t *testing.T) {

	// Using shared SDK instance to increase test speed.
	testSetup := mainTestSetup

	// Create SDK setup for channel client with dynamic selection
	sdk, err := fabsdk.New(integration.ConfigBackend,
		fabsdk.WithServicePkg(&DynamicSelectionProviderFactory{}))

	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}
	defer sdk.Close()

	if err = testSetup.Initialize(sdk); err != nil {
		t.Fatal(err)
	}

	chainCodeID := integration.GenerateRandomID()
	resp, err := integration.InstallAndInstantiateExampleCC(sdk, fabsdk.WithUser("Admin"), testSetup.OrgID, chainCodeID)
	require.Nil(t, err, "InstallAndInstantiateExampleCC return error")
	require.NotEmpty(t, resp, "instantiate response should be populated")

	//prepare contexts
	org1ChannelClientContext := sdk.ChannelContext(testSetup.ChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))

	chClient, err := channel.New(org1ChannelClientContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	response, err := chClient.Query(channel.Request{ChaincodeID: chainCodeID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()},
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		t.Fatalf("Failed to query funds: %s", err)
	}
	value := response.Payload

	// Move funds
	response, err = chClient.Execute(channel.Request{ChaincodeID: chainCodeID, Fcn: "invoke", Args: integration.ExampleCCTxArgs()},
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	valueInt, _ := strconv.Atoi(string(value))
	success := false
	for i := 0; i < 5; i++ {
		// Verify move funds transaction result
		response, err = chClient.Query(channel.Request{ChaincodeID: chainCodeID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()})
		if err != nil {
			t.Fatalf("Failed to query funds after transaction: %s", err)
		}

		valueAfterInvokeInt, _ := strconv.Atoi(string(response.Payload))
		if valueInt+1 == valueAfterInvokeInt {
			success = true
			break
		}
		t.Logf("Execute failed. Before: %s, after: %s", value, response.Payload)
		time.Sleep(2 * time.Second)
	}
	require.Truef(t, success, "Execute failed. Value was not updated")
}

// DynamicSelectionProviderFactory is configured with dynamic (endorser) selection provider
type DynamicSelectionProviderFactory struct {
	defsvc.ProviderFactory
}

// CreateChannelProvider returns a new default implementation of channel provider
func (f *DynamicSelectionProviderFactory) CreateChannelProvider(config fab.EndpointConfig) (fab.ChannelProvider, error) {
	chProvider, err := chpvdr.New(config)
	if err != nil {
		return nil, err
	}
	return &dynamicSelectionChannelProvider{
		ChannelProvider: chProvider,
		services:        make(map[string]*dynamicselection.SelectionService),
	}, nil
}

type dynamicSelectionChannelProvider struct {
	fab.ChannelProvider
	services map[string]*dynamicselection.SelectionService
}

type initializer interface {
	Initialize(providers context.Providers) error
}

// Initialize sets the provider context
func (cp *dynamicSelectionChannelProvider) Initialize(providers context.Providers) error {
	if init, ok := cp.ChannelProvider.(initializer); ok {
		init.Initialize(providers)
	}
	return nil
}

type closable interface {
	Close()
}

// Close frees resources and caches.
func (cp *dynamicSelectionChannelProvider) Close() {
	if c, ok := cp.ChannelProvider.(closable); ok {
		c.Close()
	}

	for _, service := range cp.services {
		service.Close()
	}
}

// ChannelService creates a ChannelService
func (cp *dynamicSelectionChannelProvider) ChannelService(ctx fab.ClientContext, channelID string) (fab.ChannelService, error) {
	chService, err := cp.ChannelProvider.ChannelService(ctx, channelID)
	if err != nil {
		return nil, err
	}

	selection, ok := cp.services[channelID]
	if !ok {
		discovery, err := chService.Discovery()
		if err != nil {
			return nil, err
		}
		selection, err := dynamicselection.NewService(ctx, channelID, discovery)
		if err != nil {
			return nil, err
		}
		cp.services[channelID] = selection
	}

	return &dynamicSelectionChannelService{
		ChannelService: chService,
		selection:      selection,
	}, nil
}

type dynamicSelectionChannelService struct {
	fab.ChannelService
	selection fab.SelectionService
}

func (cs *dynamicSelectionChannelService) Selection() (fab.SelectionService, error) {
	return cs.selection, nil
}
