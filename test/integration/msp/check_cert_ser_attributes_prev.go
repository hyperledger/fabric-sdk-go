// +build prev

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
)

func checkCertAttributes(t *testing.T, certBytes []byte, expected []msp.Attribute) {
	// Do nothing, as the previous CA version wasn't setting attributes in generated certs.
}
