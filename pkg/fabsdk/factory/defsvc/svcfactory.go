/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defsvc

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"

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
func (f *ProviderFactory) CreateDiscoveryProvider(config core.Config) (fab.DiscoveryProvider, error) {
	return discovery.New(config)
}

// CreateSelectionProvider returns a new default implementation of selection service
func (f *ProviderFactory) CreateSelectionProvider(config core.Config) (fab.SelectionProvider, error) {
	return selection.New(config)
}
