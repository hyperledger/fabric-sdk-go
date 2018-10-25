/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

func TestGetCAInfo(t *testing.T) {
	mspClient, sdk := setupClient(t)
	defer integration.CleanupUserData(t, sdk)

	resp, err := mspClient.GetCAInfo()
	if err != nil {
		t.Fatalf("Get CAInfo failed: %s", err)
	}
	if resp.CAName != "ca.org1.example.com" {
		t.Fatalf("Name should be 'ca.org1.example.com'")
	}

	if resp.CAChain == nil {
		t.Fatalf("CAChain shouldn't be nil")
	}
	t.Logf("CAChain: %+v", resp.CAChain)

	if resp.IssuerPublicKey == nil {
		t.Fatalf("IssuerPublicKey shouldn't be nil")
	}
	t.Logf("IssuerPublicKey: %+v", resp.IssuerPublicKey)

	if resp.Version == "" {
		t.Fatalf("Version shouldn't be empty")
	}
	t.Logf("Version: %+v", resp.Version)
}
