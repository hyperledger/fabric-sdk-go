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
	caName          string
	orgMSPID        string
	cryptoSuite     core.CryptoSuite
	identityManager msp.IdentityManager
	userStore       msp.UserStore
	adapter         *fabricCAAdapter
	registrar       msp.EnrollCredentials
}

// CAClientOption describes a functional parameter for NewCAClient
type CAClientOption func(*caClientOption) error

type caClientOption struct {
	caID string
}

// WithCAInstance allows for specifying optional CA name (within the CA server instance)
func WithCAInstance(caID string) CAClientOption {
	return func(o *caClientOption) error {
		o.caID = caID
		return nil
	}
}

// NewCAClient creates a new CA CAClient instance
func NewCAClient(orgName string, ctx contextApi.Client, opts ...CAClientOption) (*CAClientImpl, error) {

	if orgName == "" {
		orgName = ctx.IdentityConfig().Client().Organization
	}

	if orgName == "" {
		return nil, errors.New("organization is missing")
	}

	options, err := processCAClientOptions(opts...)
	if err != nil {
		return nil, err
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

	caID := options.caID
	if caID == "" {
		caID = orgConfig.CertificateAuthorities[0]
	}
	caConfig, ok := ctx.IdentityConfig().CAConfig(caID)
	if !ok {
		return nil, errors.Errorf("error initializing CA [%s]", caID)
	}
	adapter, err := newFabricCAAdapter(caID, ctx.CryptoSuite(), ctx.IdentityConfig())
	if err != nil {
		return nil, errors.Wrapf(err, "error initializing CA [%s]", caID)
	}

	identityManager, ok := ctx.IdentityManager(orgName)
	if !ok {
		return nil, fmt.Errorf("identity manager not found for organization '%s", orgName)
	}

	mgr := &CAClientImpl{
		orgName:         orgName,
		caName:          caConfig.CAName,
		orgMSPID:        orgConfig.MSPID,
		cryptoSuite:     ctx.CryptoSuite(),
		identityManager: identityManager,
		userStore:       ctx.UserStore(),
		adapter:         adapter,
		registrar:       caConfig.Registrar,
	}
	return mgr, nil
}

func processCAClientOptions(opts ...CAClientOption) (*caClientOption, error) {
	options := caClientOption{}

	for _, param := range opts {
		err := param(&options)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to create CA Client")
		}
	}
	return &options, nil
}

// Enroll a registered user in order to receive a signed X509 certificate.
// A new key pair is generated for the user. The private key and the
// enrollment certificate issued by the CA are stored in SDK stores.
// They can be retrieved by calling IdentityManager.GetSigningIdentity().
//
// enrollmentID The registered ID to use for enrollment
// enrollmentSecret The secret associated with the enrollment ID
func (c *CAClientImpl) Enroll(request *api.EnrollmentRequest) error {

	if c.adapter == nil {
		return fmt.Errorf("no CAs configured for organization: %s", c.orgName)
	}
	if request.Name == "" {
		return errors.New("enrollmentID is required")
	}
	if request.Secret == "" {
		return errors.New("enrollmentSecret is required")
	}
	// TODO add attributes
	cert, err := c.adapter.Enroll(request)
	if err != nil {
		return errors.Wrap(err, "enroll failed")
	}
	userData := &msp.UserData{
		MSPID:                 c.orgMSPID,
		ID:                    request.Name,
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

	// Check required parameters (ID)
	if request.ID == "" {
		return nil, errors.New("ID is required")
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

	// Check required parameters (ID)
	if request.ID == "" {
		return nil, errors.New("ID is required")
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

	// Check required parameters (ID)
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

	// Check required parameters (ID and affiliation)
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
func (c *CAClientImpl) Reenroll(request *api.ReenrollmentRequest) error {

	if c.adapter == nil {
		return fmt.Errorf("no CAs configured for organization: %s", c.orgName)
	}
	if request.Name == "" {
		logger.Info("invalid re-enroll request, missing enrollmentID")
		return errors.New("user name missing")
	}

	user, err := c.identityManager.GetSigningIdentity(request.Name)
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve user: %s", request.Name)
	}

	cert, err := c.adapter.Reenroll(user.PrivateKey(), user.EnrollmentCertificate(), request)
	if err != nil {
		return errors.Wrap(err, "reenroll failed")
	}
	userData := &msp.UserData{
		MSPID:                 c.orgMSPID,
		ID:                    user.Identifier().ID,
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

// GetCAInfo returns generic CA information
func (c *CAClientImpl) GetCAInfo() (*api.GetCAInfoResponse, error) {
	if c.adapter == nil {
		return nil, fmt.Errorf("no CAs configured for organization: %s", c.orgName)
	}

	return c.adapter.GetCAInfo(c.caName)
}

// GetAffiliation returns information about the requested affiliation
func (c *CAClientImpl) GetAffiliation(affiliation, caname string) (*api.AffiliationResponse, error) {
	if c.adapter == nil {
		return nil, fmt.Errorf("no CAs configured for organization: %s", c.orgName)
	}

	// Check required parameters (affiliation)
	if affiliation == "" {
		return nil, errors.New("affiliation is required")
	}

	registrar, err := c.getRegistrar(c.registrar.EnrollID, c.registrar.EnrollSecret)
	if err != nil {
		return nil, err
	}

	return c.adapter.GetAffiliation(registrar.PrivateKey(), registrar.EnrollmentCertificate(), affiliation, caname)
}

// GetAllAffiliations returns all affiliations that the caller is authorized to see
func (c *CAClientImpl) GetAllAffiliations(caname string) (*api.AffiliationResponse, error) {
	if c.adapter == nil {
		return nil, fmt.Errorf("no CAs configured for organization %s", c.orgName)
	}

	registrar, err := c.getRegistrar(c.registrar.EnrollID, c.registrar.EnrollSecret)
	if err != nil {
		return nil, err
	}

	return c.adapter.GetAllAffiliations(registrar.PrivateKey(), registrar.EnrollmentCertificate(), caname)
}

// AddAffiliation adds a new affiliation to the server
func (c *CAClientImpl) AddAffiliation(request *api.AffiliationRequest) (*api.AffiliationResponse, error) {
	if c.adapter == nil {
		return nil, fmt.Errorf("no CAs configured for organization: %s", c.orgName)
	}

	if request == nil {
		return nil, errors.New("must provide affiliation request")
	}

	// Check required parameters (Name)
	if request.Name == "" {
		return nil, errors.New("Name is required")
	}

	registrar, err := c.getRegistrar(c.registrar.EnrollID, c.registrar.EnrollSecret)
	if err != nil {
		return nil, err
	}

	return c.adapter.AddAffiliation(registrar.PrivateKey(), registrar.EnrollmentCertificate(), request)
}

// ModifyAffiliation renames an existing affiliation on the server
func (c *CAClientImpl) ModifyAffiliation(request *api.ModifyAffiliationRequest) (*api.AffiliationResponse, error) {
	if c.adapter == nil {
		return nil, fmt.Errorf("no CAs configured for organization: %s", c.orgName)
	}

	if request == nil {
		return nil, errors.New("must provide affiliation request")
	}

	// Check required parameters (Name and NewName)
	if request.Name == "" || request.NewName == "" {
		return nil, errors.New("Name and NewName are required")
	}

	registrar, err := c.getRegistrar(c.registrar.EnrollID, c.registrar.EnrollSecret)
	if err != nil {
		return nil, err
	}

	return c.adapter.ModifyAffiliation(registrar.PrivateKey(), registrar.EnrollmentCertificate(), request)
}

// RemoveAffiliation removes an existing affiliation from the server
func (c *CAClientImpl) RemoveAffiliation(request *api.AffiliationRequest) (*api.AffiliationResponse, error) {
	if c.adapter == nil {
		return nil, fmt.Errorf("no CAs configured for organization: %s", c.orgName)
	}

	if request == nil {
		return nil, errors.New("must provide remove affiliation request")
	}

	// Check required parameters (Name)
	if request.Name == "" {
		return nil, errors.New("Name is required")
	}

	registrar, err := c.getRegistrar(c.registrar.EnrollID, c.registrar.EnrollSecret)
	if err != nil {
		return nil, err
	}

	return c.adapter.RemoveAffiliation(registrar.PrivateKey(), registrar.EnrollmentCertificate(), request)
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
		err = c.Enroll(&api.EnrollmentRequest{Name: enrollID, Secret: enrollSecret})
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
