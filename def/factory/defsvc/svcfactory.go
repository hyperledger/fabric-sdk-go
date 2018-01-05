/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defsvc

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"

	discovery "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/discovery/staticdiscovery"
	selection "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/selection/staticselection"
)

// ProviderFactory represents the default SDK provider factory for services.
type ProviderFactory struct{}

// NewProviderFactory returns the default SDK provider factory for services.
func NewProviderFactory() *ProviderFactory {
	f := ProviderFactory{}
	return &f
}

// NewDiscoveryProvider returns a new default implementation of discovery provider
func (f *ProviderFactory) NewDiscoveryProvider(config apiconfig.Config) (fab.DiscoveryProvider, error) {
	return discovery.NewDiscoveryProvider(config)
}

// NewSelectionProvider returns a new default implementation of selection service
func (f *ProviderFactory) NewSelectionProvider(config apiconfig.Config) (fab.SelectionProvider, error) {
	return selection.NewSelectionProvider(config)
}
