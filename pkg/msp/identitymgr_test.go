/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
)

// TODO Add tests

// TestInterfaces will test if the interface instantiation happens properly, ie no nil returned
func TestManagerInterfaces(t *testing.T) {
	var apiIM msp.IdentityManager
	var im IdentityManager

	apiIM = &im
	if apiIM == nil {
		t.Fatal("this shouldn't happen.")
	}
}
