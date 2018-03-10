/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identitymgr

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	config "github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
)

var logger = logging.NewLogger("fabsdk/core")

// IdentityManager implements fab/IdentityManager
type IdentityManager struct {
	orgName         string
	orgMspID        string
	config          core.Config
	cryptoSuite     core.CryptoSuite
	embeddedUsers   map[string]core.TLSKeyPair
	mspPrivKeyStore core.KVStore
	mspCertStore    core.KVStore
	userStore       UserStore
}

// New creates a new instance of IdentityManager
// @param {string} organization
// @param {Config} client config for fabric-ca services
// @returns {IdentityManager} IdentityManager instance
// @returns {error} error, if any
func New(orgName string, stateStore core.KVStore, cryptoSuite core.CryptoSuite, config config.Config) (*IdentityManager, error) {

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

	var mspPrivKeyStore core.KVStore
	var mspCertStore core.KVStore

	orgCryptoPathTemplate := orgConfig.CryptoPath
	if orgCryptoPathTemplate != "" {
		if !filepath.IsAbs(orgCryptoPathTemplate) {
			orgCryptoPathTemplate = filepath.Join(config.CryptoConfigPath(), orgCryptoPathTemplate)
		}
		mspPrivKeyStore, err = NewFileKeyStore(orgCryptoPathTemplate)
		if err != nil {
			return nil, errors.Wrapf(err, "creating a private key store failed")
		}
		mspCertStore, err = NewFileCertStore(orgCryptoPathTemplate)
		if err != nil {
			return nil, errors.Wrapf(err, "creating a cert store failed")
		}
	} else {
		logger.Warnf("Cryptopath not provided for organization [%s], MSP stores not created", orgName)
	}

	userStore, err := NewCertFileUserStore1(stateStore)
	if err != nil {
		return nil, errors.Wrapf(err, "creating a user store failed")
	}

	mgr := &IdentityManager{
		orgName:         orgName,
		orgMspID:        orgConfig.MspID,
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
