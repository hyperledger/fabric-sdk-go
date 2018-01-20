/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicselection

import (
	"strings"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

func TestCCPolicyProvider(t *testing.T) {
	// Create SDK setup for channel client with dynamic selection
	sdk, err := fabsdk.New(config.FromFile("../../../../test/fixtures/config/config_test.yaml"))
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	// Nil sdk
	ccPolicyProvider, err := newCCPolicyProvider(nil, "mychannel", "User1", "Org1")
	if err == nil {
		t.Fatalf("Should have failed for nil sdk")
	}

	// Invalid channelID
	ccPolicyProvider, err = newCCPolicyProvider(sdk, "", "User1", "Org1")
	if err == nil {
		t.Fatalf("Should have failed for empty channel")
	}

	// Empty user name
	ccPolicyProvider, err = newCCPolicyProvider(sdk, "mychannel", "", "Prg1")
	if err == nil {
		t.Fatalf("Should have failed for empty user name")
	}

	// Empty org name
	ccPolicyProvider, err = newCCPolicyProvider(sdk, "mychannel", "User1", "")
	if err == nil {
		t.Fatalf("Should have failed for nil sdk")
	}

	// Invalid channel
	ccPolicyProvider, err = newCCPolicyProvider(sdk, "non-existent", "User1", "Org1")
	if err == nil {
		t.Fatalf("Should have failed for invalid channel name")
	}

	// All good
	ccPolicyProvider, err = newCCPolicyProvider(sdk, "mychannel", "User1", "Org1")
	if err != nil {
		t.Fatalf("Failed to setup cc policy provider: %s", err)
	}

	// Empty chaincode ID
	_, err = ccPolicyProvider.GetChaincodePolicy("")
	if err == nil {
		t.Fatalf("Should have failed to retrieve chaincode policy for empty chaincode id")
	}

	// Non-existent chaincode ID
	_, err = ccPolicyProvider.GetChaincodePolicy("abc")
	if err == nil {
		t.Fatalf("Should have failed to retrieve non-existent cc policy")
	}
}

func TestBadClient(t *testing.T) {
	// Create SDK setup for channel client with dynamic selection
	sdk, err := fabsdk.New(config.FromFile("../../../../test/fixtures/config/config_test.yaml"))
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	// Non-existent user
	ccPolicyProvider, err := newCCPolicyProvider(sdk, "mychannel", "Invalid", "Org1")
	_, err = ccPolicyProvider.GetChaincodePolicy("mychannel")
	if !strings.Contains(err.Error(), "Unable to load identity") {
		t.Fatalf("Should have failed for invalid user name: %v", err)
	}

	// Invalid org
	ccPolicyProvider, err = newCCPolicyProvider(sdk, "mychannel", "User1", "Invalid")
	_, err = ccPolicyProvider.GetChaincodePolicy("mychannel")
	if !strings.Contains(err.Error(), "Unable to load identity") {
		t.Fatalf("Should have failed for invalid org name")
	}
}
