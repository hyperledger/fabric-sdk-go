/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/api"
	caapi "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/api"
	calib "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/lib"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/msp"
)

// FabricCAAdapter translates between SDK lingo to native Fabric CA API
type FabricCAAdapter struct {
	caName      string
	config      core.Config
	cryptoSuite core.CryptoSuite
	caClient    *calib.Client
}

func newFabricCAAdapter(orgName string, caName string, cryptoSuite core.CryptoSuite, config core.Config) (*FabricCAAdapter, error) {

	caClient, err := createFabricCAClient(orgName, cryptoSuite, config)
	if err != nil {
		return nil, err
	}

	a := &FabricCAAdapter{
		caName:      caName,
		config:      config,
		cryptoSuite: cryptoSuite,
		caClient:    caClient,
	}
	return a, nil
}

// CAName returns the CA name.
func (c *FabricCAAdapter) CAName() string {
	return c.caName
}

// Enroll handles enrollment.
func (c *FabricCAAdapter) Enroll(req *api.EnrollmentRequest) ([]byte, error) {

	logger.Debugf("Enrolling user [%s]", req.Name)

	// TODO add attributes
	careq := &caapi.EnrollmentRequest{
		CAName: c.caClient.Config.CAName,
		Name:   req.Name,
		Secret: req.Secret,
	}
	caresp, err := c.caClient.Enroll(careq)
	if err != nil {
		return nil, errors.WithMessage(err, "enroll failed")
	}
	return caresp.Identity.GetECert().Cert(), nil
}

// Reenroll handles re-enrollment
func (c *FabricCAAdapter) Reenroll(key core.Key, cert []byte, req *api.ReenrollmentRequest) ([]byte, error) {

	logger.Debugf("Enrolling user [%s]")

	careq := &caapi.ReenrollmentRequest{
		CAName: c.caClient.Config.CAName,
	}
	caidentity, err := c.caClient.NewIdentity(key, cert)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create CA signing identity")
	}

	caresp, err := caidentity.Reenroll(careq)
	if err != nil {
		return nil, errors.WithMessage(err, "reenroll failed")
	}

	return caresp.Identity.GetECert().Cert(), nil
}

// Register handles user registration
// key: registrar private key
// cert: registrar enrollment certificate
// request: Registration Request
// Returns Enrolment Secret
func (c *FabricCAAdapter) Register(key core.Key, cert []byte, request *msp.RegistrationRequest) (string, error) {
	// Contruct request for Fabric CA client
	var attributes []caapi.Attribute
	for i := range request.Attributes {
		attributes = append(attributes, caapi.Attribute{Name: request.
			Attributes[i].Key, Value: request.Attributes[i].Value})
	}
	var req = caapi.RegistrationRequest{
		CAName:         request.CAName,
		Name:           request.Name,
		Type:           request.Type,
		MaxEnrollments: request.MaxEnrollments,
		Affiliation:    request.Affiliation,
		Secret:         request.Secret,
		Attributes:     attributes}

	registrar, err := c.caClient.NewIdentity(key, cert)
	if err != nil {
		return "", errors.Wrap(err, "failed to create CA signing identity")
	}

	response, err := registrar.Register(&req)
	if err != nil {
		return "", errors.Wrap(err, "failed to register user")
	}

	return response.Secret, nil
}

// Revoke handles user revocation.
// key: registrar private key
// cert: registrar enrollment certificate
// request: Revocation Request
func (c *FabricCAAdapter) Revoke(key core.Key, cert []byte, request *msp.RevocationRequest) (*msp.RevocationResponse, error) {
	// Create revocation request
	var req = caapi.RevocationRequest{
		CAName: request.CAName,
		Name:   request.Name,
		Serial: request.Serial,
		AKI:    request.AKI,
		Reason: request.Reason,
	}

	registrar, err := c.caClient.NewIdentity(key, cert)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create CA signing identity")
	}

	resp, err := registrar.Revoke(&req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to revoke")
	}
	var revokedCerts []msp.RevokedCert
	for i := range resp.RevokedCerts {
		revokedCerts = append(
			revokedCerts,
			msp.RevokedCert{
				Serial: resp.RevokedCerts[i].Serial,
				AKI:    resp.RevokedCerts[i].AKI,
			})
	}

	return &msp.RevocationResponse{
		RevokedCerts: revokedCerts,
		CRL:          resp.CRL,
	}, nil
}
