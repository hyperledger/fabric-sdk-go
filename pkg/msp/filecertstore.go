/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"fmt"
	"path"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/keyvaluestore"
	"github.com/pkg/errors"
)

// NewFileCertStore ...
func NewFileCertStore(cryptoConfigMSPPath string) (core.KVStore, error) {
	_, orgName := path.Split(path.Dir(path.Dir(path.Dir(cryptoConfigMSPPath))))
	opts := &keyvaluestore.FileKeyValueStoreOptions{
		Path: cryptoConfigMSPPath,
		KeySerializer: func(key interface{}) (string, error) {
			ck, ok := key.(*msp.IdentityIdentifier)
			if !ok {
				return "", errors.New("converting key to CertKey failed")
			}
			if ck == nil || ck.MSPID == "" || ck.ID == "" {
				return "", errors.New("invalid key")
			}

			// TODO: refactor to case insensitive or remove eventually.
			r := strings.NewReplacer("{userName}", ck.ID, "{username}", ck.ID)
			certDir := path.Join(r.Replace(cryptoConfigMSPPath), "signcerts")
			return path.Join(certDir, fmt.Sprintf("%s@%s-cert.pem", ck.ID, orgName)), nil
		},
	}
	return keyvaluestore.New(opts)
}
