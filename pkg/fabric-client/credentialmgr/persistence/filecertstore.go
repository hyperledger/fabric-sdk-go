/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package persistence

import (
	"fmt"
	"path"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/api/kvstore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/keyvaluestore"
	"github.com/pkg/errors"
)

// NewFileCertStore ...
func NewFileCertStore(cryptoConfogMspPath string) (kvstore.KVStore, error) {
	_, orgName := path.Split(path.Dir(path.Dir(path.Dir(cryptoConfogMspPath))))
	opts := &keyvaluestore.FileKeyValueStoreOptions{
		Path: cryptoConfogMspPath,
		KeySerializer: func(key interface{}) (string, error) {
			ck, ok := key.(*CertKey)
			if !ok {
				return "", errors.New("converting key to CertKey failed")
			}
			if ck == nil || ck.MspID == "" || ck.UserName == "" {
				return "", errors.New("invalid key")
			}
			certDir := path.Join(strings.Replace(cryptoConfogMspPath, "{userName}", ck.UserName, -1), "signcerts")
			return path.Join(certDir, fmt.Sprintf("%s@%s-cert.pem", ck.UserName, orgName)), nil
		},
	}
	return keyvaluestore.NewFileKeyValueStore(opts)
}
