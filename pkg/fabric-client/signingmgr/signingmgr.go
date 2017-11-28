/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package signingmgr

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"

	"github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
)

// SigningManager is used for signing objects with private key
type SigningManager struct {
	cryptoProvider apicryptosuite.CryptoSuite
	hashOpts       apicryptosuite.HashOpts
	signerOpts     apicryptosuite.SignerOpts
}

// NewSigningManager Constructor for a signing manager.
// @param {BCCSP} cryptoProvider - crypto provider
// @param {Config} config - configuration provider
// @returns {SigningManager} new signing manager
func NewSigningManager(cryptoProvider apicryptosuite.CryptoSuite, config apiconfig.Config) (*SigningManager, error) {
	return &SigningManager{cryptoProvider: cryptoProvider, hashOpts: cryptosuite.GetSHAOpts()}, nil
}

// Sign will sign the given object using provided key
func (mgr *SigningManager) Sign(object []byte, key apicryptosuite.Key) ([]byte, error) {

	if object == nil || len(object) == 0 {
		return nil, errors.New("object (to sign) required")
	}

	if key == nil {
		return nil, errors.New("key (for signing) required")
	}

	digest, err := mgr.cryptoProvider.Hash(object, mgr.hashOpts)
	if err != nil {
		return nil, err
	}
	signature, err := mgr.cryptoProvider.Sign(key, digest, mgr.signerOpts)
	if err != nil {
		return nil, err
	}
	return signature, nil
}
