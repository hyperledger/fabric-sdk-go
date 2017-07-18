/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fabca "github.com/hyperledger/fabric-sdk-go/api/apifabca"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/opt"
	"github.com/hyperledger/fabric/bccsp"
	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
)

// SDKProviderFactory allows overriding default providers of an SDK
type SDKProviderFactory interface {
	NewConfigProvider(o opt.ConfigOpts, a opt.SDKOpts) (apiconfig.Config, error)
	NewStateStoreProvider(o opt.StateStoreOpts, config apiconfig.Config) (fab.KeyValueStore, error)
	NewCryptoSuiteProvider(config *bccspFactory.FactoryOpts) (bccsp.BCCSP, error)
}

// OrgClientFactory allows overriding default clients and providers of an organization
// Currently, a context is created for each organization that the client app needs.
type OrgClientFactory interface {
	NewMSPClient(orgName string, config apiconfig.Config) (fabca.FabricCAClient, error)
}

// SessionClientFactory allows overriding default clients and providers of a session
type SessionClientFactory interface {
	NewSystemClient(context SDK, session Session, config apiconfig.Config) (fab.FabricClient, error)
	//NewChannelClient(session Session) fab.Channel
}
