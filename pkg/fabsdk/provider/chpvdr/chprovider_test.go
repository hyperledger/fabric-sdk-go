/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chpvdr

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	channelImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/chconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/fabpvdr"
	"github.com/pkg/errors"
)

func TestBasicValidChannel(t *testing.T) {
	ctx := mocks.NewMockProviderContext()
	pf := &MockProviderFactory{}

	user := mocks.NewMockUser("user")

	fp, err := pf.NewFabricProvider(ctx)
	if err != nil {
		t.Fatalf("Unexpected error creating Fabric Provider: %v", err)
	}

	cp, err := New(fp)
	if err != nil {
		t.Fatalf("Unexpected error creating Channel Provider: %v", err)
	}

	channelService, err := cp.NewChannelService(user, "mychannel")
	if err != nil {
		t.Fatalf("Unexpected error creating Channel Service: %v", err)
	}

	_, err = channelService.Channel()
	if err != nil {
		t.Fatalf("Unexpected error creating Channel Service: %v", err)
	}

	// System channel
	channelService, err = cp.NewChannelService(user, "")
	if err != nil {
		t.Fatalf("Unexpected error creating Channel Service: %v", err)
	}

	_, err = channelService.Channel()
	if err != nil {
		t.Fatalf("Unexpected error creating Channel Service: %v", err)
	}
}

// MockProviderFactory is configured to retrieve channel config from orderer
type MockProviderFactory struct {
	defcore.ProviderFactory
}

// CustomFabricProvider overrides channel config default implementation
type MockFabricProvider struct {
	*fabpvdr.FabricProvider
	providerContext apifabclient.ProviderContext
}

// CreateChannelConfig initializes the channel config
func (f *MockFabricProvider) CreateChannelConfig(ic apifabclient.IdentityContext, channelID string) (apifabclient.ChannelConfig, error) {

	ctx := chconfig.Context{
		ProviderContext: f.providerContext,
		IdentityContext: ic,
	}

	return mocks.NewMockChannelConfig(ctx, "mychannel")

}

// CreateChannelClient overrides the default.
func (f *MockFabricProvider) CreateChannelClient(ic apifabclient.IdentityContext, cfg apifabclient.ChannelCfg) (apifabclient.Channel, error) {
	ctx := chconfig.Context{
		ProviderContext: f.providerContext,
		IdentityContext: ic,
	}
	channel, err := channelImpl.New(ctx, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "NewChannel failed")
	}

	return channel, nil
}

// NewFabricProvider mocks new default implementation of fabric primitives
func (f *MockProviderFactory) NewFabricProvider(context apifabclient.ProviderContext) (api.FabricProvider, error) {
	fabProvider := fabpvdr.New(context)

	cfp := MockFabricProvider{
		FabricProvider:  fabProvider,
		providerContext: context,
	}
	return &cfp, nil
}
