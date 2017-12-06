/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricca

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"

	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	sdkApi "github.com/hyperledger/fabric-sdk-go/api/apifabca"
	api "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/api"
	fabric_ca "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/lib"
	"github.com/hyperledger/fabric-sdk-go/pkg/config/urlutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"

	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
)

var logger = logging.NewLogger("fabric_sdk_go")

// FabricCA represents a client to Fabric CA.
type FabricCA struct {
	fabricCAClient *fabric_ca.Client
}

// NewFabricCAClient creates a new fabric-ca client
// @param {string} organization for this CA
// @param {api.Config} client config for fabric-ca services
// @returns {api.FabricCAClient} FabricCAClient implementation
// @returns {error} error, if any
func NewFabricCAClient(org string, config config.Config, cryptoSuite apicryptosuite.CryptoSuite) (*FabricCA, error) {
	if org == "" || config == nil || cryptoSuite == nil {
		return nil, errors.New("organization, config and cryptoSuite are required to load CA config")
	}

	// Create new Fabric-ca client without configs
	c := &fabric_ca.Client{
		Config: &fabric_ca.ClientConfig{},
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
	c.Config.URL = urlutil.ToAddress(conf.URL)
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

	// get Client configs
	_, err = config.Client()
	if err != nil {
		return nil, err
	}

	//TLS flag enabled/disabled
	c.Config.TLS.Enabled = urlutil.IsTLSEnabled(conf.URL)
	c.Config.MSPDir = config.CAKeyStorePath()

	//Factory opts
	c.Config.CSP = cryptoSuite

	fabricCAClient := FabricCA{fabricCAClient: c}

	err = c.Init()
	if err != nil {
		return nil, errors.Wrap(err, "init failed")
	}

	return &fabricCAClient, nil
}

// CAName returns the CA name.
func (fabricCAServices *FabricCA) CAName() string {
	return fabricCAServices.fabricCAClient.Config.CAName
}

// Enroll a registered user in order to receive a signed X509 certificate.
// enrollmentID The registered ID to use for enrollment
// enrollmentSecret The secret associated with the enrollment ID
// Returns X509 certificate
func (fabricCAServices *FabricCA) Enroll(enrollmentID string, enrollmentSecret string) (apicryptosuite.Key, []byte, error) {
	if enrollmentID == "" {
		return nil, nil, errors.New("enrollmentID required")
	}
	if enrollmentSecret == "" {
		return nil, nil, errors.New("enrollmentSecret required")
	}
	req := &api.EnrollmentRequest{
		CAName: fabricCAServices.fabricCAClient.Config.CAName,
		Name:   enrollmentID,
		Secret: enrollmentSecret,
	}
	enrollmentResponse, err := fabricCAServices.fabricCAClient.Enroll(req)
	if err != nil {
		return nil, nil, errors.Wrap(err, "enroll failed")
	}
	return enrollmentResponse.Identity.GetECert().Key(), enrollmentResponse.Identity.GetECert().Cert(), nil
}

// Reenroll an enrolled user in order to receive a signed X509 certificate
// Returns X509 certificate
func (fabricCAServices *FabricCA) Reenroll(user sdkApi.User) (apicryptosuite.Key, []byte, error) {
	if user == nil {
		return nil, nil, errors.New("user required")
	}
	if user.Name() == "" {
		logger.Infof("Invalid re-enroll request, missing argument user")
		return nil, nil, errors.New("user name missing")
	}
	req := &api.ReenrollmentRequest{
		CAName: fabricCAServices.fabricCAClient.Config.CAName,
	}
	// Create signing identity
	identity, err := fabricCAServices.createSigningIdentity(user)
	if err != nil {
		logger.Debugf("Invalid re-enroll request, %s is not a valid user  %s\n", user.Name(), err)
		return nil, nil, errors.Wrap(err, "createSigningIdentity failed")
	}

	reenrollmentResponse, err := identity.Reenroll(req)
	if err != nil {
		return nil, nil, errors.Wrap(err, "reenroll failed")
	}
	return reenrollmentResponse.Identity.GetECert().Key(), reenrollmentResponse.Identity.GetECert().Cert(), nil
}

// Register a User with the Fabric CA
// registrar: The User that is initiating the registration
// request: Registration Request
// Returns Enrolment Secret
func (fabricCAServices *FabricCA) Register(registrar sdkApi.User,
	request *sdkApi.RegistrationRequest) (string, error) {
	// Validate registration request
	if request == nil {
		return "", errors.New("registration request required")
	}
	// Create request signing identity
	identity, err := fabricCAServices.createSigningIdentity(registrar)
	if err != nil {
		return "", errors.Wrap(err, "failed to create request for signing identity")
	}
	// Contruct request for Fabric CA client
	var attributes []api.Attribute
	for i := range request.Attributes {
		attributes = append(attributes, api.Attribute{Name: request.
			Attributes[i].Key, Value: request.Attributes[i].Value})
	}
	var req = api.RegistrationRequest{
		CAName:         request.CAName,
		Name:           request.Name,
		Type:           request.Type,
		MaxEnrollments: request.MaxEnrollments,
		Affiliation:    request.Affiliation,
		Secret:         request.Secret,
		Attributes:     attributes}
	// Make registration request
	response, err := identity.Register(&req)
	if err != nil {
		return "", errors.Wrap(err, "failed to register user")
	}

	return response.Secret, nil
}

// Revoke a User with the Fabric CA
// registrar: The User that is initiating the revocation
// request: Revocation Request
func (fabricCAServices *FabricCA) Revoke(registrar sdkApi.User,
	request *sdkApi.RevocationRequest) error {
	// Validate revocation request
	if request == nil {
		return errors.New("revocation request required")
	}
	// Create request signing identity
	identity, err := fabricCAServices.createSigningIdentity(registrar)
	if err != nil {
		return errors.Wrap(err, "failed to create request for signing identity")
	}
	// Create revocation request
	var req = api.RevocationRequest{
		CAName: request.CAName,
		Name:   request.Name,
		Serial: request.Serial,
		AKI:    request.AKI,
		Reason: request.Reason}
	return identity.Revoke(&req)
}

// createSigningIdentity creates an identity to sign Fabric CA requests with
func (fabricCAServices *FabricCA) createSigningIdentity(user sdkApi.
	User) (*fabric_ca.Identity, error) {
	// Validate user
	if user == nil {
		return nil, errors.New("user required")
	}
	// Validate enrolment information
	cert := user.EnrollmentCertificate()
	key := user.PrivateKey()
	if key == nil || cert == nil {
		return nil, errors.New(
			"Unable to read user enrolment information to create signing identity")
	}
	return fabricCAServices.fabricCAClient.NewIdentity(key, cert)
}
