/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fabca "github.com/hyperledger/fabric-sdk-go/api/apifabca"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	txn "github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/opt"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/bccsp"
	bccspFactory "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/bccsp/factory"
)

// SDKProviderFactory allows overriding default providers of an SDK
type SDKProviderFactory interface {
	NewConfigProvider(o opt.ConfigOpts, a opt.SDKOpts) (apiconfig.Config, error)
	NewStateStoreProvider(o opt.StateStoreOpts, config apiconfig.Config) (fab.KeyValueStore, error)
	NewCryptoSuiteProvider(config *bccspFactory.FactoryOpts) (bccsp.BCCSP, error)
	NewSigningManager(cryptoProvider bccsp.BCCSP, config apiconfig.Config) (fab.SigningManager, error)
	NewDiscoveryProvider(config apiconfig.Config) (fab.DiscoveryProvider, error)
	NewSelectionProvider(config apiconfig.Config) (fab.SelectionProvider, error)
}

// OrgClientFactory allows overriding default clients and providers of an organization
// Currently, a context is created for each organization that the client app needs.
type OrgClientFactory interface {
	NewMSPClient(orgName string, config apiconfig.Config) (fabca.FabricCAClient, error)
	NewCredentialManager(orgName string, config apiconfig.Config, cryptoProvider bccsp.BCCSP) (fab.CredentialManager, error)
}

// SessionClientFactory allows overriding default clients and providers of a session
type SessionClientFactory interface {
	NewSystemClient(context SDK, session Session, config apiconfig.Config) (fab.FabricClient, error)
	NewChannelClient(context SDK, session Session, config apiconfig.Config, channelID string) (txn.ChannelClient, error)
}
