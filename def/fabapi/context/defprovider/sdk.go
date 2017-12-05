/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defprovider

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"

	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/opt"
	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite"
	cryptosuiteimpl "github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite/bccsp"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/keyvaluestore"
	signingMgr "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/signingmgr"
	discovery "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/discovery/staticdiscovery"
	selection "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/selection/staticselection"
)

// DefaultProviderFactory represents the default SDK provider factory.
type DefaultProviderFactory struct{}

// NewDefaultProviderFactory returns the default SDK provider factory.
func NewDefaultProviderFactory() *DefaultProviderFactory {
	f := DefaultProviderFactory{}
	return &f
}

// NewConfigProvider creates a Config using the SDK's default implementation
func (f *DefaultProviderFactory) NewConfigProvider(o opt.ConfigOpts, a opt.SDKOpts) (apiconfig.Config, error) {
	// configBytes takes precedence over configFile
	if a.ConfigBytes != nil && len(a.ConfigBytes) > 0 {
		return configImpl.InitConfigFromBytes(a.ConfigBytes, a.ConfigType)
	}
	return configImpl.InitConfig(a.ConfigFile)
}

// NewStateStoreProvider creates a KeyValueStore using the SDK's default implementation
func (f *DefaultProviderFactory) NewStateStoreProvider(o opt.StateStoreOpts, config apiconfig.Config) (fab.KeyValueStore, error) {

	var stateStorePath = o.Path
	if stateStorePath == "" {
		clientCofig, err := config.Client()
		if err != nil {
			return nil, err
		}
		stateStorePath = clientCofig.CredentialStore.Path
	}

	stateStore, err := kvs.CreateNewFileKeyValueStore(stateStorePath)
	if err != nil {
		return nil, errors.WithMessage(err, "CreateNewFileKeyValueStore failed")
	}
	return stateStore, nil
}

// NewCryptoSuiteProvider returns a new default implementation of BCCSP
func (f *DefaultProviderFactory) NewCryptoSuiteProvider(config apiconfig.Config) (apicryptosuite.CryptoSuite, error) {
	cryptoSuiteProvider, err := cryptosuiteimpl.GetSuiteByConfig(config)
	//Setting this cryptosuite as a factory default too
	if cryptoSuiteProvider != nil {
		cryptosuite.SetDefault(cryptoSuiteProvider)
	}
	return cryptoSuiteProvider, err
}

// NewSigningManager returns a new default implementation of signing manager
func (f *DefaultProviderFactory) NewSigningManager(cryptoProvider apicryptosuite.CryptoSuite, config apiconfig.Config) (fab.SigningManager, error) {
	return signingMgr.NewSigningManager(cryptoProvider, config)
}

// NewDiscoveryProvider returns a new default implementation of discovery provider
func (f *DefaultProviderFactory) NewDiscoveryProvider(config apiconfig.Config) (fab.DiscoveryProvider, error) {
	return discovery.NewDiscoveryProvider(config)
}

// NewSelectionProvider returns a new default implementation of selection service
func (f *DefaultProviderFactory) NewSelectionProvider(config apiconfig.Config) (fab.SelectionProvider, error) {
	return selection.NewSelectionProvider(config)
}
