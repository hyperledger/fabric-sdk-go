/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chpvdr

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/chconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/fabpvdr"
	"github.com/stretchr/testify/assert"
)

func TestBasicValidChannel(t *testing.T) {
	ctx := mocks.NewMockProviderContext()
	pf := &MockProviderFactory{}

	user := mocks.NewMockUser("user")

	fp, err := pf.CreateFabricProvider(ctx)
	if err != nil {
		t.Fatalf("Unexpected error creating Fabric Provider: %v", err)
	}

	cp, err := New(fp)
	if err != nil {
		t.Fatalf("Unexpected error creating Channel Provider: %v", err)
	}

	channelService, err := cp.ChannelService(user, "mychannel")
	if err != nil {
		t.Fatalf("Unexpected error creating Channel Service: %v", err)
	}

	// System channel
	channelService, err = cp.ChannelService(user, "")
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
}

// CustomFabricProvider overrides channel config default implementation
type MockFabricProvider struct {
	*fabpvdr.FabricProvider
	providerContext context.Providers
}

// CreateChannelConfig initializes the channel config
func (f *MockFabricProvider) CreateChannelConfig(ic fab.IdentityContext, channelID string) (fab.ChannelConfig, error) {

	ctx := chconfig.Context{
		Providers: f.providerContext,
		Identity:  ic,
	}

	return mocks.NewMockChannelConfig(ctx, "mychannel")

}

// CreateFabricProvider mocks new default implementation of fabric primitives
func (f *MockProviderFactory) CreateFabricProvider(context context.Providers) (fab.InfraProvider, error) {
	fabProvider := fabpvdr.New(context)

	cfp := MockFabricProvider{
		FabricProvider:  fabProvider,
		providerContext: context,
	}
	return &cfp, nil
}
