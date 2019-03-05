/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package provider

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/dynamicselection"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/chpvdr"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/stretchr/testify/require"
)

func TestDynamicSelection(t *testing.T) {

	// Using shared SDK instance to increase test speed.
	testSetup := mainTestSetup
	aKey := integration.GetKeyName(t)
	bKey := integration.GetKeyName(t)
	moveTxArg := integration.ExampleCCTxArgs(aKey, bKey, "1")
	queryArg := integration.ExampleCCQueryArgs(bKey)

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

	chaincodeID := integration.GenerateExampleID(false)
	err = integration.PrepareExampleCC(sdk, fabsdk.WithUser("Admin"), testSetup.OrgID, chaincodeID)
	require.NoError(t, err, "InstallAndInstantiateExampleCC returned error")

	//prepare contexts
	org1ChannelClientContext := sdk.ChannelContext(testSetup.ChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))

	//Reset example cc keys
	integration.ResetKeys(t, org1ChannelClientContext, chaincodeID, "200", aKey, bKey)

	chClient, err := channel.New(org1ChannelClientContext)
	require.NoError(t, err, "Failed to create new channel client")

	//response, err := chClient.Query(channel.Request{ChaincodeID: chaincodeID, Fcn: "invoke", Args: queryArg},
	//	channel.WithRetry(retry.TestRetryOpts))
	response, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			b, e := chClient.Query(channel.Request{ChaincodeID: chaincodeID, Fcn: "invoke", Args: queryArg},
				channel.WithRetry(retry.TestRetryOpts))
			if e != nil {
				// return a retryable code if key/value query is nil (ie not propagated to all peers yet)
				if strings.Contains(e.Error(), "Nil amount for") {
					return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("QueryBlock returned error: %v", e), nil)
				}
			}
			return b, e
		},
	)
	require.NoError(t, err, "Failed to query funds, ccID: %s, queryArgs: [%+v]", chaincodeID, queryArg)
	value := response.(channel.Response).Payload

	// Move funds
	_, err = chClient.Execute(channel.Request{ChaincodeID: chaincodeID, Fcn: "invoke", Args: moveTxArg},
		channel.WithRetry(retry.DefaultChannelOpts))
	require.NoError(t, err, "Failed to move funds, ccID: %s, queryArgs:[%+v]", chaincodeID, moveTxArg)

	valueInt, _ := strconv.Atoi(string(value))
	verifyValue(t, chClient, queryArg, valueInt+1, chaincodeID)
}

func verifyValue(t *testing.T, chClient *channel.Client, queryArg [][]byte, expectedValue int, ccID string) {
	req := channel.Request{
		ChaincodeID: ccID,
		Fcn:         "invoke",
		Args:        queryArg,
	}

	_, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			resp, err := chClient.Query(req, channel.WithRetry(retry.DefaultChannelOpts))
			require.NoError(t, err, "query funds failed")

			// Verify that transaction changed block state
			actualValue, _ := strconv.Atoi(string(resp.Payload))
			if expectedValue != actualValue {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("ledger value didn't match expectation [%d, %d]", expectedValue, actualValue), nil)
			}
			return &actualValue, nil
		},
	)
	require.NoError(t, err, "Execute failed. Value was not updated")
}

// DynamicSelectionProviderFactory is configured with dynamic (endorser) selection provider
type DynamicSelectionProviderFactory struct {
	defsvc.ProviderFactory
}

// CreateChannelProvider returns a new default implementation of channel provider
func (f *DynamicSelectionProviderFactory) CreateChannelProvider(config fab.EndpointConfig, opts ...options.Opt) (fab.ChannelProvider, error) {
	chProvider, err := chpvdr.New(config, opts...)
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

type closable interface {
	Close()
}
