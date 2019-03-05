/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
)

// Providers represents the SDK configured providers context.
type Providers interface {
	core.Providers
	msp.Providers
	fab.Providers
}

// CoreProviderFactory allows overriding of primitives and the fabric core object provider
type CoreProviderFactory interface {
	CreateCryptoSuiteProvider(config core.CryptoSuiteConfig) (core.CryptoSuite, error)
	CreateSigningManager(cryptoProvider core.CryptoSuite) (core.SigningManager, error)
	CreateInfraProvider(config fab.EndpointConfig) (fab.InfraProvider, error)
}

// MSPProviderFactory allows overriding providers of MSP services
type MSPProviderFactory interface {
	CreateUserStore(config msp.IdentityConfig) (msp.UserStore, error)
	CreateIdentityManagerProvider(config fab.EndpointConfig, cryptoProvider core.CryptoSuite, userStore msp.UserStore) (msp.IdentityManagerProvider, error)
}

// ServiceProviderFactory allows overriding default service providers (such as peer discovery)
type ServiceProviderFactory interface {
	CreateLocalDiscoveryProvider(config fab.EndpointConfig) (fab.LocalDiscoveryProvider, error)
	CreateChannelProvider(config fab.EndpointConfig, opts ...options.Opt) (fab.ChannelProvider, error)
}
