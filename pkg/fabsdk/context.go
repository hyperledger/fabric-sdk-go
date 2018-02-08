/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	apifabclient "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/kvstore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/pkg/errors"
)

type fabContext struct {
	sdk *FabricSDK
}

type sdkContext struct {
	fabContext
}

// Config returns the Config provider of sdk.
func (c *fabContext) Config() apiconfig.Config {
	return c.sdk.config
}

// CryptoSuite returns the BCCSP provider of sdk.
func (c *fabContext) CryptoSuite() apicryptosuite.CryptoSuite {
	return c.sdk.cryptoSuite
}

// SigningManager returns signing manager
func (c *fabContext) SigningManager() apifabclient.SigningManager {
	return c.sdk.signingManager
}

// StateStore returns state store
func (c *sdkContext) StateStore() kvstore.KVStore {
	return c.sdk.stateStore
}

// DiscoveryProvider returns discovery provider
func (c *sdkContext) DiscoveryProvider() apifabclient.DiscoveryProvider {
	return c.sdk.discoveryProvider
}

// SelectionProvider returns selection provider
func (c *sdkContext) SelectionProvider() apifabclient.SelectionProvider {
	return c.sdk.selectionProvider
}

// FabricProvider provides fabric objects such as peer and user
func (c *sdkContext) FabricProvider() api.FabricProvider {
	return c.sdk.fabricProvider
}

// ChannelProvider provides channel services.
func (c *sdkContext) ChannelProvider() apifabclient.ChannelProvider {
	return c.sdk.channelProvider
}

type identityOptions struct {
	identity apifabclient.IdentityContext
	ok       bool
}

// IdentityOption provides parameters for creating a session (primarily from a fabric identity/user)
type IdentityOption func(s *identityOptions, sdk *FabricSDK, orgName string) error

// WithUser uses the named user to load the identity
func WithUser(name string) IdentityOption {
	return func(o *identityOptions, sdk *FabricSDK, orgName string) error {
		if o.ok {
			return errors.New("Identity already determined")
		}

		identity, err := sdk.newUser(orgName, name)
		if err != nil {
			return errors.WithMessage(err, "Unable to load identity")
		}
		o.identity = identity
		o.ok = true
		return nil

	}
}

// WithIdentity uses a pre-constructed identity object as the credential for the session
func WithIdentity(identity apifabclient.IdentityContext) IdentityOption {
	return func(o *identityOptions, sdk *FabricSDK, orgName string) error {
		if o.ok {
			return errors.New("Identity already determined")
		}
		o.identity = identity
		o.ok = true
		return nil
	}
}

func (sdk *FabricSDK) newIdentity(orgName string, options ...IdentityOption) (apifabclient.IdentityContext, error) {
	opts := identityOptions{}

	for _, option := range options {
		err := option(&opts, sdk, orgName)
		if err != nil {
			return nil, errors.WithMessage(err, "Error in option passed to client")
		}
	}

	if !opts.ok {
		return nil, errors.New("Missing identity")
	}

	return opts.identity, nil
}

// session represents an identity being used with clients along with services
// that associate with that identity (particularly the channel service).
type session struct {
	apifabclient.IdentityContext
}

// newSession creates a session from a context and a user (TODO)
func newSession(ic apifabclient.IdentityContext, cp apifabclient.ChannelProvider) *session {
	s := session{
		IdentityContext: ic,
	}

	return &s
}

// FabricProvider provides fabric objects such as peer and user
//
// TODO: move under Providers()
func (sdk *FabricSDK) FabricProvider() api.FabricProvider {
	return sdk.fabricProvider
}
