/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
)

// Providers represents the SDK configured providers context.
type Providers interface {
	core.Providers
	fab.Providers
}

// CoreProviderFactory allows overriding of primitives and the fabric core object provider
type CoreProviderFactory interface {
	CreateStateStoreProvider(config core.Config) (core.KVStore, error)
	CreateCryptoSuiteProvider(config core.Config) (core.CryptoSuite, error)
	CreateSigningManager(cryptoProvider core.CryptoSuite, config core.Config) (core.SigningManager, error)
	CreateIdentityManager(orgName string, stateStore core.KVStore, cryptoProvider core.CryptoSuite, config core.Config) (core.IdentityManager, error)
	CreateInfraProvider(config core.Config) (fab.InfraProvider, error)
}

// ServiceProviderFactory allows overriding default service providers (such as peer discovery)
type ServiceProviderFactory interface {
	CreateDiscoveryProvider(config core.Config, fabPvdr fab.InfraProvider) (fab.DiscoveryProvider, error)
	CreateSelectionProvider(config core.Config) (fab.SelectionProvider, error)
}
