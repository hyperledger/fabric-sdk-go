/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricca

import (
	"fmt"

	api "github.com/hyperledger/fabric-ca/api"
	fabric_ca "github.com/hyperledger/fabric-ca/lib"
	sdkApi "github.com/hyperledger/fabric-sdk-go/api"

	"github.com/hyperledger/fabric/bccsp"
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("fabric_sdk_go")

type services struct {
	fabricCAClient *fabric_ca.Client
}

// NewFabricCAClient ...
/**
 * @param {string} clientConfigFile for fabric-ca services"
 */
func NewFabricCAClient(config sdkApi.Config) (sdkApi.Services, error) {

	// Create new Fabric-ca client without configs
	c := &fabric_ca.Client{
		Config: &fabric_ca.ClientConfig{},
	}

	//set server CAName
	c.Config.CAName = config.GetFabricCAName()
	//set server URL
	c.Config.URL = config.GetServerURL()
	//certs file list
	c.Config.TLS.CertFiles = config.GetServerCertFiles()

	// set key file and cert file
	c.Config.TLS.Client.CertFile = string(config.GetFabricCAClientCertFile())
	c.Config.TLS.Client.KeyFile = string(config.GetFabricCAClientKeyFile())

	//TLS falg enabled/disabled
	c.Config.TLS.Enabled = config.GetFabricCATLSEnabledFlag()
	fabricCAClient := &services{fabricCAClient: c}
	logger.Infof("Constructed fabricCAClient instance: %v", fabricCAClient)
	c.HomeDir = config.GetFabricCAHomeDir()
	c.Config.MSPDir = config.GetFabricCAMspDir()
	c.Config.CSP = config.GetCSPConfig()

	err := c.Init()
	if err != nil {
		return nil, fmt.Errorf("New fabricCAClient failed: %s", err)
	}

	return fabricCAClient, nil
}

func (fabricCAServices *services) GetCAName() string {
	return fabricCAServices.fabricCAClient.Config.CAName
}

// Enroll ...
/**
 * Enroll a registered user in order to receive a signed X509 certificate
 * @param {string} enrollmentID The registered ID to use for enrollment
 * @param {string} enrollmentSecret The secret associated with the enrollment ID
 * @returns {[]byte} X509 certificate
 * @returns {[]byte} private key
 */
func (fabricCAServices *services) Enroll(enrollmentID string, enrollmentSecret string) (bccsp.Key, []byte, error) {
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

/**
 * ReEnroll an enrolled user in order to receive a signed X509 certificate
 * @param {user} User to be reenrolled
 * @returns {[]byte} X509 certificate
 * @returns {[]byte} private key
 */
func (fabricCAServices *services) Reenroll(user sdkApi.User) (bccsp.Key, []byte, error) {
	if user == nil {
		return nil, nil, fmt.Errorf("User does not exist")
	}
	if user.GetName() == "" {
		logger.Infof("Invalid re-enroll request, missing argument user")
		return nil, nil, fmt.Errorf("User is empty")
	}
	req := &api.ReenrollmentRequest{
		CAName: fabricCAServices.fabricCAClient.Config.CAName,
	}
	// Create signing identity
	identity, err := fabricCAServices.createSigningIdentity(user)
	if err != nil {
		logger.Infof("Invalid re-enroll request, %s is not a valid user  %s\n", user.GetName(), err)
		return nil, nil, fmt.Errorf("Reenroll has failed; Cannot create user identity: %s", err)
	}

	if identity.GetECert() == nil {
		logger.Infof("Invalid re-enroll request for user '%s'. Enrollment cert does not exist %s\n", user.GetName(), err)
		return nil, nil, fmt.Errorf("Reenroll has failed; enrollment cert does not exist: %s", err)
	}

	reenrollmentResponse, err := identity.Reenroll(req)
	if err != nil {
		return nil, nil, fmt.Errorf("ReEnroll failed: %s", err)
	}
	return reenrollmentResponse.Identity.GetECert().Key(), reenrollmentResponse.Identity.GetECert().Cert(), nil
}

// Register a User with the Fabric CA
// @param {User} registrar The User that is initiating the registration
// @param {RegistrationRequest} request Registration Request
// @returns {string} Enrolment Secret
// @returns {error} Error
func (fabricCAServices *services) Register(registrar sdkApi.User,
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
// @param {User} registrar The User that is initiating the revocation
// @param {RevocationRequest} request Revocation Request
// @returns {error} Error
func (fabricCAServices *services) Revoke(registrar sdkApi.User,
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
func (fabricCAServices *services) createSigningIdentity(user sdkApi.
	User) (*fabric_ca.Identity, error) {
	// Validate user
	if user == nil {
		return nil, fmt.Errorf("Valid user required to create signing identity")
	}
	// Validate enrolment information
	cert := user.GetEnrollmentCertificate()
	key := user.GetPrivateKey()
	if key == nil || cert == nil {
		return nil, fmt.Errorf(
			"Unable to read user enrolment information to create signing identity")
	}
	return fabricCAServices.fabricCAClient.NewIdentity(key, cert)
}
