/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defsvc

import (
	"testing"

	discovery "github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/staticdiscovery"
	selection "github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/staticselection"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
)

func TestCreateDiscoveryProvider(t *testing.T) {

	factory := NewProviderFactory()
	config := mocks.NewMockEndpointConfig()

	dp, err := factory.CreateDiscoveryProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating discovery provider %v", err)
	}

	_, ok := dp.(*discovery.DiscoveryProvider)
	if !ok {
		t.Fatalf("Unexpected discovery provider created")
	}
}

func TestCreateSelectionProvider(t *testing.T) {
	factory := NewProviderFactory()

	config := mocks.NewMockEndpointConfig()

	dp, err := factory.CreateSelectionProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating discovery provider %v", err)
	}

	_, ok := dp.(*selection.SelectionProvider)
	if !ok {
		t.Fatalf("Unexpected selection provider created")
	}
}
