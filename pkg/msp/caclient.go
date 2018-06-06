/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"fmt"

	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp/api"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/msp")

// CAClientImpl implements api/msp/CAClient
type CAClientImpl struct {
	orgName         string
	orgMSPID        string
	cryptoSuite     core.CryptoSuite
	identityManager msp.IdentityManager
	userStore       msp.UserStore
	adapter         *fabricCAAdapter
	registrar       msp.EnrollCredentials
}

// NewCAClient creates a new CA CAClient instance
func NewCAClient(orgName string, ctx contextApi.Client) (*CAClientImpl, error) {

	if orgName == "" {
		orgName = ctx.IdentityConfig().Client().Organization
	}

	if orgName == "" {
		return nil, errors.New("organization is missing")
	}

	netConfig := ctx.EndpointConfig().NetworkConfig()
	// viper keys are case insensitive
	orgConfig, ok := netConfig.Organizations[strings.ToLower(orgName)]
	if !ok {
		return nil, errors.New("org config retrieval failed")
	}
	if len(orgConfig.CertificateAuthorities) == 0 {
		return nil, errors.New("no CAs configured")
	}

	var adapter *fabricCAAdapter
	var registrar msp.EnrollCredentials
	var err error

	// Currently, an organization can be associated with only one CA
	caName := orgConfig.CertificateAuthorities[0]
	caConfig, ok := ctx.IdentityConfig().CAConfig(orgName)
	if ok {
		adapter, err = newFabricCAAdapter(orgName, ctx.CryptoSuite(), ctx.IdentityConfig())
		if err == nil {
			registrar = caConfig.Registrar
		} else {
			return nil, errors.Wrapf(err, "error initializing CA [%s]", caName)
		}
	} else {
		return nil, errors.Errorf("error initializing CA [%s]", caName)
	}

	identityManager, ok := ctx.IdentityManager(orgName)
	if !ok {
		return nil, fmt.Errorf("identity manager not found for organization '%s", orgName)
	}

	mgr := &CAClientImpl{
		orgName:         orgName,
		orgMSPID:        orgConfig.MSPID,
		cryptoSuite:     ctx.CryptoSuite(),
		identityManager: identityManager,
		userStore:       ctx.UserStore(),
		adapter:         adapter,
		registrar:       registrar,
	}
	return mgr, nil
}

// Enroll a registered user in order to receive a signed X509 certificate.
// A new key pair is generated for the user. The private key and the
// enrollment certificate issued by the CA are stored in SDK stores.
// They can be retrieved by calling IdentityManager.GetSigningIdentity().
//
// enrollmentID The registered ID to use for enrollment
// enrollmentSecret The secret associated with the enrollment ID
func (c *CAClientImpl) Enroll(enrollmentID string, enrollmentSecret string) error {

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
	cert, err := c.adapter.Enroll(enrollmentID, enrollmentSecret)
	if err != nil {
		return errors.Wrap(err, "enroll failed")
	}
	userData := &msp.UserData{
		MSPID: c.orgMSPID,
		ID:    enrollmentID,
		EnrollmentCertificate: cert,
	}
	err = c.userStore.Store(userData)
	if err != nil {
		return errors.Wrap(err, "enroll failed")
	}
	return nil
}

// CreateIdentity create a new identity with the Fabric CA server. An enrollment secret is returned which can then be used,
// along with the enrollment ID, to enroll a new identity.
//  Parameters:
//  request holds info about identity
//
//  Returns:
//  Return identity info including secret
func (c *CAClientImpl) CreateIdentity(request *api.IdentityRequest) (*api.IdentityResponse, error) {

	if c.adapter == nil {
		return nil, fmt.Errorf("no CAs configured for organization: %s", c.orgName)
	}

	if request == nil {
		return nil, errors.New("must provide identity request")
	}

	// Checke required parameters (ID and affiliation)
	if request.ID == "" || request.Affiliation == "" {
		return nil, errors.New("ID and affiliation are required")
	}

	registrar, err := c.getRegistrar(c.registrar.EnrollID, c.registrar.EnrollSecret)
	if err != nil {
		return nil, err
	}

	return c.adapter.CreateIdentity(registrar.PrivateKey(), registrar.EnrollmentCertificate(), request)
}

// ModifyIdentity modifies identity with the Fabric CA server.
//  Parameters:
//  request holds info about identity
//
//  Returns:
//  Return modified identity info
func (c *CAClientImpl) ModifyIdentity(request *api.IdentityRequest) (*api.IdentityResponse, error) {

	if c.adapter == nil {
		return nil, fmt.Errorf("no CAs configured for organization: %s", c.orgName)
	}

	if request == nil {
		return nil, errors.New("must provide identity request")
	}

	// Checke required parameters (ID and affiliation)
	if request.ID == "" || request.Affiliation == "" {
		return nil, errors.New("ID and affiliation are required")
	}

	registrar, err := c.getRegistrar(c.registrar.EnrollID, c.registrar.EnrollSecret)
	if err != nil {
		return nil, err
	}

	return c.adapter.ModifyIdentity(registrar.PrivateKey(), registrar.EnrollmentCertificate(), request)
}

// RemoveIdentity removes identity from the Fabric CA server.
//  Parameters:
//  request holds info about identity to be removed
//
//  Returns:
//  Return removed identity info
func (c *CAClientImpl) RemoveIdentity(request *api.RemoveIdentityRequest) (*api.IdentityResponse, error) {

	if c.adapter == nil {
		return nil, fmt.Errorf("no CAs configured for organization: %s", c.orgName)
	}

	if request == nil {
		return nil, errors.New("must provide remove identity request")
	}

	// Checke required parameters (ID)
	if request.ID == "" {
		return nil, errors.New("ID is required")
	}

	registrar, err := c.getRegistrar(c.registrar.EnrollID, c.registrar.EnrollSecret)
	if err != nil {
		return nil, err
	}

	return c.adapter.RemoveIdentity(registrar.PrivateKey(), registrar.EnrollmentCertificate(), request)

}

// GetIdentity retrieves identity information.
//  Parameters:
//  id is required identity id
//
//  Returns:
//  Returns identity information
func (c *CAClientImpl) GetIdentity(id, caname string) (*api.IdentityResponse, error) {

	if c.adapter == nil {
		return nil, fmt.Errorf("no CAs configured for organization: %s", c.orgName)
	}

	// Checke required parameters (ID and affiliation)
	if id == "" {
		return nil, errors.New("id is required")
	}

	registrar, err := c.getRegistrar(c.registrar.EnrollID, c.registrar.EnrollSecret)
	if err != nil {
		return nil, err
	}

	return c.adapter.GetIdentity(registrar.PrivateKey(), registrar.EnrollmentCertificate(), id, caname)
}

// GetAllIdentities returns all identities that the caller is authorized to see
//
//  Returns:
//  Response containing identities
func (c *CAClientImpl) GetAllIdentities(caname string) ([]*api.IdentityResponse, error) {

	if c.adapter == nil {
		return nil, fmt.Errorf("no CAs configured for organization: %s", c.orgName)
	}

	registrar, err := c.getRegistrar(c.registrar.EnrollID, c.registrar.EnrollSecret)
	if err != nil {
		return nil, err
	}

	return c.adapter.GetAllIdentities(registrar.PrivateKey(), registrar.EnrollmentCertificate(), caname)
}

// Reenroll an enrolled user in order to obtain a new signed X509 certificate
func (c *CAClientImpl) Reenroll(enrollmentID string) error {

	if c.adapter == nil {
		return fmt.Errorf("no CAs configured for organization: %s", c.orgName)
	}
	if enrollmentID == "" {
		logger.Info("invalid re-enroll request, missing enrollmentID")
		return errors.New("user name missing")
	}

	user, err := c.identityManager.GetSigningIdentity(enrollmentID)
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve user: %s", enrollmentID)
	}

	cert, err := c.adapter.Reenroll(user.PrivateKey(), user.EnrollmentCertificate())
	if err != nil {
		return errors.Wrap(err, "reenroll failed")
	}
	userData := &msp.UserData{
		MSPID: c.orgMSPID,
		ID:    user.Identifier().ID,
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
func (c *CAClientImpl) Register(request *api.RegistrationRequest) (string, error) {
	if c.adapter == nil {
		return "", fmt.Errorf("no CAs configured for organization: %s", c.orgName)
	}
	if c.registrar.EnrollID == "" {
		return "", api.ErrCARegistrarNotFound
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

	secret, err := c.adapter.Register(registrar.PrivateKey(), registrar.EnrollmentCertificate(), request)
	if err != nil {
		return "", errors.Wrap(err, "failed to register user")
	}

	return secret, nil
}

// Revoke a User with the Fabric CA
// registrar: The User that is initiating the revocation
// request: Revocation Request
func (c *CAClientImpl) Revoke(request *api.RevocationRequest) (*api.RevocationResponse, error) {
	if c.adapter == nil {
		return nil, fmt.Errorf("no CAs configured for organization: %s", c.orgName)
	}
	if c.registrar.EnrollID == "" {
		return nil, api.ErrCARegistrarNotFound
	}
	// Validate revocation request
	if request == nil {
		return nil, errors.New("revocation request is required")
	}

	registrar, err := c.getRegistrar(c.registrar.EnrollID, c.registrar.EnrollSecret)
	if err != nil {
		return nil, err
	}

	resp, err := c.adapter.Revoke(registrar.PrivateKey(), registrar.EnrollmentCertificate(), request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to revoke")
	}
	return resp, nil
}

func (c *CAClientImpl) getRegistrar(enrollID string, enrollSecret string) (msp.SigningIdentity, error) {

	if enrollID == "" {
		return nil, api.ErrCARegistrarNotFound
	}

	registrar, err := c.identityManager.GetSigningIdentity(enrollID)
	if err != nil {
		if err != msp.ErrUserNotFound {
			return nil, err
		}
		if enrollSecret == "" {
			return nil, api.ErrCARegistrarNotFound
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
