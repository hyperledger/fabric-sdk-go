/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"testing"

	cryptosuite "github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/identitymgr"
)

const (
	org1Name = "Org1"
	org2Name = "Org2"
)

func TestEnrollOrg2(t *testing.T) {

	cryptoSuiteProvider, err := cryptosuite.GetSuiteByConfig(testFabricConfig)
	if err != nil {
		t.Fatalf("Failed getting cryptosuite from config : %s", err)
	}

	caClient, err := identitymgr.New(org2Name, testFabricConfig, cryptoSuiteProvider)
	if err != nil {
		t.Fatalf("NewFabricCAClient return error: %v", err)
	}

	key, cert, err := caClient.Enroll("admin", "adminpw")
	if err != nil {
		t.Fatalf("Enroll returned error: %v", err)
	}
	if key == nil {
		t.Fatalf("Expected enrol to return a private key")
	}
	if cert == nil {
		t.Fatalf("Expected enrol to return an enrolment cert")
	}
}
