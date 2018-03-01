/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identitymgr

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	caapi "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/api"
	calib "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/lib"
	config "github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/identitymgr/persistence"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"

	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
)

var logger = logging.NewLogger("fabric_sdk_go")

// IdentityManager implements fab/IdentityManager
type IdentityManager struct {
	orgName         string
	orgMspID        string
	caName          string
	config          core.Config
	cryptoSuite     core.CryptoSuite
	embeddedUsers   map[string]core.TLSKeyPair
	mspPrivKeyStore contextApi.KVStore
	mspCertStore    contextApi.KVStore
	userStore       contextApi.UserStore

	// CA Client state
	caClient  *calib.Client
	registrar core.EnrollCredentials
}

// New creates a new instance of IdentityManager
// @param {string} organization for this CA
// @param {Config} client config for fabric-ca services
// @returns {IdentityManager} IdentityManager instance
// @returns {error} error, if any
func New(orgName string, cryptoSuite core.CryptoSuite, config config.Config) (*IdentityManager, error) {

	netConfig, err := config.NetworkConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "network config retrieval failed")
	}

	// viper keys are case insensitive
	orgConfig, ok := netConfig.Organizations[strings.ToLower(orgName)]
	if !ok {
		return nil, errors.New("org config retrieval failed")
	}

	if orgConfig.CryptoPath == "" && len(orgConfig.Users) == 0 {
		return nil, errors.New("Either a cryptopath or an embedded list of users is required")
	}

	var mspPrivKeyStore contextApi.KVStore
	var mspCertStore contextApi.KVStore

	orgCryptoPathTemplate := orgConfig.CryptoPath
	if orgCryptoPathTemplate != "" {
		if !filepath.IsAbs(orgCryptoPathTemplate) {
			orgCryptoPathTemplate = filepath.Join(config.CryptoConfigPath(), orgCryptoPathTemplate)
		}
		mspPrivKeyStore, err = persistence.NewFileKeyStore(orgCryptoPathTemplate)
		if err != nil {
			return nil, errors.Wrapf(err, "creating a private key store failed")
		}
		mspCertStore, err = persistence.NewFileCertStore(orgCryptoPathTemplate)
		if err != nil {
			return nil, errors.Wrapf(err, "creating a cert store failed")
		}
	} else {
		logger.Warnf("Cryptopath not provided for organization [%s], MSP stores not created", orgName)
	}

	// In the future, shared UserStore from the SDK context will be used
	var userStore contextApi.UserStore
	if config.CredentialStorePath() != "" {
		userStore, err = identity.NewCertFileUserStore(config.CredentialStorePath(), cryptoSuite)
		if err != nil {
			return nil, errors.Wrapf(err, "creating a user store failed")
		}
	}

	var caName string
	if len(orgConfig.CertificateAuthorities) > 0 {
		caName = orgConfig.CertificateAuthorities[0]
	}

	mgr := &IdentityManager{
		orgName:         orgName,
		orgMspID:        orgConfig.MspID,
		caName:          caName,
		config:          config,
		cryptoSuite:     cryptoSuite,
		mspPrivKeyStore: mspPrivKeyStore,
		mspCertStore:    mspCertStore,
		embeddedUsers:   orgConfig.Users,
		userStore:       userStore,
		// CA Client state is created lazily, when (if) needed
	}
	return mgr, nil
}

// CAName returns the CA name.
func (im *IdentityManager) CAName() string {
	return im.caName
}

// Enroll a registered user in order to receive a signed X509 certificate.
// enrollmentID The registered ID to use for enrollment
// enrollmentSecret The secret associated with the enrollment ID
// Returns X509 certificate
func (im *IdentityManager) Enroll(enrollmentID string, enrollmentSecret string) (core.Key, []byte, error) {

	if err := im.initCAClient(); err != nil {
		return nil, nil, err
	}
	if enrollmentID == "" {
		return nil, nil, errors.New("enrollmentID is required")
	}
	if enrollmentSecret == "" {
		return nil, nil, errors.New("enrollmentSecret is required")
	}
	// TODO add attributes
	careq := &caapi.EnrollmentRequest{
		CAName: im.caClient.Config.CAName,
		Name:   enrollmentID,
		Secret: enrollmentSecret,
	}
	caresp, err := im.caClient.Enroll(careq)
	if err != nil {
		return nil, nil, errors.Wrap(err, "enroll failed")
	}
	user := identity.NewUser(im.orgMspID, enrollmentID)
	user.SetEnrollmentCertificate(caresp.Identity.GetECert().Cert())
	user.SetPrivateKey(caresp.Identity.GetECert().Key())
	err = im.userStore.Store(user)
	if err != nil {
		return nil, nil, errors.Wrap(err, "enroll failed")
	}
	return caresp.Identity.GetECert().Key(), caresp.Identity.GetECert().Cert(), nil
}

// Reenroll an enrolled user in order to receive a signed X509 certificate
// Returns X509 certificate
func (im *IdentityManager) Reenroll(user contextApi.User) (core.Key, []byte, error) {

	if err := im.initCAClient(); err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, errors.New("user required")
	}
	if user.Name() == "" {
		logger.Infof("Invalid re-enroll request, missing argument user")
		return nil, nil, errors.New("user name missing")
	}
	req := &caapi.ReenrollmentRequest{
		CAName: im.caClient.Config.CAName,
	}
	caidentity, err := im.caClient.NewIdentity(user.PrivateKey(), user.EnrollmentCertificate())
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create CA signing identity")
	}

	caresp, err := caidentity.Reenroll(req)
	if err != nil {
		return nil, nil, errors.Wrap(err, "reenroll failed")
	}
	newUser := identity.NewUser(im.orgMspID, user.Name())
	newUser.SetEnrollmentCertificate(caresp.Identity.GetECert().Cert())
	newUser.SetPrivateKey(caresp.Identity.GetECert().Key())
	err = im.userStore.Store(newUser)
	if err != nil {
		return nil, nil, errors.Wrap(err, "reenroll failed")
	}

	return caresp.Identity.GetECert().Key(), caresp.Identity.GetECert().Cert(), nil
}

// Register a User with the Fabric CA
// request: Registration Request
// Returns Enrolment Secret
func (im *IdentityManager) Register(request *contextApi.RegistrationRequest) (string, error) {
	if err := im.initCAClient(); err != nil {
		return "", err
	}
	if im.registrar.EnrollID == "" {
		return "", contextApi.ErrCARegistrarNotFound
	}
	// Validate registration request
	if request == nil {
		return "", errors.New("registration request is required")
	}
	if request.Name == "" {
		return "", errors.New("request.Name is required")
	}
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

	registrar, err := im.getRegistrarSI(im.registrar.EnrollID, im.registrar.EnrollSecret)
	if err != nil {
		return "", err
	}

	response, err := registrar.Register(&req)
	if err != nil {
		return "", errors.Wrap(err, "failed to register user")
	}

	return response.Secret, nil
}

// Revoke a User with the Fabric CA
// registrar: The User that is initiating the revocation
// request: Revocation Request
func (im *IdentityManager) Revoke(request *contextApi.RevocationRequest) (*contextApi.RevocationResponse, error) {
	if err := im.initCAClient(); err != nil {
		return nil, err
	}
	if im.registrar.EnrollID == "" {
		return nil, contextApi.ErrCARegistrarNotFound
	}
	// Validate revocation request
	if request == nil {
		return nil, errors.New("revocation request is required")
	}
	// Create revocation request
	var req = caapi.RevocationRequest{
		CAName: request.CAName,
		Name:   request.Name,
		Serial: request.Serial,
		AKI:    request.AKI,
		Reason: request.Reason,
	}

	registrar, err := im.getRegistrarSI(im.registrar.EnrollID, im.registrar.EnrollSecret)
	if err != nil {
		return nil, err
	}

	resp, err := registrar.Revoke(&req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to revoke")
	}
	var revokedCerts []contextApi.RevokedCert
	for i := range resp.RevokedCerts {
		revokedCerts = append(
			revokedCerts,
			contextApi.RevokedCert{
				Serial: resp.RevokedCerts[i].Serial,
				AKI:    resp.RevokedCerts[i].AKI,
			})
	}

	// TODO complete the response mapping
	return &contextApi.RevocationResponse{
		RevokedCerts: revokedCerts,
		CRL:          resp.CRL,
	}, nil
}
