/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package persistence

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

// PrivKeyKey is a composite key for accessing a private key in the key store
type PrivKeyKey struct {
	MspID    string
	UserName string
	SKI      []byte
}

// CertKey is a composite key for accessing a cert in the cert store
type CertKey struct {
	MspID    string
	UserName string
}
