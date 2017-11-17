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

	if orgConfig.CryptoPath == "" {
		return nil, errors.New("CryptoPath is required")
	}

	orgCryptoPath := orgConfig.CryptoPath
	if !filepath.IsAbs(orgCryptoPath) {
		orgCryptoPath = filepath.Join(config.CryptoConfigPath(), orgCryptoPath)
	}

	return &CredentialManager{orgName: orgName, config: config, keyDir: orgCryptoPath + "/keystore", certDir: orgCryptoPath + "/signcerts", cryptoProvider: cryptoProvider}, nil
}

// GetSigningIdentity will sign the given object with provided key,
func (mgr *CredentialManager) GetSigningIdentity(userName string) (*apifabclient.SigningIdentity, error) {

	if userName == "" {
		return nil, errors.New("username is required")
	}

	privateKeyDir := strings.Replace(mgr.keyDir, "{userName}", userName, -1)
	enrollmentCertDir := strings.Replace(mgr.certDir, "{userName}", userName, -1)

	privateKeyPath, err := getFirstPathFromDir(privateKeyDir)
	if err != nil {
		return nil, errors.WithMessage(err, "find private key path failed")
	}

	enrollmentCertPath, err := getFirstPathFromDir(enrollmentCertDir)
	if err != nil {
		return nil, errors.WithMessage(err, "find enrollment cert path failed")
	}

	mspID, err := mgr.config.MspID(mgr.orgName)
	if err != nil {
		return nil, errors.WithMessage(err, "MSP ID config read failed")
	}

	privateKey, err := fabricCaUtil.ImportBCCSPKeyFromPEM(privateKeyPath, mgr.cryptoProvider, true)
	if err != nil {
		return nil, errors.Wrap(err, "import private key failed")
	}
	enrollmentCert, err := ioutil.ReadFile(enrollmentCertPath)
	if err != nil {
		return nil, errors.Wrap(err, "reading enrollment cert path failed")
	}

	signingIdentity := &apifabclient.SigningIdentity{MspID: mspID, PrivateKey: privateKey, EnrollmentCert: enrollmentCert}

	return signingIdentity, nil

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
