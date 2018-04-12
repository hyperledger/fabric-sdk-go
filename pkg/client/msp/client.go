/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	mspctx "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp"
	mspapi "github.com/hyperledger/fabric-sdk-go/pkg/msp/api"
	"github.com/pkg/errors"
)

// Client enables access to Client services
type Client struct {
	orgName string
	ctx     context.Client
}

// ClientOption describes a functional parameter for the New constructor
type ClientOption func(*Client) error

// WithOrg option
func WithOrg(orgName string) ClientOption {
	return func(msp *Client) error {
		msp.orgName = orgName
		return nil
	}
}

// New creates a new Client instance
func New(clientProvider context.ClientProvider, opts ...ClientOption) (*Client, error) {

	ctx, err := clientProvider()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create Client")
	}

	msp := Client{
		ctx: ctx,
	}

	for _, param := range opts {
		err := param(&msp)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to create Client")
		}
	}

	if msp.orgName == "" {
		clientConfig, err := ctx.IdentityConfig().Client()
		if err != nil {
			return nil, errors.WithMessage(err, "failed to create Client")
		}
		msp.orgName = clientConfig.Organization
	}

	return &msp, nil
}

func newCAClient(ctx context.Client, orgName string) (mspapi.CAClient, error) {

	caClient, err := msp.NewCAClient(orgName, ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create CA Client")
	}

	return caClient, nil
}

// enrollmentOptions represent enrollment options
type enrollmentOptions struct {
	secret string
}

// EnrollmentOption describes a functional parameter for Enroll
type EnrollmentOption func(*enrollmentOptions) error

// WithSecret enrollment option
func WithSecret(secret string) EnrollmentOption {
	return func(o *enrollmentOptions) error {
		o.secret = secret
		return nil
	}
}

// Enroll enrolls a registered user in order to receive a signed X509 certificate.
// A new key pair is generated for the user. The private key and the
// enrollment certificate issued by the CA are stored in SDK stores.
// They can be retrieved by calling IdentityManager.GetSigningIdentity().
//
// enrollmentID enrollment ID of a registered user
// opts represent enrollment options
func (c *Client) Enroll(enrollmentID string, opts ...EnrollmentOption) error {

	eo := enrollmentOptions{}
	for _, param := range opts {
		err := param(&eo)
		if err != nil {
			return errors.WithMessage(err, "failed to enroll")
		}
	}

	ca, err := newCAClient(c.ctx, c.orgName)
	if err != nil {
		return err
	}
	return ca.Enroll(enrollmentID, eo.secret)
}

// Reenroll reenrolls an enrolled user in order to obtain a new signed X509 certificate
func (c *Client) Reenroll(enrollmentID string) error {
	ca, err := newCAClient(c.ctx, c.orgName)
	if err != nil {
		return err
	}
	return ca.Reenroll(enrollmentID)
}

// Register registers a User with the Fabric CA
// request: Registration Request
// Returns Enrolment Secret
func (c *Client) Register(request *RegistrationRequest) (string, error) {
	ca, err := newCAClient(c.ctx, c.orgName)
	if err != nil {
		return "", err
	}

	var a []mspapi.Attribute
	for i := range request.Attributes {
		a = append(a, mspapi.Attribute{Name: request.Attributes[i].Name, Value: request.Attributes[i].Value, ECert: request.Attributes[i].ECert})
	}

	r := mspapi.RegistrationRequest{
		Name:           request.Name,
		Type:           request.Type,
		MaxEnrollments: request.MaxEnrollments,
		Affiliation:    request.Affiliation,
		Attributes:     a,
		CAName:         request.CAName,
		Secret:         request.Secret,
	}
	return ca.Register(&r)
}

// Revoke revokes a User with the Fabric CA
// request: Revocation Request
func (c *Client) Revoke(request *RevocationRequest) (*RevocationResponse, error) {
	ca, err := newCAClient(c.ctx, c.orgName)
	if err != nil {
		return nil, err
	}
	req := mspapi.RevocationRequest(*request)
	resp, err := ca.Revoke(&req)
	if err != nil {
		return nil, err
	}
	var revokedCerts []RevokedCert
	for i := range resp.RevokedCerts {
		revokedCerts = append(
			revokedCerts,
			RevokedCert{
				Serial: resp.RevokedCerts[i].Serial,
				AKI:    resp.RevokedCerts[i].AKI,
			})
	}

	return &RevocationResponse{
		RevokedCerts: revokedCerts,
		CRL:          resp.CRL,
	}, nil
}

// GetSigningIdentity returns signing identity for id
func (c *Client) GetSigningIdentity(id string) (mspctx.SigningIdentity, error) {
	im, _ := c.ctx.IdentityManager(c.orgName)
	si, err := im.GetSigningIdentity(id)
	if err != nil {
		if err == mspctx.ErrUserNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return si, nil
}
