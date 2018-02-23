/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/chclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
)

// CoreProviderFactory allows overriding of primitives and the fabric core object provider
type CoreProviderFactory interface {
	NewStateStoreProvider(config core.Config) (api.KVStore, error)
	NewCryptoSuiteProvider(config core.Config) (core.CryptoSuite, error)
	NewSigningManager(cryptoProvider core.CryptoSuite, config core.Config) (api.SigningManager, error)
	NewFabricProvider(context context.ProviderContext) (FabricProvider, error)
}

// ServiceProviderFactory allows overriding default service providers (such as peer discovery)
type ServiceProviderFactory interface {
	NewDiscoveryProvider(config core.Config) (fab.DiscoveryProvider, error)
	NewSelectionProvider(config core.Config) (fab.SelectionProvider, error)
	//	NewChannelProvider(ctx Context, channelID string) (ChannelProvider, error)
}

// OrgClientFactory allows overriding default clients and providers of an organization
// Currently, a context is created for each organization that the client app needs.
type OrgClientFactory interface {
	//NewMSPClient(orgName string, config apiconfig.Config, cryptoProvider apicryptosuite.CryptoSuite) (fabca.FabricCAClient, error)
	NewCredentialManager(orgName string, config core.Config, cryptoProvider core.CryptoSuite) (api.CredentialManager, error)
}

// SessionClientFactory allows overriding default clients and providers of a session
type SessionClientFactory interface {
	NewChannelClient(sdk Providers, session context.SessionContext, channelID string, targetFilter fab.TargetFilter) (*chclient.ChannelClient, error)
}
