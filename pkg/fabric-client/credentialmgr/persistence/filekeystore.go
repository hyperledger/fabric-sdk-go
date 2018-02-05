/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package persistence

import (
	"encoding/hex"
	"path"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/api/kvstore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/keyvaluestore"
	"github.com/pkg/errors"
)

// NewFileKeyStore ...
func NewFileKeyStore(cryptoConfogMspPath string) (kvstore.KVStore, error) {
	opts := &keyvaluestore.FileKeyValueStoreOptions{
		Path: cryptoConfogMspPath,
		KeySerializer: func(key interface{}) (string, error) {
			pkk, ok := key.(*PrivKeyKey)
			if !ok {
				return "", errors.New("converting key to PrivKeyKey failed")
			}
			if pkk == nil || pkk.MspID == "" || pkk.UserName == "" || pkk.SKI == nil {
				return "", errors.New("invalid key")
			}
			keyDir := path.Join(strings.Replace(cryptoConfogMspPath, "{userName}", pkk.UserName, -1), "keystore")
			return path.Join(keyDir, hex.EncodeToString(pkk.SKI)+"_sk"), nil
		},
	}
	return keyvaluestore.NewFileKeyValueStore(opts)
}
