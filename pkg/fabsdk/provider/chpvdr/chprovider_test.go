/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chpvdr

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/fabpvdr"
	"github.com/stretchr/testify/assert"
)

type mockClientContext struct {
	context.Providers
	msp.Identity
}

func TestBasicValidChannel(t *testing.T) {
	ctx := mocks.NewMockProviderContext()

	user := mocks.NewMockUser("user")

	clientCtx := &mockClientContext{
		Providers: ctx,
		Identity:  user,
	}

	pf := &MockProviderFactory{ctx: clientCtx}

	fp, err := pf.CreateInfraProvider(ctx.Config())
	if err != nil {
		t.Fatalf("Unexpected error creating Fabric Provider: %v", err)
	}

	cp, err := New(fp)
	if err != nil {
		t.Fatalf("Unexpected error creating Channel Provider: %v", err)
	}

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
}

// MockProviderFactory is configured to retrieve channel config from orderer
type MockProviderFactory struct {
	defcore.ProviderFactory
	ctx context.Client
}

// MockInfraProvider overrides channel config default implementation
type MockInfraProvider struct {
	*fabpvdr.InfraProvider
	providerContext context.Providers
	ctx             context.Client
}

// CreateChannelConfig initializes the channel config
func (f *MockInfraProvider) CreateChannelConfig(channelID string) (fab.ChannelConfig, error) {
	return mocks.NewMockChannelConfig(f.ctx, "mychannel")

}

// CreateInfraProvider mocks new default implementation of fabric primitives
func (f *MockProviderFactory) CreateInfraProvider(config core.Config) (fab.InfraProvider, error) {
	fabProvider := fabpvdr.New(config)

	cfp := MockInfraProvider{
		InfraProvider: fabProvider,
		ctx:           f.ctx,
	}
	return &cfp, nil
}
