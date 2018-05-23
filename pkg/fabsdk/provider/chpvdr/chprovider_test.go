// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chpvdr

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/chconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/stretchr/testify/assert"
)

type mockClientContext struct {
	context.Providers
	msp.SigningIdentity
}

func TestBasicValidChannel(t *testing.T) {
	ctx := mocks.NewMockProviderContext()

	user := mspmocks.NewMockSigningIdentity("user", "user")

	clientCtx := &mockClientContext{
		Providers:       ctx,
		SigningIdentity: user,
	}

	cp, err := New(clientCtx.EndpointConfig())
	if err != nil {
		t.Fatalf("Unexpected error creating Channel Provider: %v", err)
	}

	err = cp.Initialize(ctx)
	assert.NoError(t, err)

	mockChConfigCache := newMockChCfgCache(chconfig.NewChannelCfg(""))
	mockChConfigCache.Put(chconfig.NewChannelCfg("mychannel"))
	cp.chCfgCache = mockChConfigCache

	channelService, err := cp.ChannelService(clientCtx, "mychannel")
	if err != nil {
		t.Fatalf("Unexpected error creating Channel Service: %v", err)
	}

	// System channel
	channelService, err = cp.ChannelService(clientCtx, "")
	if err != nil {
		t.Fatalf("Unexpected error creating Channel Service: %v", err)
	}

	m, err := channelService.Membership()
	assert.Nil(t, err)
	assert.NotNil(t, m)

	chConfig, err := channelService.Config()
	assert.Nil(t, err)
	assert.NotNil(t, chConfig)

	channelConfig, err := channelService.ChannelConfig()
	assert.Nil(t, err)
	assert.NotNil(t, channelConfig)
}

func TestResolveEventServiceType(t *testing.T) {
	ctx := mocks.NewMockContext(mspmocks.NewMockSigningIdentity("test", "Org1MSP"))
	chConfig := mocks.NewMockChannelCfg("mychannel")

	useDeliver, err := useDeliverEvents(ctx, chConfig)
	assert.NoError(t, err)
	assert.Falsef(t, useDeliver, "expecting deliver events not to be used")

	chConfig.MockCapabilities[fab.ApplicationGroupKey][fab.V1_1Capability] = true

	useDeliver, err = useDeliverEvents(ctx, chConfig)
	assert.NoError(t, err)
	assert.Truef(t, useDeliver, "expecting deliver events to be used")
}
