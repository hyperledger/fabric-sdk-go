/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/msp"
	"github.com/pkg/errors"
)

var (
	// ErrUserNotFound indicates the user was not found
	ErrUserNotFound = errors.New("user not found")
)

// IdentityManager provides management of identities in a Fabric network
type IdentityManager interface {
	GetSigningIdentity(name string) (msp.SigningIdentity, error)
}
