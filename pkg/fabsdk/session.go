/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fabca "github.com/hyperledger/fabric-sdk-go/api/apifabca"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	apisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
)

// OrgContext currently represents the clients for an organization that the app is dealing with.
// TODO: better decription (e.g., possibility of holding discovery resources for the org & peers).
type OrgContext struct {
	mspClient fabca.FabricCAClient
}

// newOrgContext creates a context based on the providers in the SDK
func newOrgContext(factory apisdk.OrgClientFactory, orgID string, config apiconfig.Config) (*OrgContext, error) {
	c := OrgContext{}

	// TODO: Evaluate context contents during credential client design

	/*
		// Initialize MSP client
		client, err := factory.NewMSPClient(orgID, config)
		if err != nil {
			return nil, errors.WithMessage(err, "MSP client init failed")
		}
		c.mspClient = client
	*/

	return &c, nil
}

// MSPClient provides the MSP client of the context.
func (c *OrgContext) MSPClient() fabca.FabricCAClient {
	return c.mspClient
}

type identityOptions struct {
	identity fab.User
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

		identity, err := sdk.NewPreEnrolledUser(orgName, name)
		if err != nil {
			return errors.WithMessage(err, "Unable to load identity")
		}
		o.identity = identity
		o.ok = true
		return nil

	}
}

// WithIdentity uses a pre-constructed identity object as the credential for the session
func WithIdentity(identity fab.User) IdentityOption {
	return func(o *identityOptions, sdk *FabricSDK, orgName string) error {
		if o.ok {
			return errors.New("Identity already determined")
		}
		o.identity = identity
		o.ok = true
		return nil
	}
}

func (sdk *FabricSDK) newIdentity(orgName string, options ...IdentityOption) (fab.User, error) {
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

// Session represents an identity being used with clients.
// TODO: Better description
// TODO: consider removing this extra wrapper.
type Session struct {
	user fab.User
}

// newSession creates a session from a context and a user (TODO)
func newSession(user fab.User) *Session {
	s := Session{
		user: user,
	}

	return &s
}

// Identity returns the User in the session.
// TODO: reduce interface to identity
func (s *Session) Identity() fab.User {
	return s.user
}
