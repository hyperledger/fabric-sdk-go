/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
)

// IdentityManagerOption type define various initialization options
type IdentityManagerOption func(*identityManagerOptions) error

// IdentityManager implements fab/IdentityManager
type IdentityManager struct {
	orgName         string
	orgMSPID        string
	config          fab.EndpointConfig
	cryptoSuite     core.CryptoSuite
	embeddedUsers   map[string]fab.CertKeyPair
	mspPrivKeyStore core.KVStore
	mspCertStore    core.KVStore
	userStore       msp.UserStore
}

type identityManagerOptions struct {
	filesystem fs.FS
}

// WithFS allows to load certificates and keys from a virtual filesystem
func WithFS(filesystem fs.FS) IdentityManagerOption {
	return func(imo *identityManagerOptions) error {
		if filesystem == nil {
			return errors.New("filesystem is nil")
		}

		imo.filesystem = filesystem

		return nil
	}
}

// NewIdentityManager creates a new instance of IdentityManager
func NewIdentityManager(orgName string, userStore msp.UserStore, cryptoSuite core.CryptoSuite, endpointConfig fab.EndpointConfig, opts ...IdentityManagerOption) (*IdentityManager, error) {

	netConfig := endpointConfig.NetworkConfig()
	// viper keys are case insensitive
	orgConfig, ok := netConfig.Organizations[strings.ToLower(orgName)]
	if !ok {
		return nil, errors.New("org config retrieval failed")
	}

	if orgConfig.CryptoPath == "" && len(orgConfig.Users) == 0 {
		return nil, errors.New("Either a cryptopath or an embedded list of users is required")
	}

	var (
		imOpts          = new(identityManagerOptions)
		mspPrivKeyStore core.KVStore
		mspCertStore    core.KVStore
	)

	for _, opt := range opts {
		if err := opt(imOpts); err != nil {
			return nil, errors.Wrap(err, "creating a cert store failed")
		}
	}

	orgCryptoPathTemplate := orgConfig.CryptoPath
	if orgCryptoPathTemplate != "" {
		var err error
		if !filepath.IsAbs(orgCryptoPathTemplate) {
			orgCryptoPathTemplate = filepath.Join(endpointConfig.CryptoConfigPath(), orgCryptoPathTemplate)
		}

		if imOpts.filesystem == nil {
			mspPrivKeyStore, err = NewFileKeyStore(orgCryptoPathTemplate)
			if err != nil {
				return nil, errors.Wrap(err, "creating a private key store failed")
			}
			mspCertStore, err = NewFileCertStore(orgCryptoPathTemplate)
			if err != nil {
				return nil, errors.Wrap(err, "creating a cert store failed")
			}
		} else {
			mspPrivKeyStore, err = NewFileKeyStoreFS(orgCryptoPathTemplate, imOpts.filesystem)
			if err != nil {
				return nil, errors.Wrap(err, "creating a private key store fs failed")
			}
			mspCertStore, err = NewFileCertStoreFS(orgCryptoPathTemplate, imOpts.filesystem)
			if err != nil {
				return nil, errors.Wrap(err, "creating a cert store fs failed")
			}
		}

	} else {
		logger.Warnf("Cryptopath not provided for organization [%s], MSP stores not created", orgName)
	}

	mgr := &IdentityManager{
		orgName:         orgName,
		orgMSPID:        orgConfig.MSPID,
		config:          endpointConfig,
		cryptoSuite:     cryptoSuite,
		mspPrivKeyStore: mspPrivKeyStore,
		mspCertStore:    mspCertStore,
		embeddedUsers:   orgConfig.Users,
		userStore:       userStore,
		// CA Client state is created lazily, when (if) needed
	}

	return mgr, nil
}
