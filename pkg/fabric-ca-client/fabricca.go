/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricca

import (
	"fmt"

	api "github.com/hyperledger/fabric-ca/api"
	fabric_ca "github.com/hyperledger/fabric-ca/lib"
	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	sdkApi "github.com/hyperledger/fabric-sdk-go/api/apifabca"

	"github.com/hyperledger/fabric/bccsp"
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("fabric_sdk_go")

// FabricCA represents a client to Fabric CA.
type FabricCA struct {
	fabricCAClient *fabric_ca.Client
}

// NewFabricCAClient creates a new fabric-ca client
// @param {api.Config} client config for fabric-ca services
// @param {string} organization for this CA
// @returns {api.FabricCAClient} FabricCAClient implementation
// @returns {error} error, if any
func NewFabricCAClient(config config.Config, org string) (*FabricCA, error) {
	if org == "" || config == nil {
		return nil, fmt.Errorf("Organization and config are required to load CA config")
	}

	// Create new Fabric-ca client without configs
	c := &fabric_ca.Client{
		Config: &fabric_ca.ClientConfig{},
	}

	conf, err := config.CAConfig(org)
	if err != nil {
		return nil, err
	}

	//set server CAName
	c.Config.CAName = conf.Name
	//set server URL
	c.Config.URL = conf.ServerURL
	//certs file list
	c.Config.TLS.CertFiles, err = config.CAServerCertFiles(org)
	if err != nil {
		return nil, err
	}

	// set key file and cert file
	c.Config.TLS.Client.CertFile, err = config.CAClientCertFile(org)
	if err != nil {
		return nil, err
	}

	c.Config.TLS.Client.KeyFile, err = config.CAClientKeyFile(org)
	if err != nil {
		return nil, err
	}

	//TLS flag enabled/disabled
	c.Config.TLS.Enabled = conf.TLSEnabled
	c.Config.MSPDir = config.CAKeyStorePath()
	c.Config.CSP = config.CSPConfig()

	fabricCAClient := FabricCA{fabricCAClient: c}
	logger.Infof("Constructed fabricCAClient instance: %v", fabricCAClient)

	err = c.Init()
	if err != nil {
		return nil, fmt.Errorf("New fabricCAClient failed: %s", err)
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
func (fabricCAServices *FabricCA) Enroll(enrollmentID string, enrollmentSecret string) (bccsp.Key, []byte, error) {
	if enrollmentID == "" {
		return nil, nil, fmt.Errorf("enrollmentID is empty")
	}
	if enrollmentSecret == "" {
		return nil, nil, fmt.Errorf("enrollmentSecret is empty")
	}
	req := &api.EnrollmentRequest{
		CAName: fabricCAServices.fabricCAClient.Config.CAName,
		Name:   enrollmentID,
		Secret: enrollmentSecret,
	}
	enrollmentResponse, err := fabricCAServices.fabricCAClient.Enroll(req)
	if err != nil {
		return nil, nil, fmt.Errorf("Enroll failed: %s", err)
	}
	return enrollmentResponse.Identity.GetECert().Key(), enrollmentResponse.Identity.GetECert().Cert(), nil
}

// Reenroll an enrolled user in order to receive a signed X509 certificate
// Returns X509 certificate
func (fabricCAServices *FabricCA) Reenroll(user sdkApi.User) (bccsp.Key, []byte, error) {
	if user == nil {
		return nil, nil, fmt.Errorf("User does not exist")
	}
	if user.Name() == "" {
		logger.Infof("Invalid re-enroll request, missing argument user")
		return nil, nil, fmt.Errorf("User is empty")
	}
	req := &api.ReenrollmentRequest{
		CAName: fabricCAServices.fabricCAClient.Config.CAName,
	}
	// Create signing identity
	identity, err := fabricCAServices.createSigningIdentity(user)
	if err != nil {
		logger.Infof("Invalid re-enroll request, %s is not a valid user  %s\n", user.Name(), err)
		return nil, nil, fmt.Errorf("Reenroll has failed; Cannot create user identity: %s", err)
	}

	reenrollmentResponse, err := identity.Reenroll(req)
	if err != nil {
		return nil, nil, fmt.Errorf("ReEnroll failed: %s", err)
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
		return "", fmt.Errorf("Registration request cannot be nil")
	}
	// Create request signing identity
	identity, err := fabricCAServices.createSigningIdentity(registrar)
	if err != nil {
		return "", fmt.Errorf("Error creating signing identity: %s", err.Error())
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
		return "", fmt.Errorf("Error Registering User: %s", err.Error())
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
		return fmt.Errorf("Revocation request cannot be nil")
	}
	// Create request signing identity
	identity, err := fabricCAServices.createSigningIdentity(registrar)
	if err != nil {
		return fmt.Errorf("Error creating signing identity: %s", err.Error())
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
		return nil, fmt.Errorf("Valid user required to create signing identity")
	}
	// Validate enrolment information
	cert := user.EnrollmentCertificate()
	key := user.PrivateKey()
	if key == nil || cert == nil {
		return nil, fmt.Errorf(
			"Unable to read user enrolment information to create signing identity")
	}
	return fabricCAServices.fabricCAClient.NewIdentity(key, cert)
}
