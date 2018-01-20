/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabapi

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

func TestNewSDK(t *testing.T) {
	setup := Options{
		ConfigFile: "../../test/fixtures/config/invalid.yaml",
	}

	// Test new SDK with invalid config file
	_, err := NewSDK(setup)
	if err == nil {
		t.Fatalf("Should have failed for invalid config file")
	}

	// Test New SDK with valid config file
	setup.ConfigFile = "../../test/fixtures/config/config_test.yaml"
	sdk, err := NewSDK(setup)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Default channel client (uses organisation from client configuration)
	_, err = sdk.NewClient(fabsdk.WithUser("User1")).Channel("mychannel")
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	// keep a common client context for the next two tests
	c2 := sdk.NewClient(fabsdk.WithUser("User1"), fabsdk.WithOrg("Org2"))

	// Test configuration failure for channel client (mychannel does't have event source configured for Org2)
	_, err = c2.Channel("mychannel")
	if err == nil {
		t.Fatalf("Should have failed to create channel client since event source not configured for Org2")
	}

	// Test new channel client with options
	_, err = c2.Channel("orgchannel")
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}
}
