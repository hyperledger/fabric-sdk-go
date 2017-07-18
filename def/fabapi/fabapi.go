/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package fabapi enables client usage of a Hyperledger Fabric network
package fabapi

import (
	"fmt"

	"github.com/hyperledger/fabric/bccsp"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/context"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/context/defprovider"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/opt"
)

// Options encapsulates configuration for the SDK
type Options struct {
	// Quick access options
	ConfigFile string
	//OrgID      string // TODO: separate into context options

	// Options for default providers
	ConfigOpts     opt.ConfigOpts
	StateStoreOpts opt.StateStoreOpts

	// Factory methods to create clients and providers
	ProviderFactory context.SDKProviderFactory
	ContextFactory  context.OrgClientFactory
	SessionFactory  context.SessionClientFactory

	// TODO extract hard-coded logger
}

// FabricSDK provides access (and context) to clients being managed by the SDK
type FabricSDK struct {
	Options

	// Implementations of client functionality (defaults are used if not specified)
	configProvider apiconfig.Config
	stateStore     apifabclient.KeyValueStore
	cryptoSuite    bccsp.BCCSP // TODO - maybe copy this interface into the API package
}

// NewSDK initializes default clients
func NewSDK(options Options) (*FabricSDK, error) {
	// Construct SDK opts from the quick access options in setup
	sdkOpts := opt.SDKOpts{
		ConfigFile: options.ConfigFile,
	}

	sdk := FabricSDK{
		Options: options,
	}

	// Initialize default factories (if needed)
	if sdk.ProviderFactory == nil {
		sdk.ProviderFactory = defprovider.NewDefaultProviderFactory()
	}
	if sdk.ContextFactory == nil {
		sdk.ContextFactory = defprovider.NewOrgClientFactory()
	}
	if sdk.SessionFactory == nil {
		sdk.SessionFactory = defprovider.NewSessionClientFactory()
	}

	// Initialize config provider
	config, err := sdk.ProviderFactory.NewConfigProvider(sdk.ConfigOpts, sdkOpts)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize config [%s]", err)
	}
	sdk.configProvider = config

	// Initialize crypto provider
	cryptosuite, err := sdk.ProviderFactory.NewCryptoSuiteProvider(sdk.configProvider.CSPConfig())
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize crypto suite [%s]", err)
	}
	sdk.cryptoSuite = cryptosuite

	// Initialize state store
	store, err := sdk.ProviderFactory.NewStateStoreProvider(sdk.StateStoreOpts, sdk.configProvider)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize state store [%s]", err)
	}
	sdk.stateStore = store

	return &sdk, nil
}

// ConfigProvider returns the Config provider of sdk.
func (sdk *FabricSDK) ConfigProvider() apiconfig.Config {
	return sdk.configProvider
}

// CryptoSuiteProvider returns the BCCSP provider of sdk.
func (sdk *FabricSDK) CryptoSuiteProvider() bccsp.BCCSP {
	return sdk.cryptoSuite
}

// StateStoreProvider returns the BCCSP provider of sdk.
func (sdk *FabricSDK) StateStoreProvider() apifabclient.KeyValueStore {
	return sdk.stateStore
}

// NewContext creates a context from an org
func (sdk *FabricSDK) NewContext(orgID string) (*OrgContext, error) {
	return NewOrgContext(sdk.ContextFactory, orgID, sdk.configProvider)
}

// NewSession creates a session from a context and a user (TODO)
func (sdk *FabricSDK) NewSession(c context.Org, user apifabclient.User) (*Session, error) {
	return NewSession(user, sdk.SessionFactory), nil
}

// NewSystemClient returns a new client for the system (operations not on a channel)
// TODO: Reduced immutable interface
// TODO: Parameter for setting up the peers
func (sdk *FabricSDK) NewSystemClient(s context.Session) (apifabclient.FabricClient, error) {
	client := NewSystemClient(sdk.configProvider)

	client.SetCryptoSuite(sdk.cryptoSuite)
	client.SetStateStore(sdk.stateStore)

	return client, nil
}

/*
TODO
// NewChannelClient returns a new client for a channel.
func (sdk *FabricSDK) NewChannelClient(s Session) apifabclient.Channel {
	return nil
}
*/
