/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/pkg/errors"
)

type identityOptions struct {
	signingIdentity msp.SigningIdentity
	orgName         string
	username        string
}

// ContextOption provides parameters for creating a session (primarily from a fabric identity/user)
type ContextOption func(s *identityOptions) error

// WithUser uses the named user to load the identity
func WithUser(username string) ContextOption {
	return func(o *identityOptions) error {
		o.username = username
		return nil
	}
}

// WithIdentity uses a pre-constructed identity object as the credential for the session
func WithIdentity(signingIdentity msp.SigningIdentity) ContextOption {
	return func(o *identityOptions) error {
		o.signingIdentity = signingIdentity
		return nil
	}
}

// WithOrg uses the named organization
func WithOrg(org string) ContextOption {
	return func(o *identityOptions) error {
		o.orgName = org
		return nil
	}
}

// ErrAnonymousIdentity is returned when options for identity creation
// don't include neither username nor identity
var ErrAnonymousIdentity = errors.New("missing credentials")

func (sdk *FabricSDK) newIdentity(options ...ContextOption) (msp.SigningIdentity, error) {
	opts := identityOptions{
		orgName: sdk.provider.IdentityConfig().Client().Organization,
	}

	for _, option := range options {
		err1 := option(&opts)
		if err1 != nil {
			return nil, errors.WithMessage(err1, "error in option passed to create identity")
		}
	}

	if opts.signingIdentity == nil && opts.username == "" {
		return nil, ErrAnonymousIdentity
	}

	if opts.signingIdentity != nil {
		return opts.signingIdentity, nil
	}

	if opts.username == "" || opts.orgName == "" {
		return nil, errors.New("invalid options to create identity")
	}

	mgr, ok := sdk.provider.IdentityManager(opts.orgName)
	if !ok {
		return nil, errors.New("invalid options to create identity, invalid org name")
	}

	user, err := mgr.GetSigningIdentity(opts.username)
	if err != nil {
		return nil, err
	}

	return user, nil
}
