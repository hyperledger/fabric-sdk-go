/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defsvc

import (
	"testing"

	discovery "github.com/hyperledger/fabric-sdk-go/pkg/client/discovery/staticdiscovery"
	selection "github.com/hyperledger/fabric-sdk-go/pkg/client/selection/staticselection"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
)

func TestNewDiscoveryProvider(t *testing.T) {
	factory := NewProviderFactory()

	config := mocks.NewMockConfig()

	dp, err := factory.NewDiscoveryProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating discovery provider %v", err)
	}

	_, ok := dp.(*discovery.DiscoveryProvider)
	if !ok {
		t.Fatalf("Unexpected discovery provider created")
	}
}

func TestNewSelectionProvider(t *testing.T) {
	factory := NewProviderFactory()

	config := mocks.NewMockConfig()

	dp, err := factory.NewSelectionProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating discovery provider %v", err)
	}

	_, ok := dp.(*selection.SelectionProvider)
	if !ok {
		t.Fatalf("Unexpected selection provider created")
	}
}
