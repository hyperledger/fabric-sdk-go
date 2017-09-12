/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package signingmgr

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/bccsp"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
)

// SigningManager is used for signing objects with private key
type SigningManager struct {
	cryptoProvider bccsp.BCCSP
	hashOpts       bccsp.HashOpts
	signerOpts     bccsp.SignerOpts
}

// NewSigningManager Constructor for a signing manager.
// @param {BCCSP} cryptoProvider - crypto provider
// @param {Config} config - configuration provider
// @returns {SigningManager} new signing manager
func NewSigningManager(cryptoProvider bccsp.BCCSP, config apiconfig.Config) (*SigningManager, error) {
	return &SigningManager{cryptoProvider: cryptoProvider, hashOpts: &bccsp.SHAOpts{}}, nil
}

// Sign will sign the given object using provided key
func (mgr *SigningManager) Sign(object []byte, key bccsp.Key) ([]byte, error) {

	if object == nil || len(object) == 0 {
		return nil, fmt.Errorf("Must provide object to sign")
	}

	if key == nil {
		return nil, fmt.Errorf("Must provide key for signing")
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
