/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defsvc

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"

	discovery "github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/staticdiscovery"
	selection "github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/staticselection"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/chpvdr"
)

// ProviderFactory represents the default SDK provider factory for services.
type ProviderFactory struct{}

// NewProviderFactory returns the default SDK provider factory for services.
func NewProviderFactory() *ProviderFactory {
	f := ProviderFactory{}
	return &f
}

// CreateDiscoveryProvider returns a new default implementation of discovery provider
func (f *ProviderFactory) CreateDiscoveryProvider(config fab.EndpointConfig) (fab.DiscoveryProvider, error) {
	return discovery.New(config)
}

// CreateLocalDiscoveryProvider returns a new default implementation of the local discovery provider
func (f *ProviderFactory) CreateLocalDiscoveryProvider(config fab.EndpointConfig) (fab.LocalDiscoveryProvider, error) {
	return discovery.New(config)
}

// CreateChannelProvider returns a new default implementation of channel provider
func (f *ProviderFactory) CreateChannelProvider(config fab.EndpointConfig) (fab.ChannelProvider, error) {
	return chpvdr.New(config)
}

// CreateSelectionProvider returns a new default implementation of selection service
func (f *ProviderFactory) CreateSelectionProvider(config fab.EndpointConfig) (fab.SelectionProvider, error) {
	return selection.New(config)
}
