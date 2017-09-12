/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

import "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/bccsp"

// SigningManager signs object with provided key
type SigningManager interface {
	Sign([]byte, bccsp.Key) ([]byte, error)
}
