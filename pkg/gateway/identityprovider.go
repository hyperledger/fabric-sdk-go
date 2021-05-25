/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
)

// IdentityProvider is the interface that represents an identity provider
type IdentityProvider interface {
	GetCryptoSuite() core.CryptoSuite
	FromJSON(data []byte) (Identity, error)
	ToJSON() ([]byte, error)
	GetUserContext(identity Hsmx509Identity, name string)
}