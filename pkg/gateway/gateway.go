/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	mspProvider "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
)

type gateway struct {
	sdk     *fabsdk.FabricSDK
	options *gatewayOptions
	cfg     core.ConfigBackend
	org     string
}

// Connect to a gateway defined by a network config file.
// Must specify a config option, an identity option and zero or more strategy options.
func Connect(config ConfigOption, identity IdentityOption, options ...Option) (Gateway, error) {

	g := &gateway{
		options: &gatewayOptions{
			CommitHandler: DefaultCommitHandlers.OrgAll,
			Discovery:     true,
		},
	}

	err := config(g, g.options)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to apply config option")
	}

	err = identity(g, g.options)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to apply identity option")
	}

	for _, option := range options {
		err = option(g, g.options)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to apply gateway option")
		}
	}

	return g, nil
}

// WithConfig configures the gateway from a network config, such as a ccp file.
func WithConfig(config core.ConfigProvider) ConfigOption {
	return func(gw Gateway, o *gatewayOptions) error {
		var err error
		sdk, err := fabsdk.New(config)

		if err != nil {
			return err
		}

		gw.(*gateway).sdk = sdk

		configBackend, err := config()
		if err != nil {
			return err
		}
		if len(configBackend) != 1 {
			return errors.New("invalid config file")
		}

		cfg := configBackend[0]
		gw.(*gateway).cfg = cfg

		value, ok := cfg.Lookup("client.organization")
		if !ok {
			return errors.New("No client organization defined in the config")
		}
		gw.(*gateway).org = value.(string)

		return nil
	}
}

// WithSDK configures the gateway with the configuration from an existing FabricSDK instance
func WithSDK(sdk *fabsdk.FabricSDK) ConfigOption {
	return func(gw Gateway, o *gatewayOptions) error {
		gw.(*gateway).sdk = sdk

		cfg, err := sdk.Config()

		if err != nil {
			return errors.Wrap(err, "Unable to access SDK configuration")
		}

		value, ok := cfg.Lookup("client.organization")
		if !ok {
			return errors.New("No client organization defined in the config")
		}
		gw.(*gateway).org = value.(string)

		return nil
	}
}

// WithIdentity is an optional argument to the Connect method which specifies
// the identity that is to be used to connect to the network.
// All operations under this gateway connection will be performed using this identity.
func WithIdentity(wallet Wallet, label string) IdentityOption {
	return func(gw Gateway, o *gatewayOptions) error {
		mspClient, err := msp.New(gw.getSdk().Context(), msp.WithOrg(gw.getOrg()))
		if err != nil {
			return err
		}

		creds, err := wallet.Get(label)
		if err != nil {
			return err
		}

		var identity mspProvider.SigningIdentity
		switch v := creds.(type) {
		case *X509Identity:
			identity, err = mspClient.CreateSigningIdentity(mspProvider.WithCert([]byte(v.GetCert())), mspProvider.WithPrivateKey([]byte(v.GetKey())))
			if err != nil {
				return err
			}
		}

		o.Identity = identity
		return nil
	}
}

// WithUser is an optional argument to the Connect method which specifies
// the identity that is to be used to connect to the network.
// All operations under this gateway connection will be performed using this identity.
func WithUser(user string) IdentityOption {
	return func(gw Gateway, o *gatewayOptions) error {
		o.User = user
		return nil
	}
}

// WithCommitHandler is an optional argument to the Connect method which
// allows an alternative commit handler to be specified. The commit handler defines how
// client code should wait to receive commit events from peers following submit of a transaction.
// Currently unimplemented.
func WithCommitHandler(handler CommitHandlerFactory) Option {
	return func(gw Gateway, o *gatewayOptions) error {
		o.CommitHandler = handler
		return nil
	}
}

// WithDiscovery is an optional argument to the Connect method which
// enables or disables service discovery for all transaction submissions for this gateway.
func WithDiscovery(discovery bool) Option {
	return func(gw Gateway, o *gatewayOptions) error {
		o.Discovery = discovery
		return nil
	}
}

func (gw *gateway) getSdk() *fabsdk.FabricSDK {
	return gw.sdk
}

func (gw *gateway) getOrg() string {
	return gw.org
}

func (gw *gateway) getPeersForOrg(org string) ([]string, error) {
	value, ok := gw.cfg.Lookup("organizations." + org + ".peers")
	if !ok {
		return nil, errors.New("No client organization defined in the config")
	}

	val := value.([]interface{})
	s := make([]string, len(val))
	for i, v := range val {
		s[i] = fmt.Sprint(v)
	}

	return s, nil
}

func (gw *gateway) GetNetwork(name string) (Network, error) {
	var channelProvider context.ChannelProvider
	if gw.options.Identity != nil {
		channelProvider = gw.sdk.ChannelContext(name, fabsdk.WithIdentity(gw.options.Identity), fabsdk.WithOrg(gw.org))
	} else {
		channelProvider = gw.sdk.ChannelContext(name, fabsdk.WithUser(gw.options.User), fabsdk.WithOrg(gw.org))
	}
	return newNetwork(gw, channelProvider)
}

func (gw *gateway) Close() {
	// future use
}
