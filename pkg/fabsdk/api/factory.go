/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicore"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
	txn "github.com/hyperledger/fabric-sdk-go/api/apitxn"
	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"
)

// CoreProviderFactory allows overriding of primitives and the fabric core object provider
type CoreProviderFactory interface {
	NewStateStoreProvider(config apiconfig.Config) (fab.KeyValueStore, error)
	NewCryptoSuiteProvider(config apiconfig.Config) (apicryptosuite.CryptoSuite, error)
	NewSigningManager(cryptoProvider apicryptosuite.CryptoSuite, config apiconfig.Config) (fab.SigningManager, error)
	NewFabricProvider(config apiconfig.Config, stateStore fab.KeyValueStore, cryptoSuite apicryptosuite.CryptoSuite, signer fab.SigningManager) (apicore.FabricProvider, error)
}

// ServiceProviderFactory allows overriding default service providers (such as peer discovery)
type ServiceProviderFactory interface {
	NewDiscoveryProvider(config apiconfig.Config) (fab.DiscoveryProvider, error)
	NewSelectionProvider(config apiconfig.Config) (fab.SelectionProvider, error)
}

// OrgClientFactory allows overriding default clients and providers of an organization
// Currently, a context is created for each organization that the client app needs.
type OrgClientFactory interface {
	//NewMSPClient(orgName string, config apiconfig.Config, cryptoProvider apicryptosuite.CryptoSuite) (fabca.FabricCAClient, error)
	NewCredentialManager(orgName string, config apiconfig.Config, cryptoProvider apicryptosuite.CryptoSuite) (fab.CredentialManager, error)
}

// SessionClientFactory allows overriding default clients and providers of a session
type SessionClientFactory interface {
	NewChannelMgmtClient(sdk Providers, session Session, config apiconfig.Config) (chmgmt.ChannelMgmtClient, error)
	NewResourceMgmtClient(sdk Providers, session Session, config apiconfig.Config, filter resmgmt.TargetFilter) (resmgmt.ResourceMgmtClient, error)
	NewChannelClient(sdk Providers, session Session, config apiconfig.Config, channelID string) (txn.ChannelClient, error)
}

// PkgSuite provides the package factories that create clients and providers
type PkgSuite interface {
	Core() (CoreProviderFactory, error)
	Service() (ServiceProviderFactory, error)
	Context() (OrgClientFactory, error)
	Session() (SessionClientFactory, error)
	Logger() (apilogging.LoggerProvider, error)
}
