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

// SDKOpts provides bootstrap setup
type SDKOpts struct {
	//ConfigFile to load from a predefined path
	ConfigFile string
	//ConfigBytes to load from an bytes array
	ConfigBytes []byte
	//ConfigType to specify the type of the config (mainly used with ConfigBytes as ConfigFile has a file extension to specify the type)
	// valid values: yaml, json, etc.
	ConfigType string
}

// StateStoreOpts provides setup parameters for KeyValueStore
type StateStoreOpts struct {
	Path string
}

// CoreProviderFactory allows overriding of primitives and the fabric core object provider
type CoreProviderFactory interface {
	NewConfigProvider(a SDKOpts) (apiconfig.Config, error)
	NewStateStoreProvider(o StateStoreOpts, config apiconfig.Config) (fab.KeyValueStore, error)
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
	//NewSystemClient(context SDK, session Session, config apiconfig.Config) (fab.FabricClient, error)
	NewChannelMgmtClient(context SDK, session Session, config apiconfig.Config) (chmgmt.ChannelMgmtClient, error)
	NewResourceMgmtClient(context SDK, session Session, config apiconfig.Config, filter resmgmt.TargetFilter) (resmgmt.ResourceMgmtClient, error)
	NewChannelClient(context SDK, session Session, config apiconfig.Config, channelID string) (txn.ChannelClient, error)
}

// PkgSuite holds the package factories that create clients and providers
type PkgSuite struct {
	Core    CoreProviderFactory
	Service ServiceProviderFactory
	Context OrgClientFactory
	Session SessionClientFactory
	Logger  apilogging.LoggerProvider
}
