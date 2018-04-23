// +build devstable

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/stretchr/testify/require"

	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/dynamicdiscovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

func TestDynamicDiscovery(t *testing.T) {
	testSetup := mainTestSetup

	// Create SDK setup for channel client with dynamic selection
	sdk, err := fabsdk.New(config.FromFile("../../fixtures/config/config_test.yaml"),
		fabsdk.WithServicePkg(&DynamicDiscoveryProviderFactory{}))
	require.NoError(t, err, "Failed to create new SDK")
	defer sdk.Close()

	err = testSetup.Initialize(sdk)
	require.NoError(t, err, "Failed to initialize test setup")

	chProvider := sdk.ChannelContext(testSetup.ChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))
	chCtx, err := chProvider()
	require.NoError(t, err, "Error creating channel context")

	peers, err := chCtx.DiscoveryService().GetPeers()
	require.NoErrorf(t, err, "Error getting peers for channel [%s]", testSetup.ChannelID)
	require.NotEmptyf(t, peers, "No peers were found for channel [%s]", testSetup.ChannelID)

	t.Logf("Peers of channel [%s]:", testSetup.ChannelID)
	for _, p := range peers {
		t.Logf("- [%s] - MSP [%s]", p.URL(), p.MSPID())
	}
}

func TestDynamicLocalDiscovery(t *testing.T) {
	testSetup := mainTestSetup

	// Create SDK setup for channel client with dynamic selection
	sdk, err := fabsdk.New(config.FromFile("../../fixtures/config/config_test.yaml"),
		fabsdk.WithServicePkg(&DynamicDiscoveryProviderFactory{}))
	require.NoError(t, err, "Failed to create new SDK")
	defer sdk.Close()

	err = testSetup.Initialize(sdk)
	require.NoError(t, err, "Failed to initialize test setup")

	// By default, query for local peers (outside of a channel) requires admin privileges.
	// To bypass this restriction, set peer.discovery.orgMembersAllowedAccess=true in core.yaml.
	ctxProvider := sdk.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1Name))

	locCtx, err := contextImpl.NewLocal(ctxProvider)
	require.NoError(t, err, "Error creating local context")

	peers, err := locCtx.LocalDiscoveryService().GetPeers()
	require.NoErrorf(t, err, "Error getting local peers for MSP [%s]", locCtx.Identifier().MSPID)
	require.NotEmptyf(t, peers, "No local peers were found for MSP [%s]", locCtx.Identifier().MSPID)

	t.Logf("Local peers for MSP [%s]:", locCtx.Identifier().MSPID)
	for _, p := range peers {
		t.Logf("- [%s] - MSP [%s]", p.URL(), p.MSPID())
	}
}

// DynamicDiscoveryProviderFactory is configured with dynamic (endorser) selection provider
type DynamicDiscoveryProviderFactory struct {
	defsvc.ProviderFactory
}

// CreateDiscoveryProvider returns a new dynamic discovery provider
func (f *DynamicDiscoveryProviderFactory) CreateDiscoveryProvider(config fab.EndpointConfig) (fab.DiscoveryProvider, error) {
	return dynamicdiscovery.New(config), nil
}

// CreateLocalDiscoveryProvider returns a new local dynamic discovery provider
func (f *DynamicDiscoveryProviderFactory) CreateLocalDiscoveryProvider(config fab.EndpointConfig) (fab.LocalDiscoveryProvider, error) {
	return dynamicdiscovery.New(config), nil
}
