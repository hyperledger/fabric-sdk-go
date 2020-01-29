/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
)

func TestNewNetwork(t *testing.T) {
	c := mockChannelProvider("mychannel")

	gw := &gateway{}

	nw, err := newNetwork(gw, c)

	if err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	if nw.GetName() != "mychannel" {
		t.Fatalf("Incorrect network name: %s", nw.GetName())
	}
}

func TestGetContract(t *testing.T) {
	c := mockChannelProvider("mychannel")

	gw := &gateway{}

	nw, err := newNetwork(gw, c)

	if err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	contr := nw.GetContract("contract1")
	name := contr.GetName()

	if name != "contract1" {
		t.Fatalf("Incorrect contract name: %s", err)
	}
}

func mockChannelProvider(channelID string) context.ChannelProvider {

	channelProvider := func() (context.Channel, error) {
		return mocks.NewMockChannel(channelID)
	}

	return channelProvider
}
