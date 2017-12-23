/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialmgr

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"

	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	fabricCaUtil "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

// CredentialManager is used for retriving user's signing identity (ecert + private key)
type CredentialManager struct {
	orgName        string
	embeddedUsers  map[string]apiconfig.TLSKeyPair
	keyDir         string
	certDir        string
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

	orgCryptoPath := orgConfig.CryptoPath
	if !filepath.IsAbs(orgCryptoPath) {
		orgCryptoPath = filepath.Join(config.CryptoConfigPath(), orgCryptoPath)
	}

	return &CredentialManager{orgName: orgName, config: config, embeddedUsers: orgConfig.Users, keyDir: orgCryptoPath + "/keystore", certDir: orgCryptoPath + "/signcerts", cryptoProvider: cryptoProvider}, nil
}

// GetSigningIdentity will sign the given object with provided key,
func (mgr *CredentialManager) GetSigningIdentity(userName string) (*apifabclient.SigningIdentity, error) {
	if userName == "" {
		return nil, errors.New("username is required")
	}

	mspID, err := mgr.config.MspID(mgr.orgName)
	if err != nil {
		return nil, errors.WithMessage(err, "MSP ID config read failed")
	}

	privateKey, err := mgr.getPrivateKey(userName)

	if err != nil {
		return nil, err
	}

	enrollmentCert, err := mgr.getEnrollmentCert(userName)

	if err != nil {
		return nil, err
	}

	signingIdentity := &apifabclient.SigningIdentity{MspID: mspID, PrivateKey: privateKey, EnrollmentCert: enrollmentCert}

	return signingIdentity, nil
}

func (mgr *CredentialManager) getPrivateKey(userName string) (apicryptosuite.Key, error) {
	keyPem := mgr.embeddedUsers[strings.ToLower(userName)].Key.Pem
	keyPath := mgr.embeddedUsers[strings.ToLower(userName)].Key.Path

	var privateKey apicryptosuite.Key
	var err error

	if keyPem != "" {
		// First try importing from the Embedded Pem
		privateKey, err = fabricCaUtil.ImportBCCSPKeyFromPEMBytes([]byte(keyPem), mgr.cryptoProvider, true)

		if err != nil {
			return nil, errors.Wrapf(err, "import private key failed %v", keyPem)
		}
	} else if keyPath != "" {
		// Then try importing from the Embedded Path
		privateKey, err = fabricCaUtil.ImportBCCSPKeyFromPEM(keyPath, mgr.cryptoProvider, true)

		if err != nil {
			return nil, errors.Wrap(err, "import private key failed")
		}
	} else if mgr.keyDir != "" {
		// Then try importing from the Crypto Path

		privateKeyDir := strings.Replace(mgr.keyDir, "{userName}", userName, -1)

		privateKeyPath, err := getFirstPathFromDir(privateKeyDir)

		if err != nil {
			return nil, errors.WithMessage(err, "find private key path failed")
		}

		privateKey, err = fabricCaUtil.ImportBCCSPKeyFromPEM(privateKeyPath, mgr.cryptoProvider, true)

		if err != nil {
			return nil, errors.Wrap(err, "import private key failed")
		}
	} else {
		return nil, errors.Errorf("failed to find a private key for user %s", userName)
	}

	return privateKey, nil
}

func (mgr *CredentialManager) getEnrollmentCert(userName string) ([]byte, error) {
	var err error

	certPem := mgr.embeddedUsers[strings.ToLower(userName)].Cert.Pem
	certPath := mgr.embeddedUsers[strings.ToLower(userName)].Cert.Path

	var enrollmentCertBytes []byte

	if certPem != "" {
		enrollmentCertBytes = []byte(certPem)
	} else if certPath != "" {
		enrollmentCertBytes, err = ioutil.ReadFile(certPath)

		if err != nil {
			return nil, errors.Wrap(err, "reading enrollment cert path failed")
		}
	} else if mgr.certDir != "" {
		enrollmentCertDir := strings.Replace(mgr.certDir, "{userName}", userName, -1)
		enrollmentCertPath, err := getFirstPathFromDir(enrollmentCertDir)

		if err != nil {
			return nil, errors.WithMessage(err, "find enrollment cert path failed")
		}

		enrollmentCertBytes, err = ioutil.ReadFile(enrollmentCertPath)

		if err != nil {
			return nil, errors.WithMessage(err, "reading enrollment cert path failed")
		}
	} else {
		return nil, errors.Errorf("failed to find enrollment cert for user %s", userName)
	}

	return enrollmentCertBytes, nil
}

// Gets the first path from the dir directory
func getFirstPathFromDir(dir string) (string, error) {

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", errors.Wrap(err, "read directory failed")
	}

	for _, p := range files {
		if p.IsDir() {
			continue
		}

		fullName := filepath.Join(dir, string(filepath.Separator), p.Name())
		logger.Debugf("Reading file %s\n", fullName)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		fullName := filepath.Join(dir, string(filepath.Separator), f.Name())
		return fullName, nil
	}

	return "", errors.New("no paths found")
}
