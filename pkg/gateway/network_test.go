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

	gw := &Gateway{}

	nw, err := newNetwork(gw, c)

	if err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	if nw.Name() != "mychannel" {
		t.Fatalf("Incorrect network name: %s", nw.Name())
	}
}

func TestGetContract(t *testing.T) {
	c := mockChannelProvider("mychannel")

	gw := &Gateway{}

	nw, err := newNetwork(gw, c)

	if err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	contr := nw.GetContract("contract1")
	name := contr.Name()

	if name != "contract1" {
		t.Fatalf("Incorrect contract name: %s", err)
	}
}

func TestBlockEvent(t *testing.T) {

	gw := &Gateway{
		options: &gatewayOptions{
			Timeout:   defaultTimeout,
		},
	}

	c := mockChannelProvider("mychannel")

	nw, err := newNetwork(gw, c)

	if err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	reg, _, err := nw.RegisterBlockEvent()
	if err != nil {
		t.Fatalf("Failed to register block event: %s", err)
	}

	nw.Unregister(reg)
}

func TestFilteredBlocktEvent(t *testing.T) {

	gw := &Gateway{
		options: &gatewayOptions{
			Timeout:   defaultTimeout,
		},
	}

	c := mockChannelProvider("mychannel")

	nw, err := newNetwork(gw, c)

	if err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	reg, _, err := nw.RegisterFilteredBlockEvent()
	if err != nil {
		t.Fatalf("Failed to register filtered block event: %s", err)
	}

	nw.Unregister(reg)
}

func mockChannelProvider(channelID string) context.ChannelProvider {

	channelProvider := func() (context.Channel, error) {
		return mocks.NewMockChannel(channelID)
	}

	return channelProvider
}
