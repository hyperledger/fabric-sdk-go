/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialmgr

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/config/cryptoutil"

	fabricCaUtil "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/credentialmgr/persistence"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabric_sdk_go")

// CredentialManager is used for retriving user's signing identity (ecert + private key)
type CredentialManager struct {
	orgName         string
	orgMspID        string
	embeddedUsers   map[string]core.TLSKeyPair
	mspPrivKeyStore api.KVStore
	mspCertStore    api.KVStore
	config          core.Config
	cryptoProvider  core.CryptoSuite
	userStore       api.UserStore
}

// NewCredentialManager Constructor for a credential manager.
// @param {string} orgName - organisation id
// @returns {CredentialManager} new credential manager
func NewCredentialManager(orgName string, config core.Config, cryptoProvider core.CryptoSuite) (api.CredentialManager, error) {

	netConfig, err := config.NetworkConfig()
	if err != nil {
		return nil, errors.New("network config retrieval failed")
	}

	// viper keys are case insensitive
	orgConfig, ok := netConfig.Organizations[strings.ToLower(orgName)]
	if !ok {
		return nil, errors.New("org config retrieval failed")
	}

	if orgConfig.CryptoPath == "" && len(orgConfig.Users) == 0 {
		return nil, errors.New("Either a cryptopath or an embedded list of users is required")
	}

	var mspPrivKeyStore api.KVStore
	var mspCertStore api.KVStore

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
		logger.Warnf("Cryptopath not provided for organization [%s], MSP store(s) not created", orgName)
	}

	// In the future, shared UserStore from the SDK context will be used
	var userStore api.UserStore
	clientCofig, err := config.Client()
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to retrieve client config")
	}
	if clientCofig.CredentialStore.Path != "" {
		userStore, err = identity.NewCertFileUserStore(clientCofig.CredentialStore.Path, cryptoProvider)
	}

	return &CredentialManager{
		orgName:         orgName,
		orgMspID:        orgConfig.MspID,
		config:          config,
		embeddedUsers:   orgConfig.Users,
		mspPrivKeyStore: mspPrivKeyStore,
		mspCertStore:    mspCertStore,
		cryptoProvider:  cryptoProvider,
		userStore:       userStore,
	}, nil
}

// GetSigningIdentity will sign the given object with provided key,
func (mgr *CredentialManager) GetSigningIdentity(userName string) (*api.SigningIdentity, error) {
	if userName == "" {
		return nil, errors.New("username is required")
	}

	var signingIdentity *api.SigningIdentity

	if mgr.userStore != nil {
		user, err := mgr.userStore.Load(api.UserKey{MspID: mgr.orgMspID, Name: userName})
		if err == nil {
			signingIdentity = &api.SigningIdentity{MspID: user.MspID(), PrivateKey: user.PrivateKey(), EnrollmentCert: user.EnrollmentCertificate()}
		} else {
			if err != api.ErrUserNotFound {
				return nil, errors.Wrapf(err, "getting private key from cert failed")
			}
			// Not found, continue
		}
	}
	if signingIdentity == nil {
		certBytes, err := mgr.getEmbeddedCertBytes(userName)
		if err != nil && err != api.ErrUserNotFound {
			return nil, errors.WithMessage(err, "fetching embedded cert failed")
		}
		if certBytes == nil {
			certBytes, err = mgr.getCertBytesFromCertStore(userName)
			if err != nil && err != api.ErrUserNotFound {
				return nil, errors.WithMessage(err, "fetching cert from store failed")
			}
		}
		if certBytes == nil {
			return nil, api.ErrUserNotFound
		}
		privateKey, err := mgr.getEmbeddedPrivateKey(userName)
		if err != nil {
			return nil, errors.WithMessage(err, "fetching embedded private key failed")
		}
		if privateKey == nil {
			privateKey, err = mgr.getPrivateKeyFromCert(userName, certBytes)
			if err != nil {
				return nil, errors.Wrapf(err, "getting private key from cert failed")
			}
		}
		if privateKey == nil {
			return nil, fmt.Errorf("unable to find private key for user [%s]", userName)
		}
		mspID, err := mgr.config.MspID(mgr.orgName)
		if err != nil {
			return nil, errors.WithMessage(err, "MSP ID config read failed")
		}
		signingIdentity = &api.SigningIdentity{MspID: mspID, PrivateKey: privateKey, EnrollmentCert: certBytes}
	}

	return signingIdentity, nil
}

func (mgr *CredentialManager) getEmbeddedCertBytes(userName string) ([]byte, error) {
	certPem := mgr.embeddedUsers[strings.ToLower(userName)].Cert.Pem
	certPath := mgr.embeddedUsers[strings.ToLower(userName)].Cert.Path

	if certPem == "" && certPath == "" {
		return nil, api.ErrUserNotFound
	}

	var pemBytes []byte
	var err error

	if certPem != "" {
		pemBytes = []byte(certPem)
	} else if certPath != "" {
		pemBytes, err = ioutil.ReadFile(certPath)
		if err != nil {
			return nil, errors.Wrapf(err, "reading cert from embedded path failed")
		}
	}

	return pemBytes, nil
}

func (mgr *CredentialManager) getEmbeddedPrivateKey(userName string) (core.Key, error) {
	keyPem := mgr.embeddedUsers[strings.ToLower(userName)].Key.Pem
	keyPath := mgr.embeddedUsers[strings.ToLower(userName)].Key.Path

	var privateKey core.Key
	var pemBytes []byte
	var err error

	if keyPem != "" {
		// Try importing from the Embedded Pem
		pemBytes = []byte(keyPem)
	} else if keyPath != "" {
		// Try importing from the Embedded Path
		_, err := os.Stat(keyPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, errors.Wrapf(err, "OS stat embedded path failed")
			}
			// file doesn't exist, continue
		} else {
			// file exists, try to read it
			pemBytes, err = ioutil.ReadFile(keyPath)
			if err != nil {
				return nil, errors.Wrapf(err, "reading private key from embedded path failed")
			}
		}
	}

	if pemBytes != nil {
		// Try the crypto provider as a SKI
		privateKey, err = mgr.cryptoProvider.GetKey(pemBytes)
		if err != nil || privateKey == nil {
			// Try as a pem
			privateKey, err = fabricCaUtil.ImportBCCSPKeyFromPEMBytes(pemBytes, mgr.cryptoProvider, true)
			if err != nil {
				return nil, errors.Wrapf(err, "import private key failed %v", keyPem)
			}
		}
	}

	return privateKey, nil
}

func (mgr *CredentialManager) getPrivateKeyPemFromKeyStore(userName string, ski []byte) ([]byte, error) {
	if mgr.mspPrivKeyStore == nil {
		return nil, nil
	}
	key, err := mgr.mspPrivKeyStore.Load(
		&persistence.PrivKeyKey{
			MspID:    mgr.orgMspID,
			UserName: userName,
			SKI:      ski,
		})
	if err != nil {
		return nil, err
	}
	keyBytes, ok := key.([]byte)
	if !ok {
		return nil, errors.New("key from store is not []byte")
	}
	return keyBytes, nil
}

func (mgr *CredentialManager) getCertBytesFromCertStore(userName string) ([]byte, error) {
	if mgr.mspCertStore == nil {
		return nil, api.ErrUserNotFound
	}
	cert, err := mgr.mspCertStore.Load(&persistence.CertKey{
		MspID:    mgr.orgMspID,
		UserName: userName,
	})
	if err != nil {
		if err == api.ErrNotFound {
			return nil, api.ErrUserNotFound
		}
		return nil, err
	}
	certBytes, ok := cert.([]byte)
	if !ok {
		return nil, errors.New("cert from store is not []byte")
	}
	return certBytes, nil
}

func (mgr *CredentialManager) getPrivateKeyFromCert(userName string, cert []byte) (core.Key, error) {
	if cert == nil {
		return nil, errors.New("cert is nil")
	}
	pubKey, err := cryptoutil.GetPublicKeyFromCert(cert, mgr.cryptoProvider)
	if err != nil {
		return nil, errors.WithMessage(err, "fetching public key from cert failed")
	}
	privKey, err := mgr.getPrivateKeyFromKeyStore(userName, pubKey.SKI())
	if err == nil {
		return privKey, nil
	}
	if err != api.ErrNotFound {
		return nil, errors.WithMessage(err, "fetching private key from key store failed")
	}
	return mgr.cryptoProvider.GetKey(pubKey.SKI())
}

func (mgr *CredentialManager) getPrivateKeyFromKeyStore(userName string, ski []byte) (core.Key, error) {
	pemBytes, err := mgr.getPrivateKeyPemFromKeyStore(userName, ski)
	if err != nil {
		return nil, err
	}
	if pemBytes != nil {
		return fabricCaUtil.ImportBCCSPKeyFromPEMBytes(pemBytes, mgr.cryptoProvider, true)
	}
	return nil, api.ErrNotFound
}
