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
