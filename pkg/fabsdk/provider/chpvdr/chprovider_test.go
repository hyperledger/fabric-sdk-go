/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chpvdr

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
)

func TestBasicValidChannel(t *testing.T) {
	pf := defcore.NewProviderFactory()
	ctx := mocks.NewMockProviderContext()
	user := mocks.NewMockUser("user")

	fp, err := pf.NewFabricProvider(ctx)
	if err != nil {
		t.Fatalf("Unexpected error creating Fabric Provider: %v", err)
	}

	cp, err := New(fp)
	if err != nil {
		t.Fatalf("Unexpected error creating Channel Provider: %v", err)
	}

	channelService, err := cp.NewChannelService(user, "mychannel")
	if err != nil {
		t.Fatalf("Unexpected error creating Channel Service: %v", err)
	}

	_, err = channelService.Channel()
	if err != nil {
		t.Fatalf("Unexpected error creating Channel Service: %v", err)
	}
}
