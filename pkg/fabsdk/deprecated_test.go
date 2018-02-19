/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"testing"

	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/config"
)

const (
	txClientConfigFile = "testdata/test.yaml"
	txValidClientUser  = "User1"
	txValidClientAdmin = "Admin"
	txValidClientOrg   = "Org2"
)

func TestNewResourceMgmtWithOptsClient(t *testing.T) {
	sdk, err := New(configImpl.FromFile("../../test/fixtures/config/config_test.yaml"))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Test configuration failure for resource management client (invalid user/default organisation)
	_, err = sdk.New("Invalid")
	if err == nil {
		t.Fatalf("Should have failed to create resource management client due to invalid user")
	}

	// Test valid configuration for resource management client
	_, err = sdk.New(clientValidAdmin)
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}

	// Test configuration failure for new resource management client with options (invalid org)
	_, err = sdk.NewResourceMgmtClientWithOpts(txValidClientAdmin, &ResourceMgmtClientOpts{OrgName: "Invalid"})
	if err == nil {
		t.Fatalf("Should have failed to create resource management client due to invalid organization")
	}

	// Test new resource management client with options (Org2 configuration)
	_, err = sdk.NewResourceMgmtClientWithOpts(txValidClientAdmin, &ResourceMgmtClientOpts{OrgName: "Org2"})
	if err != nil {
		t.Fatalf("Failed to create new resource management client with opts: %s", err)
	}
}

func TestNewPreEnrolledUserSession(t *testing.T) {
	sdk, err := New(configImpl.FromFile("../../test/fixtures/config/config_test.yaml"))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	_, err = sdk.newSessionFromIdentityName("org1", txValidClientUser)
	if err != nil {
		t.Fatalf("Unexpected error loading user session: %s", err)
	}

	_, err = sdk.newSessionFromIdentityName("notarealorg", txValidClientUser)
	if err == nil {
		t.Fatal("Expected error loading user session from fake org")
	}
}
