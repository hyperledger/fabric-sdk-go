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
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/fabpvdr"
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

	pf := &MockProviderFactory{ctx: clientCtx}

	fp, err := pf.CreateInfraProvider(ctx.EndpointConfig())
	if err != nil {
		t.Fatalf("Unexpected error creating Fabric Provider: %v", err)
	}

	cp, err := New(fp)
	if err != nil {
		t.Fatalf("Unexpected error creating Channel Provider: %v", err)
	}

	_, err = cp.ChannelService(clientCtx, "mychannel")
	if err != nil {
		t.Fatalf("Unexpected error creating Channel Service: %v", err)
	}

	// System channel
	channelService, err := cp.ChannelService(clientCtx, "")
	if err != nil {
		t.Fatalf("Unexpected error creating Channel Service: %v", err)
	}

	m, err := channelService.Membership()
	assert.Nil(t, err)
	assert.NotNil(t, m)
}

// MockProviderFactory is configured to retrieve channel config from orderer
type MockProviderFactory struct {
	defcore.ProviderFactory
	ctx context.Client
}

// MockInfraProvider overrides channel config default implementation
type MockInfraProvider struct {
	*fabpvdr.InfraProvider
	ctx context.Client
}

// CreateChannelConfig initializes the channel config
func (f *MockInfraProvider) CreateChannelConfig(channelID string) (fab.ChannelConfig, error) {
	return mocks.NewMockChannelConfig(f.ctx, "mychannel")
}

func (f *MockInfraProvider) CreateChannelCfg(ctx fab.ClientContext, channel string) (fab.ChannelCfg, error) {
	return mocks.NewMockChannelCfg(channel), nil
}

func (f *MockInfraProvider) CreateChannelMembership(ctx fab.ClientContext, channel string) (fab.ChannelMembership, error) {
	return mocks.NewMockMembership(), nil
}

// CreateInfraProvider mockcore new default implementation of fabric primitives
func (f *MockProviderFactory) CreateInfraProvider(config fab.EndpointConfig) (fab.InfraProvider, error) {
	fabProvider := fabpvdr.New(config)

	cfp := MockInfraProvider{
		InfraProvider: fabProvider,
		ctx:           f.ctx,
	}
	return &cfp, nil
}
