/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package gateway enables Go developers to build client applications using the Hyperledger
// Fabric programming model as described in the 'Developing Applications' chapter of the Fabric
// documentation: https://hyperledger-fabric.readthedocs.io/en/master/developapps/developing_applications.html
//
// A Gateway object is created using the Connect() function to connect to a 'gateway' peer
// as specified in a network configuration file, using an identity stored in a wallet.
// Interactions with smart contracts are then invoked within the context of this gateway connection.
//
// See https://github.com/hyperledger/fabric-samples/blob/master/fabcar/go/fabcar.go
// for a working sample.
package gateway

import (
	"os"
	"strings"
	"time"

	fabricCaUtil "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	mspProvider "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/pkg/errors"
)

const (
	defaultTimeout      = 5 * time.Minute
	localhostEnvVarName = "DISCOVERY_AS_LOCALHOST"
)

// Gateway is the entry point to a Fabric network
type Gateway struct {
	sdk        *fabsdk.FabricSDK
	options    *gatewayOptions
	cfg        core.ConfigBackend
	org        string
	mspid      string
	peers      []fab.PeerConfig
	mspfactory api.MSPProviderFactory
}

type gatewayOptions struct {
	Identity mspProvider.SigningIdentity
	User     string
	Timeout  time.Duration
}

// Option functional arguments can be supplied when connecting to the gateway.
type Option = func(*Gateway) error

// ConfigOption specifies the gateway configuration source.
type ConfigOption = func(*Gateway) error

// IdentityOption specifies the user identity under which all transactions are performed for this gateway instance.
type IdentityOption = func(*Gateway) error

// Connect to a gateway defined by a network config file.
// Must specify a config option, an identity option and zero or more strategy options.
//  Parameters:
//  config is a ConfigOption used to specify the network connection configuration.  This must contain connection details for at least one 'gateway' peer.
//  identity is an IdentityOption which assigns a signing identity for all interactions under this Gateway connection.
//  options specifies other gateway options
//
//  Returns:
//  A Transaction object for subsequent evaluation or submission.
func Connect(config ConfigOption, identity IdentityOption, options ...Option) (*Gateway, error) {

	g := &Gateway{
		options: &gatewayOptions{
			Timeout: defaultTimeout,
		},
	}

	err := identity(g)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to apply identity option")
	}

	err = config(g)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to apply config option")
	}

	for _, option := range options {
		err = option(g)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to apply gateway option")
		}
	}

	return g, nil
}

// WithConfig configures the gateway from a network config, such as a ccp file.
//
//   Parameters:
//   config is a ConfigProvider function which provides config backend
//
//   Returns:
//   A ConfigOption which can be passed as the first parameter to the Connect() function
func WithConfig(config core.ConfigProvider) ConfigOption {
	return func(gw *Gateway) error {
		config = createGatewayConfigProvider(config, gw.getOrg)

		configBackend, err := config()
		if err != nil {
			return err
		}
		if len(configBackend) != 1 {
			return errors.New("invalid config file")
		}

		gw.cfg = configBackend[0]

		value, ok := gw.cfg.Lookup("client.organization")
		if !ok {
			return errors.New("No client organization defined in the config")
		}
		gw.org = value.(string)

		value, ok = gw.cfg.Lookup("organizations." + gw.org + ".mspid")
		if !ok {
			return errors.New("No client organization defined in the config")
		}
		gw.mspid = value.(string)

		opts := []fabsdk.Option{}
		if gw.mspfactory != nil {
			opts = append(opts, fabsdk.WithMSPPkg(gw.mspfactory))
		}

		sdk, err := fabsdk.New(config, opts...)

		if err != nil {
			return err
		}

		gw.sdk = sdk

		//  find the 'gateway' peers
		ctx := sdk.Context()
		client, _ := ctx()
		gw.peers, _ = client.EndpointConfig().PeersConfig(gw.org)

		return nil
	}
}

// WithSDK configures the gateway with the configuration from an existing FabricSDK instance
//
//   Parameters:
//   sdk is an instance of fabsdk.FabricSDK from which the configuration is extracted
//
//   Returns:
//   A ConfigOption which can be passed as the first parameter to the Connect() function
func WithSDK(sdk *fabsdk.FabricSDK) ConfigOption {
	return func(gw *Gateway) error {
		gw.sdk = sdk

		cfg, err := sdk.Config()

		if err != nil {
			return errors.Wrap(err, "Unable to access SDK configuration")
		}

		value, ok := cfg.Lookup("client.organization")
		if !ok {
			return errors.New("No client organization defined in the config")
		}
		gw.org = value.(string)

		return nil
	}
}

// WithIdentity is an optional argument to the Connect method which specifies
// the identity that is to be used to connect to the network.
// All operations under this gateway connection will be performed using this identity.
//
//   Parameters:
//   wallet is a Wallet implementation that contains identities
//   label is the name of the identity in the wallet to associate with the gateway
//
//   Returns:
//   An IdentityOption which can be passed as the second parameter to the Connect() function
func WithIdentity(wallet wallet, label string) IdentityOption {
	return func(gw *Gateway) error {
		creds, err := wallet.Get(label)
		if err != nil {
			return err
		}

		privateKey, _ := fabricCaUtil.ImportBCCSPKeyFromPEMBytes([]byte(creds.(*X509Identity).Key()), cryptosuite.GetDefault(), true)
		wid := &walletIdentity{
			id:                    label,
			mspID:                 creds.mspID(),
			enrollmentCertificate: []byte(creds.(*X509Identity).Certificate()),
			privateKey:            privateKey,
		}

		gw.options.Identity = wid
		gw.mspfactory = &walletmsp{}

		return nil
	}
}

// WithUser is an optional argument to the Connect method which specifies
// the identity that is to be used to connect to the network.
// The creadentials are extracted from the credential store specified in the connection profile.
// All operations under this gateway connection will be performed using this identity.
//
//   Parameters:
//   user is the name of the user in the credential store.
//
//   Returns:
//   An IdentityOption which can be passed as the second parameter to the Connect() function
func WithUser(user string) IdentityOption {
	return func(gw *Gateway) error {
		gw.options.User = user
		return nil
	}
}

// WithTimeout is an optional argument to the Connect method which
// defines the commit timeout for all transaction submissions for this gateway.
func WithTimeout(timeout time.Duration) Option {
	return func(gw *Gateway) error {
		gw.options.Timeout = timeout
		return nil
	}
}

// GetNetwork returns an object representing a network channel.
//  Parameters:
//  name is the name of the network channel
//
//  Returns:
//  A Network object representing the channel
func (gw *Gateway) GetNetwork(name string) (*Network, error) {
	var channelProvider context.ChannelProvider
	if gw.options.Identity != nil {
		channelProvider = gw.sdk.ChannelContext(name, fabsdk.WithIdentity(gw.options.Identity), fabsdk.WithOrg(gw.org))
	} else {
		channelProvider = gw.sdk.ChannelContext(name, fabsdk.WithUser(gw.options.User), fabsdk.WithOrg(gw.org))
	}
	return newNetwork(gw, channelProvider)
}

// Close the gateway connection and all associated resources, including removing listeners attached to networks and
// contracts created by the gateway.
func (gw *Gateway) Close() {
	// future use
}

func (gw *Gateway) getOrg() string {
	return gw.org
}

func createGatewayConfigProvider(config core.ConfigProvider, org func() string) func() ([]core.ConfigBackend, error) {
	return func() ([]core.ConfigBackend, error) {
		configBackend, err := config()
		if err != nil {
			return nil, err
		}
		if len(configBackend) != 1 {
			return nil, errors.New("invalid config file")
		}

		cfg := configBackend[0]

		lhConfig := make([]core.ConfigBackend, 0)
		lhConfig = append(lhConfig, createGatewayConfig(cfg, org()))

		return lhConfig, nil
	}
}

func createGatewayConfig(backend core.ConfigBackend, org string) *gatewayConfig {
	var matchers map[string][]map[string]string
	if strings.ToUpper(os.Getenv(localhostEnvVarName)) == "TRUE" {
		matchers = createLocalhostMappings()
	}

	var channelConfig map[string]map[string]map[string]map[string]bool
	_, exists := backend.Lookup("channels")
	if !exists {
		channelConfig = createDefaultChannelConfig(backend, org)
	}

	return &gatewayConfig{
		backend:    backend,
		matchers:   matchers,
		channelDef: channelConfig,
	}
}

/* dynamically add the following to CCP:

entityMatchers:
  peer:
    - pattern: ([^:]+):(\\d+)
      urlSubstitutionExp: localhost:${2}
      sslTargetOverrideUrlSubstitutionExp: ${1}
      mappedHost: ${1}
  peer:
    - pattern: ([^:]+):(\\d+)
      urlSubstitutionExp: localhost:${2}
      sslTargetOverrideUrlSubstitutionExp: localhost
      mappedHost: ${1}
*/
func createLocalhostMappings() map[string][]map[string]string {
	matchers := make(map[string][]map[string]string)
	peerMappings := make([]map[string]string, 0)
	ordererMappings := make([]map[string]string, 0)
	mappedHost := "${1}"

	peerMapping := make(map[string]string)
	peerMapping["pattern"] = "([^:]+):(\\d+)"
	peerMapping["urlSubstitutionExp"] = "localhost:${2}"
	peerMapping["sslTargetOverrideUrlSubstitutionExp"] = mappedHost
	peerMapping["mappedHost"] = mappedHost
	peerMappings = append(peerMappings, peerMapping)

	matchers["peer"] = peerMappings

	ordererMapping := make(map[string]string)
	ordererMapping["pattern"] = "([^:]+):(\\d+)"
	ordererMapping["urlSubstitutionExp"] = "localhost:${2}"
	ordererMapping["sslTargetOverrideUrlSubstitutionExp"] = "localhost"
	ordererMapping["mappedHost"] = mappedHost
	ordererMappings = append(ordererMappings, ordererMapping)

	matchers["orderer"] = ordererMappings
	return matchers
}

/* dynamically add the following to CCP:

channels:
  _default:
    peers:
      <gateway_peer_name>:
        endorsingPeer: true
        chaincodeQuery: true
        ledgerQuery: true
        eventSource: true
*/
func createDefaultChannelConfig(backend core.ConfigBackend, org string) map[string]map[string]map[string]map[string]bool {
	channels := make(map[string]map[string]map[string]map[string]bool)
	_default := make(map[string]map[string]map[string]bool)
	gateways := make(map[string]map[string]bool)
	roles := make(map[string]bool)
	roles["endorsingPeer"] = true
	roles["chaincodeQuery"] = true
	roles["ledgerQuery"] = true
	roles["eventSource"] = true

	value, ok := backend.Lookup("organizations." + org + ".peers")
	if !ok {
		return nil
	}
	arr := value.([]interface{})
	for _, gatewayPeer := range arr {
		gateways[gatewayPeer.(string)] = roles
	}

	_default["peers"] = gateways
	channels["_default"] = _default
	return channels
}

type gatewayConfig struct {
	backend    core.ConfigBackend
	matchers   map[string][]map[string]string
	channelDef map[string]map[string]map[string]map[string]bool
}

func (gc *gatewayConfig) Lookup(key string) (interface{}, bool) {
	if key == "entityMatchers" && gc.matchers != nil {
		return gc.matchers, true
	}
	conf, exists := gc.backend.Lookup(key)
	if key == "channels" && gc.channelDef != nil {
		return gc.channelDef, true
	}
	return conf, exists
}
