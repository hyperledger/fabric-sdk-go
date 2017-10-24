/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package staticselection

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/config"
)

func TestStaticSelection(t *testing.T) {

	config, err := config.InitConfig("../../../../test/fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatalf(err.Error())
	}

	selectionProvider, err := NewSelectionProvider(config)
	if err != nil {
		t.Fatalf("Failed to setup selection provider: %s", err)
	}

	selectionService, err := selectionProvider.NewSelectionService("")
	if err != nil {
		t.Fatalf("Failed to setup selection service: %s", err)
	}

	peers, err := selectionService.GetEndorsersForChaincode(nil)
	if err == nil {
		t.Fatalf("Should have failed for no chaincode IDs provided")
	}

	peers, err = selectionService.GetEndorsersForChaincode(nil, "")
	if err != nil {
		t.Fatalf("Failed to get endorsers: %s", err)
	}

	expectedNumOfPeeers := 0
	if len(peers) != expectedNumOfPeeers {
		t.Fatalf("Expecting %d, got %d peers", expectedNumOfPeeers, len(peers))
	}

}
