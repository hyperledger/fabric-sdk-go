/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialmgr

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/kvstore"
	"github.com/hyperledger/fabric-sdk-go/pkg/config/cryptoutil"

	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	fabricCaUtil "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/credentialmgr/persistence"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabric_sdk_go")

// CredentialManager is used for retriving user's signing identity (ecert + private key)
type CredentialManager struct {
	orgName        string
	orgMspID       string
	embeddedUsers  map[string]apiconfig.TLSKeyPair
	privKeyStore   kvstore.KVStore
	certStore      kvstore.KVStore
	config         apiconfig.Config
	cryptoProvider apicryptosuite.CryptoSuite
}

// NewCredentialManager Constructor for a credential manager.
// @param {string} orgName - organisation id
// @returns {CredentialManager} new credential manager
func NewCredentialManager(orgName string, config apiconfig.Config, cryptoProvider apicryptosuite.CryptoSuite) (apifabclient.CredentialManager, error) {

	netConfig, err := config.NetworkConfig()
	if err != nil {
		return nil, err
	}

	// viper keys are case insensitive
	orgConfig, ok := netConfig.Organizations[strings.ToLower(orgName)]
	if !ok {
		return nil, errors.New("org config retrieval failed")
	}

	if orgConfig.CryptoPath == "" && len(orgConfig.Users) == 0 {
		return nil, errors.New("Either a cryptopath or an embedded list of users is required")
	}

	var privKeyStore kvstore.KVStore
	var certStore kvstore.KVStore

	orgCryptoPathTemplate := orgConfig.CryptoPath
	if orgCryptoPathTemplate != "" {
		if !filepath.IsAbs(orgCryptoPathTemplate) {
			orgCryptoPathTemplate = filepath.Join(config.CryptoConfigPath(), orgCryptoPathTemplate)
		}
		privKeyStore, err = persistence.NewFileKeyStore(orgCryptoPathTemplate)
		if err != nil {
			return nil, errors.Wrapf(err, "creating a private key store failed")
		}
		certStore, err = persistence.NewFileCertStore(orgCryptoPathTemplate)
		if err != nil {
			return nil, errors.Wrapf(err, "creating a cert store failed")
		}
	} else {
		logger.Warnf("Cryptopath not provided for organization [%s], store(s) not created", orgName)
	}

	return &CredentialManager{
		orgName:        orgName,
		orgMspID:       orgConfig.MspID,
		config:         config,
		embeddedUsers:  orgConfig.Users,
		privKeyStore:   privKeyStore,
		certStore:      certStore,
		cryptoProvider: cryptoProvider,
	}, nil
}

// GetSigningIdentity will sign the given object with provided key,
func (mgr *CredentialManager) GetSigningIdentity(userName string) (*apifabclient.SigningIdentity, error) {
	if userName == "" {
		return nil, errors.New("username is required")
	}

	privateKey, err := mgr.getEmbeddedPrivateKey(userName)
	if err != nil {
		return nil, errors.WithMessage(err, "fetching embedded private key failed")
	}

	mspID, err := mgr.config.MspID(mgr.orgName)
	if err != nil {
		return nil, errors.WithMessage(err, "MSP ID config read failed")
	}

	var certBytes []byte
	if privateKey == nil {
		certBytes, err = mgr.getEmbeddedCertBytes(userName)
		if err != nil {
			return nil, errors.WithMessage(err, "fetching enbedded cert failed")
		}
		if certBytes == nil {
			certBytes, err = mgr.getStoredCertBytes(userName)
			if err != nil {
				return nil, errors.WithMessage(err, "fetching cert from store failed")
			}
		}
		if certBytes == nil {
			return nil, fmt.Errorf("cert not found for user [%s]", userName)
		}
		privateKey, err = mgr.getPivateKeyFromCert(userName, certBytes)
		if err != nil {
			return nil, errors.Wrapf(err, "getting private key from cert failed")
		}
	}

	if privateKey == nil {
		return nil, fmt.Errorf("unable to find private key for user [%s]", userName)
	}

	signingIdentity := &apifabclient.SigningIdentity{MspID: mspID, PrivateKey: privateKey, EnrollmentCert: certBytes}

	return signingIdentity, nil
}

func (mgr *CredentialManager) getEmbeddedCertBytes(userName string) ([]byte, error) {
	certPem := mgr.embeddedUsers[strings.ToLower(userName)].Cert.Pem
	certPath := mgr.embeddedUsers[strings.ToLower(userName)].Cert.Path

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

func (mgr *CredentialManager) getEmbeddedPrivateKey(userName string) (apicryptosuite.Key, error) {
	keyPem := mgr.embeddedUsers[strings.ToLower(userName)].Key.Pem
	keyPath := mgr.embeddedUsers[strings.ToLower(userName)].Key.Path

	var privateKey apicryptosuite.Key
	var pemBytes []byte
	var err error

	if keyPem != "" {
		// Try importing from the Embedded Pem
		pemBytes = []byte(keyPem)
	} else if keyPath != "" {
		// Try importing from the Embedded Path
		pemBytes, err = ioutil.ReadFile(keyPath)
		if err != nil {
			return nil, errors.Wrapf(err, "reading private key from embedded path failed")
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

func (mgr *CredentialManager) getStoredPrivateKeyPem(userName string, ski []byte) ([]byte, error) {
	if mgr.privKeyStore == nil {
		return nil, nil
	}
	key, err := mgr.privKeyStore.Load(
		&persistence.PrivKeyKey{
			MspID:    mgr.orgMspID,
			UserName: userName,
			SKI:      ski,
		})
	if err != nil {
		if err == kvstore.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	keyBytes, ok := key.([]byte)
	if !ok {
		return nil, errors.New("key from store is not []byte")
	}
	return keyBytes, nil
}

func (mgr *CredentialManager) getStoredCertBytes(userName string) ([]byte, error) {
	if mgr.certStore == nil {
		return nil, nil
	}
	cert, err := mgr.certStore.Load(&persistence.CertKey{
		MspID:    mgr.orgMspID,
		UserName: userName,
	})
	if err != nil {
		if err == kvstore.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	certBytes, ok := cert.([]byte)
	if !ok {
		return nil, errors.New("cert from store is not []byte")
	}
	return certBytes, nil
}

func (mgr *CredentialManager) getPivateKeyFromCert(userName string, cert []byte) (apicryptosuite.Key, error) {
	if cert == nil {
		return nil, errors.New("cert is nil")
	}
	pubKey, err := cryptoutil.GetPublicKeyFromCert(cert, mgr.cryptoProvider)
	if err != nil {
		return nil, errors.WithMessage(err, "fetching public key from cert failed")
	}
	secProvider := mgr.config.SecurityProvider()
	if secProvider == "SW" {
		return mgr.getPivateKeyForSKIFromStore(userName, pubKey.SKI())
	}
	return mgr.getPivateKeyForSKIFromHSM(pubKey.SKI())
}

func (mgr *CredentialManager) getPivateKeyForSKIFromStore(userName string, ski []byte) (apicryptosuite.Key, error) {
	pemBytes, err := mgr.getStoredPrivateKeyPem(userName, ski)
	if err != nil {
		return nil, err
	}
	if pemBytes == nil {
		return nil, fmt.Errorf("private key not found in key store for user [%s]", userName)
	}
	privateKey, err := fabricCaUtil.ImportBCCSPKeyFromPEMBytes(pemBytes, mgr.cryptoProvider, true)
	if err != nil {
		return nil, errors.Wrapf(err, "import private key failed %v", pemBytes)
	}
	return privateKey, nil
}

func (mgr *CredentialManager) getPivateKeyForSKIFromHSM(ski []byte) (apicryptosuite.Key, error) {
	return mgr.cryptoProvider.GetKey(ski)
}
