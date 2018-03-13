/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	mspctx "github.com/hyperledger/fabric-sdk-go/pkg/context/api/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp"
	mspapi "github.com/hyperledger/fabric-sdk-go/pkg/msp/api"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/client")

// MSP enables access to MSP services
type MSP struct {
	orgName string
	ctx     context.Client
}

// Option describes a functional parameter for the New constructor
type Option func(*MSP) error

// WithOrg option
func WithOrg(orgName string) Option {
	return func(msp *MSP) error {
		msp.orgName = orgName
		return nil
	}
}

// New creates a new MSP instance
func New(clientProvider context.ClientProvider, opts ...Option) (*MSP, error) {

	ctx, err := clientProvider()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create MSP")
	}

	msp := MSP{
		ctx: ctx,
	}

	for _, param := range opts {
		err := param(&msp)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to create MSP")
		}
	}

	if msp.orgName == "" {
		clientConfig, err := ctx.Config().Client()
		if err != nil {
			return nil, errors.WithMessage(err, "failed to create MSP")
		}
		msp.orgName = clientConfig.Organization
	}

	return &msp, nil
}

func newCAClient(ctx context.Client, orgName string) (mspapi.CAClient, error) {

	identityManager, ok := ctx.IdentityManager(orgName)
	if !ok {
		return nil, fmt.Errorf("identity manager not found for organization '%s", orgName)
	}
	caClient, err := msp.NewCAClient(orgName, identityManager, ctx.UserStore(), ctx.CryptoSuite(), ctx.Config())
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create CA MSP")
	}

	return caClient, nil
}

// Enroll enrolls a registered user in order to receive a signed X509 certificate.
// A new key pair is generated for the user. The private key and the
// enrollment certificate issued by the CA are stored in SDK stores.
// They can be retrieved by calling IdentityManager.GetSigningIdentity().
//
// enrollmentID enrollment ID of a registered user
// enrollmentSecret secret associated with the enrollment ID
func (c *MSP) Enroll(enrollmentID string, enrollmentSecret string) error {
	ca, err := newCAClient(c.ctx, c.orgName)
	if err != nil {
		return err
	}
	return ca.Enroll(enrollmentID, enrollmentSecret)
}

// Reenroll reenrolls an enrolled user in order to obtain a new signed X509 certificate
func (c *MSP) Reenroll(enrollmentID string) error {
	ca, err := newCAClient(c.ctx, c.orgName)
	if err != nil {
		return err
	}
	return ca.Reenroll(enrollmentID)
}

// Register registers a User with the Fabric CA
// request: Registration Request
// Returns Enrolment Secret
func (c *MSP) Register(request *mspapi.RegistrationRequest) (string, error) {
	ca, err := newCAClient(c.ctx, c.orgName)
	if err != nil {
		return "", err
	}
	return ca.Register(request)
}

// Revoke revokes a User with the Fabric CA
// request: Revocation Request
func (c *MSP) Revoke(request *mspapi.RevocationRequest) (*mspapi.RevocationResponse, error) {
	ca, err := newCAClient(c.ctx, c.orgName)
	if err != nil {
		return nil, err
	}
	return ca.Revoke(request)
}

// GetSigningIdentity returns a signing identity for the given user name
func (c *MSP) GetSigningIdentity(userName string) (*mspctx.SigningIdentity, error) {
	user, err := c.GetUser(userName)
	if err != nil {
		return nil, err
	}
	signingIdentity := &mspctx.SigningIdentity{MspID: user.MspID(), PrivateKey: user.PrivateKey(), EnrollmentCert: user.EnrollmentCertificate()}
	return signingIdentity, nil
}

// GetUser returns a user for the given user name
func (c *MSP) GetUser(userName string) (mspctx.User, error) {
	im, _ := c.ctx.IdentityManager(c.orgName)
	return im.GetUser(userName)
}
