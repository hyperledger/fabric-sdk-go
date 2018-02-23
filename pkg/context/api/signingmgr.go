/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import "github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"

// SigningManager signs object with provided key
type SigningManager interface {
	Sign([]byte, core.Key) ([]byte, error)
}
