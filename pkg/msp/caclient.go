/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"fmt"

	"strings"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/identitymgr"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/ca")

// CAClient implements api/msp/CAClient
type CAClient struct {
	orgName         string
	orgMspID        string
	caName          string
	config          core.Config
	cryptoSuite     core.CryptoSuite
	identityManager msp.IdentityManager
	userStore       msp.UserStore
	adapter         *FabricCAAdapter
	registrar       core.EnrollCredentials
}

// New creates a new CA CAClient instance
func New(orgName string, identityManager msp.IdentityManager, stateStore core.KVStore, cryptoSuite core.CryptoSuite, config core.Config) (*CAClient, error) {

	userStore, err := identitymgr.NewCertFileUserStore1(stateStore)
	if err != nil {
		return nil, errors.Wrapf(err, "creating a user store failed")
	}

	netConfig, err := config.NetworkConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "network config retrieval failed")
	}

	if orgName == "" {
		clientConfig, err := config.Client()
		if err != nil {
			return nil, errors.Wrapf(err, "client config retrieval failed")
		}
		orgName = clientConfig.Organization
	}

	if orgName == "" {
		return nil, errors.New("organization is missing")
	}

	// viper keys are case insensitive
	orgConfig, ok := netConfig.Organizations[strings.ToLower(orgName)]
	if !ok {
		return nil, errors.New("org config retrieval failed")
	}

	var caName string
	var caConfig *core.CAConfig
	var adapter *FabricCAAdapter
	var registrar core.EnrollCredentials
	if len(orgConfig.CertificateAuthorities) == 0 {
		logger.Warnln("no CAs configured")
	} else {
		// Currently, an organization can be associated with only one CA
		caName = orgConfig.CertificateAuthorities[0]
		caConfig, err = config.CAConfig(orgName)
		if err == nil {
			adapter, err = newFabricCAAdapter(orgName, caName, cryptoSuite, config)
			if err == nil {
				registrar = caConfig.Registrar
			} else {
				return nil, errors.Wrapf(err, "error initializing CA [%s]", caName)
			}
		} else {
			return nil, errors.Wrapf(err, "error initializing CA [%s]", caName)
		}
	}

	mgr := &CAClient{
		orgName:         orgName,
		orgMspID:        orgConfig.MspID,
		caName:          caName,
		config:          config,
		cryptoSuite:     cryptoSuite,
		identityManager: identityManager,
		userStore:       userStore,
		adapter:         adapter,
		registrar:       registrar,
	}
	return mgr, nil
}

// CAName returns the CA name.
func (c *CAClient) CAName() string {
	return c.caName
}

// Enroll a registered user in order to receive a signed X509 certificate.
// A new key pair is generated for the user. The private key and the
// enrollment certificate issued by the CA are stored in SDK stores.
// They can be retrieved by calling IdentityManager.GetSigningIdentity().
//
// enrollmentID The registered ID to use for enrollment
// enrollmentSecret The secret associated with the enrollment ID
func (c *CAClient) Enroll(enrollmentID string, enrollmentSecret string) error {

	if c.adapter == nil {
		return fmt.Errorf("no CAs configured for organization: %s", c.orgName)
	}
	if enrollmentID == "" {
		return errors.New("enrollmentID is required")
	}
	if enrollmentSecret == "" {
		return errors.New("enrollmentSecret is required")
	}
	// TODO add attributes
	careq := &api.EnrollmentRequest{
		CAName: c.caName,
		Name:   enrollmentID,
		Secret: enrollmentSecret,
	}
	cert, err := c.adapter.Enroll(careq)
	if err != nil {
		return errors.Wrap(err, "enroll failed")
	}
	userData := msp.UserData{
		MspID: c.orgMspID,
		Name:  enrollmentID,
		EnrollmentCertificate: cert,
	}
	err = c.userStore.Store(userData)
	if err != nil {
		return errors.Wrap(err, "enroll failed")
	}
	return nil
}

// Reenroll an enrolled user in order to obtain a new signed X509 certificate
func (c *CAClient) Reenroll(enrollmentID string) error {

	if c.adapter == nil {
		return fmt.Errorf("no CAs configured for organization: %s", c.orgName)
	}
	if enrollmentID == "" {
		logger.Infof("invalid re-enroll request, missing enrollmentID")
		return errors.New("user name missing")
	}
	req := &api.ReenrollmentRequest{
		CAName: c.adapter.CAName(),
	}

	user, err := c.identityManager.GetUser(enrollmentID)
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve user: %s", enrollmentID)
	}

	cert, err := c.adapter.Reenroll(user.PrivateKey(), user.EnrollmentCertificate(), req)
	if err != nil {
		return errors.Wrap(err, "reenroll failed")
	}
	userData := msp.UserData{
		MspID: c.orgMspID,
		Name:  user.Name(),
		EnrollmentCertificate: cert,
	}
	err = c.userStore.Store(userData)
	if err != nil {
		return errors.Wrap(err, "reenroll failed")
	}

	return nil
}

// Register a User with the Fabric CA
// request: Registration Request
// Returns Enrolment Secret
func (c *CAClient) Register(request *msp.RegistrationRequest) (string, error) {
	if c.adapter == nil {
		return "", fmt.Errorf("no CAs configured for organization: %s", c.orgName)
	}
	if c.registrar.EnrollID == "" {
		return "", msp.ErrCARegistrarNotFound
	}
	// Validate registration request
	if request == nil {
		return "", errors.New("registration request is required")
	}
	if request.Name == "" {
		return "", errors.New("request.Name is required")
	}

	registrar, err := c.getRegistrar(c.registrar.EnrollID, c.registrar.EnrollSecret)
	if err != nil {
		return "", err
	}

	secret, err := c.adapter.Register(registrar.PrivateKey, registrar.EnrollmentCert, request)
	if err != nil {
		return "", errors.Wrap(err, "failed to register user")
	}

	return secret, nil
}

// Revoke a User with the Fabric CA
// registrar: The User that is initiating the revocation
// request: Revocation Request
func (c *CAClient) Revoke(request *msp.RevocationRequest) (*msp.RevocationResponse, error) {
	if c.adapter == nil {
		return nil, fmt.Errorf("no CAs configured for organization: %s", c.orgName)
	}
	if c.registrar.EnrollID == "" {
		return nil, msp.ErrCARegistrarNotFound
	}
	// Validate revocation request
	if request == nil {
		return nil, errors.New("revocation request is required")
	}

	registrar, err := c.getRegistrar(c.registrar.EnrollID, c.registrar.EnrollSecret)
	if err != nil {
		return nil, err
	}

	resp, err := c.adapter.Revoke(registrar.PrivateKey, registrar.EnrollmentCert, request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to revoke")
	}
	return resp, nil
}

func (c *CAClient) getRegistrar(enrollID string, enrollSecret string) (*msp.SigningIdentity, error) {

	if enrollID == "" {
		return nil, msp.ErrCARegistrarNotFound
	}

	registrar, err := c.identityManager.GetSigningIdentity(enrollID)
	if err != nil {
		if err != msp.ErrUserNotFound {
			return nil, err
		}
		if enrollSecret == "" {
			return nil, msp.ErrCARegistrarNotFound
		}

		// Attempt to enroll the registrar
		err = c.Enroll(enrollID, enrollSecret)
		if err != nil {
			return nil, err
		}
		registrar, err = c.identityManager.GetSigningIdentity(enrollID)
		if err != nil {
			return nil, err
		}
	}
	return registrar, nil
}
