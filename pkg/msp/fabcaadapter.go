/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"github.com/pkg/errors"

	caapi "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/api"
	calib "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/lib"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp/api"
)

// fabricCAAdapter translates between SDK lingo and native Fabric CA API
type fabricCAAdapter struct {
	config      core.Config
	cryptoSuite core.CryptoSuite
	caClient    *calib.Client
}

func newFabricCAAdapter(orgName string, cryptoSuite core.CryptoSuite, config core.Config) (*fabricCAAdapter, error) {

	caClient, err := createFabricCAClient(orgName, cryptoSuite, config)
	if err != nil {
		return nil, err
	}

	a := &fabricCAAdapter{
		config:      config,
		cryptoSuite: cryptoSuite,
		caClient:    caClient,
	}
	return a, nil
}

// Enroll handles enrollment.
func (c *fabricCAAdapter) Enroll(enrollmentID string, enrollmentSecret string) ([]byte, error) {

	logger.Debugf("Enrolling user [%s]", enrollmentID)

	// TODO add attributes
	careq := &caapi.EnrollmentRequest{
		CAName: c.caClient.Config.CAName,
		Name:   enrollmentID,
		Secret: enrollmentSecret,
	}
	caresp, err := c.caClient.Enroll(careq)
	if err != nil {
		return nil, errors.WithMessage(err, "enroll failed")
	}
	return caresp.Identity.GetECert().Cert(), nil
}

// Reenroll handles re-enrollment
func (c *fabricCAAdapter) Reenroll(key core.Key, cert []byte) ([]byte, error) {

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
func (c *fabricCAAdapter) Register(key core.Key, cert []byte, request *api.RegistrationRequest) (string, error) {
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
func (c *fabricCAAdapter) Revoke(key core.Key, cert []byte, request *api.RevocationRequest) (*api.RevocationResponse, error) {
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
	var revokedCerts []api.RevokedCert
	for i := range resp.RevokedCerts {
		revokedCerts = append(
			revokedCerts,
			api.RevokedCert{
				Serial: resp.RevokedCerts[i].Serial,
				AKI:    resp.RevokedCerts[i].AKI,
			})
	}

	return &api.RevocationResponse{
		RevokedCerts: revokedCerts,
		CRL:          resp.CRL,
	}, nil
}

func createFabricCAClient(org string, cryptoSuite core.CryptoSuite, config core.Config) (*calib.Client, error) {

	// Create new Fabric-ca client without configs
	c := &calib.Client{
		Config: &calib.ClientConfig{},
	}

	conf, err := config.CAConfig(org)
	if err != nil {
		return nil, err
	}

	if conf == nil {
		return nil, errors.Errorf("Orgnization %s have no corresponding CA in the configs", org)
	}

	//set server CAName
	c.Config.CAName = conf.CAName
	//set server URL
	c.Config.URL = endpoint.ToAddress(conf.URL)
	//certs file list
	c.Config.TLS.CertFiles, err = config.CAServerCertPaths(org)
	if err != nil {
		return nil, err
	}

	// set key file and cert file
	c.Config.TLS.Client.CertFile, err = config.CAClientCertPath(org)
	if err != nil {
		return nil, err
	}

	c.Config.TLS.Client.KeyFile, err = config.CAClientKeyPath(org)
	if err != nil {
		return nil, err
	}

	// get CAClient configs
	_, err = config.Client()
	if err != nil {
		return nil, err
	}

	//TLS flag enabled/disabled
	c.Config.TLS.Enabled = endpoint.IsTLSEnabled(conf.URL)
	c.Config.MSPDir = config.CAKeyStorePath()

	//Factory opts
	c.Config.CSP = cryptoSuite

	err = c.Init()
	if err != nil {
		return nil, errors.Wrap(err, "init failed")
	}

	return c, nil
}
