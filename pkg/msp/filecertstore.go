/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"fmt"
	"path"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/keyvaluestore"
	"github.com/pkg/errors"
)

// NewFileCertStore ...
func NewFileCertStore(cryptoConfogMspPath string) (core.KVStore, error) {
	_, orgName := path.Split(path.Dir(path.Dir(path.Dir(cryptoConfogMspPath))))
	opts := &keyvaluestore.FileKeyValueStoreOptions{
		Path: cryptoConfogMspPath,
		KeySerializer: func(key interface{}) (string, error) {
			ck, ok := key.(*msp.CertKey)
			if !ok {
				return "", errors.New("converting key to CertKey failed")
			}
			if ck == nil || ck.MSPID == "" || ck.Username == "" {
				return "", errors.New("invalid key")
			}
			certDir := path.Join(strings.Replace(cryptoConfogMspPath, "{username}", ck.Username, -1), "signcerts")
			return path.Join(certDir, fmt.Sprintf("%s@%s-cert.pem", ck.Username, orgName)), nil
		},
	}
	return keyvaluestore.New(opts)
}
