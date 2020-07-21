/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/keyvaluestore"
	"github.com/pkg/errors"
)

// NewFileKeyStore loads keys stored in the cryptoconfig directory layout.
// This function will detect if private keys are stored in v1 or v2 format.
func NewFileKeyStore(cryptoConfigMSPPath string) (core.KVStore, error) {
	opts := &keyvaluestore.FileKeyValueStoreOptions{
		Path: cryptoConfigMSPPath,
		KeySerializer: func(key interface{}) (string, error) {
			pkk, ok := key.(*msp.PrivKeyKey)
			if !ok {
				return "", errors.New("converting key to PrivKeyKey failed")
			}
			if pkk == nil || pkk.MSPID == "" || pkk.ID == "" || pkk.SKI == nil {
				return "", errors.New("invalid key")
			}

			return cryptoConfigPrivateKeyPath(cryptoConfigMSPPath, pkk.ID, pkk.SKI), nil
		},
	}
	return keyvaluestore.New(opts)
}

func cryptoConfigPrivateKeyPath(cryptoConfigMSPPath, id string, ski []byte) string {
	// TODO: refactor to case insensitive or remove eventually.
	r := strings.NewReplacer("{userName}", id, "{username}", id)
	keyDir := filepath.Join(r.Replace(cryptoConfigMSPPath), "keystore")

	keyPathPriv := filepath.Join(keyDir, "priv_sk")
	_, err := os.Stat(keyPathPriv)
	if err == nil {
		return keyPathPriv
	}

	return filepath.Join(keyDir, hex.EncodeToString(ski)+"_sk")
}
