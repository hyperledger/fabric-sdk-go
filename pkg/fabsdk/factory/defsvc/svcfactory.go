/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defsvc

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"

	discovery "github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/staticdiscovery"
	selection "github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/staticselection"
)

// ProviderFactory represents the default SDK provider factory for services.
type ProviderFactory struct{}

// NewProviderFactory returns the default SDK provider factory for services.
func NewProviderFactory() *ProviderFactory {
	f := ProviderFactory{}
	return &f
}

// CreateDiscoveryProvider returns a new default implementation of discovery provider
func (f *ProviderFactory) CreateDiscoveryProvider(config fab.EndpointConfig, fabPvdr fab.InfraProvider) (fab.DiscoveryProvider, error) {
	return discovery.New(config, fabPvdr)
}

// CreateSelectionProvider returns a new default implementation of selection service
func (f *ProviderFactory) CreateSelectionProvider(config fab.EndpointConfig) (fab.SelectionProvider, error) {
	return selection.New(config)
}
