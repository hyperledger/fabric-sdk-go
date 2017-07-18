/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defprovider

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/opt"
	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/config"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/keyvaluestore"
	"github.com/hyperledger/fabric/bccsp"
	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
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
	return configImpl.InitConfig(a.ConfigFile)
}

// NewStateStoreProvider creates a KeyValueStore using the SDK's default implementation
func (f *DefaultProviderFactory) NewStateStoreProvider(o opt.StateStoreOpts, config apiconfig.Config) (fab.KeyValueStore, error) {
	stateStore, err := kvs.CreateNewFileKeyValueStore(o.Path)
	if err != nil {
		return nil, fmt.Errorf("CreateNewFileKeyValueStore returned error[%s]", err)
	}
	return stateStore, nil
}

// NewCryptoSuiteProvider returns a new default implementation of BCCSP
func (f *DefaultProviderFactory) NewCryptoSuiteProvider(config *bccspFactory.FactoryOpts) (bccsp.BCCSP, error) {
	return bccspFactory.GetBCCSPFromOpts(config)
}
