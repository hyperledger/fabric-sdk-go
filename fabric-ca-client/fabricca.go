/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at


      http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fabricca

import (
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-ca/api"
	fabric_ca "github.com/hyperledger/fabric-ca/lib"
	"github.com/hyperledger/fabric-sdk-go/config"
	fabricclient "github.com/hyperledger/fabric-sdk-go/fabric-client"

	"io/ioutil"

	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("fabric_sdk_go")

// Services ...
type Services interface {
	Enroll(enrollmentID string, enrollmentSecret string) ([]byte, []byte, error)
	//reenroll  to renew user's enrollment certificate
	Reenroll(user fabricclient.User) ([]byte, []byte, error)
	Register(registrar fabricclient.User, request *RegistrationRequest) (string, error)
	Revoke(registrar fabricclient.User, request *RevocationRequest) error
}

type services struct {
	fabricCAClient *fabric_ca.Client
}

// RegistrationRequest defines the attributes required to register a user with the CA
type RegistrationRequest struct {
	// Name is the unique name of the identity
	Name string
	// Type of identity being registered (e.g. "peer, app, user")
	Type string
	// MaxEnrollments is the number of times the secret can  be reused to enroll.
	// if omitted, this defaults to max_enrollments configured on the server
	MaxEnrollments int
	// The identity's affiliation e.g. org1.department1
	Affiliation string
	// Optional attributes associated with this identity
	Attributes []Attribute
}

// RevocationRequest defines the attributes required to revoke credentials with the CA
type RevocationRequest struct {
	// Name of the identity whose certificates should be revoked
	// If this field is omitted, then Serial and AKI must be specified.
	Name string
	// Serial number of the certificate to be revoked
	// If this is omitted, then Name must be specified
	Serial string
	// AKI (Authority Key Identifier) of the certificate to be revoked
	AKI string
	// Reason is the reason for revocation. See https://godoc.org/golang.org/x/crypto/ocsp
	// for valid values. The default value is 0 (ocsp.Unspecified).
	Reason int
}

// Attribute defines additional attributes that may be passed along during registration
type Attribute struct {
	Key   string
	Value string
}

// NewFabricCAClient ...
/**
 * @param {string} clientConfigFile for fabric-ca services"
 */
func NewFabricCAClient() (Services, error) {

	// Create new Fabric-ca client without configs
	c, err := fabric_ca.NewClient("")
	if err != nil {
		return nil, fmt.Errorf("New fabricCAClient failed: %s", err)
	}

	certFile := config.GetFabricCAClientCertFile()
	keyFile := config.GetFabricCAClientKeyFile()
	serverCertFiles := config.GetServerCertFiles()

	//set server URL
	c.Config.URL = config.GetServerURL()
	//certs file list
	c.Config.TLS.CertFilesList = serverCertFiles
	//concat cert files
	c.Config.TLS.CertFiles = strings.Join(serverCertFiles[:], ",")
	//set cert file into TLS context
	file, err := ioutil.ReadFile(certFile)
	if err != nil {
		logger.Errorf("Error reading fabric ca client propertiy certfile: %v", err)
		return nil, fmt.Errorf("New fabricCAClient failed: %s", err)
	}
	c.Config.TLS.Client.CertFile = string(file)
	//set key file into TLS context
	keyfile, err := ioutil.ReadFile(keyFile)
	if err != nil {
		logger.Errorf("Error reading fabric ca client property keyfile: %v", err)
		return nil, fmt.Errorf("New fabricCAClient failed: %s", err)
	}
	c.Config.TLS.Client.KeyFile = string(keyfile)

	//TLS falg enabled/disabled
	c.Config.TLS.Enabled = config.GetFabricCATLSEnabledFlag()
	fabricCAClient := &services{fabricCAClient: c}
	logger.Infof("Constructed fabricCAClient instance: %v", fabricCAClient)

	return fabricCAClient, nil
}

// Enroll ...
/**
 * Enroll a registered user in order to receive a signed X509 certificate
 * @param {string} enrollmentID The registered ID to use for enrollment
 * @param {string} enrollmentSecret The secret associated with the enrollment ID
 * @returns {[]byte} X509 certificate
 * @returns {[]byte} private key
 */
func (fabricCAServices *services) Enroll(enrollmentID string, enrollmentSecret string) ([]byte, []byte, error) {
	if enrollmentID == "" {
		return nil, nil, fmt.Errorf("enrollmentID is empty")
	}
	if enrollmentSecret == "" {
		return nil, nil, fmt.Errorf("enrollmentSecret is empty")
	}
	req := &api.EnrollmentRequest{
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
 * @param {user} fabricclient.User to be reenrolled
 * @returns {[]byte} X509 certificate
 * @returns {[]byte} private key
 */
func (fabricCAServices *services) Reenroll(user fabricclient.User) ([]byte, []byte, error) {
	if user == nil {
		return nil, nil, fmt.Errorf("User does not exist")
	}
	if user.GetName() == "" {
		logger.Infof("Invalid re-enroll request, missing argument user")
		return nil, nil, fmt.Errorf("User is empty")
	}
	req := &api.ReenrollmentRequest{}
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
func (fabricCAServices *services) Register(registrar fabricclient.User,
	request *RegistrationRequest) (string, error) {
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
		Name:           request.Name,
		Type:           request.Type,
		MaxEnrollments: request.MaxEnrollments,
		Affiliation:    request.Affiliation,
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
func (fabricCAServices *services) Revoke(registrar fabricclient.User,
	request *RevocationRequest) error {
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
		Name:   request.Name,
		Serial: request.Serial,
		AKI:    request.AKI,
		Reason: request.Reason}
	return identity.Revoke(&req)
}

// createSigningIdentity creates an identity to sign Fabric CA requests with
func (fabricCAServices *services) createSigningIdentity(user fabricclient.
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
	// TODO: Right now this reads the key from a default BCCSP implementation using the SKI
	// this method signature will change to accepting a BCCSP key soon.
	// Track changes here: https://gerrit.hyperledger.org/r/#/c/6727/
	ski := key.SKI()
	if ski == nil {
		return nil, fmt.Errorf("Unable to read private key SKI")
	}
	return fabricCAServices.fabricCAClient.NewIdentity(ski, cert)
}
