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
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/fabpvdr"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
)

func TestCreateDiscoveryProvider(t *testing.T) {
	ctx := mocks.NewMockContext(mspmocks.NewMockSigningIdentity("testuser", "testuser"))
	fabPvdr := fabpvdr.New(ctx.Config())

	factory := NewProviderFactory()
	config := mocks.NewMockConfig()

	dp, err := factory.CreateDiscoveryProvider(config, fabPvdr)
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

	config := mocks.NewMockConfig()

	dp, err := factory.CreateSelectionProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating discovery provider %v", err)
	}

	_, ok := dp.(*selection.SelectionProvider)
	if !ok {
		t.Fatalf("Unexpected selection provider created")
	}
}
