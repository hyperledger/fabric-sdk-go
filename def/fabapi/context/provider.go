/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	fabca "github.com/hyperledger/fabric-sdk-go/api/apifabca"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	txn "github.com/hyperledger/fabric-sdk-go/api/apitxn"
	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/opt"
)

// SDKProviderFactory allows overriding default providers of an SDK
type SDKProviderFactory interface {
	NewConfigProvider(o opt.ConfigOpts, a opt.SDKOpts) (apiconfig.Config, error)
	NewStateStoreProvider(o opt.StateStoreOpts, config apiconfig.Config) (fab.KeyValueStore, error)
	NewCryptoSuiteProvider(config apiconfig.Config) (apicryptosuite.CryptoSuite, error)
	NewSigningManager(cryptoProvider apicryptosuite.CryptoSuite, config apiconfig.Config) (fab.SigningManager, error)
	NewDiscoveryProvider(config apiconfig.Config) (fab.DiscoveryProvider, error)
	NewSelectionProvider(config apiconfig.Config) (fab.SelectionProvider, error)
}

// OrgClientFactory allows overriding default clients and providers of an organization
// Currently, a context is created for each organization that the client app needs.
type OrgClientFactory interface {
	NewMSPClient(orgName string, config apiconfig.Config, cryptoProvider apicryptosuite.CryptoSuite) (fabca.FabricCAClient, error)
	NewCredentialManager(orgName string, config apiconfig.Config, cryptoProvider apicryptosuite.CryptoSuite) (fab.CredentialManager, error)
}

// SessionClientFactory allows overriding default clients and providers of a session
type SessionClientFactory interface {
	NewSystemClient(context SDK, session Session, config apiconfig.Config) (fab.FabricClient, error)
	NewChannelMgmtClient(context SDK, session Session, config apiconfig.Config) (chmgmt.ChannelMgmtClient, error)
	NewResourceMgmtClient(context SDK, session Session, config apiconfig.Config, filter resmgmt.TargetFilter) (resmgmt.ResourceMgmtClient, error)
	NewChannelClient(context SDK, session Session, config apiconfig.Config, channelID string) (txn.ChannelClient, error)
}
