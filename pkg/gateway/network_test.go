/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"testing"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
)

func TestNewNetwork(t *testing.T) {
	c := mockChannelProvider("mychannel")

	gw := &Gateway{
		options: &gatewayOptions{},
	}

	nw, err := newNetwork(gw, c)

	if err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	if nw.Name() != "mychannel" {
		t.Fatalf("Incorrect network name: %s", nw.Name())
	}
}

func TestNewNetworkWithEventOptions(t *testing.T) {
	c := mockChannelProvider("mychannel")

	gw := &Gateway{
		options: &gatewayOptions{
			FromBlock:    2,
			FromBlockSet: true,
		},
	}

	_, err := newNetwork(gw, c)

	if err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}
}

func TestGetContract(t *testing.T) {
	c := mockChannelProvider("mychannel")

	gw := &Gateway{
		options: &gatewayOptions{},
	}

	nw, err := newNetwork(gw, c)

	if err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	contr := nw.GetContract("contract1")
	name := contr.Name()

	if name != "contract1" {
		t.Fatalf("Incorrect contract name: %s", name)
	}
}

func TestGetContractWithName(t *testing.T) {
	c := mockChannelProvider("mychannel")

	gw := &Gateway{
		options: &gatewayOptions{},
	}

	nw, err := newNetwork(gw, c)

	if err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	contr := nw.GetContractWithName("contract1", "class1")
	name := contr.Name()

	if name != "contract1:class1" {
		t.Fatalf("Incorrect contract name: %s", name)
	}
}

func TestBlockEvent(t *testing.T) {

	gw := &Gateway{
		options: &gatewayOptions{
			Timeout: defaultTimeout,
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
			Timeout: defaultTimeout,
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

func TestNewNetworkFailure1(t *testing.T) {
	c := mockBadChannelProvider("mychannel", 2)

	gw := &Gateway{
		options: &gatewayOptions{},
	}

	_, err := newNetwork(gw, c)

	if err == nil {
		t.Fatal("Should have failed to create network")
	}
}

func TestNewNetworkFailure2(t *testing.T) {
	c := mockBadChannelProvider("mychannel", 3)

	gw := &Gateway{
		options: &gatewayOptions{},
	}

	_, err := newNetwork(gw, c)

	if err == nil {
		t.Fatal("Should have failed to create network")
	}
}

func mockChannelProvider(channelID string) context.ChannelProvider {

	channelProvider := func() (context.Channel, error) {
		return mocks.NewMockChannel(channelID)
	}

	return channelProvider
}

// mock channel provider that fails on the nth invocation
func mockBadChannelProvider(channelID string, invocations int) context.ChannelProvider {
	count := 0
	channelProvider := func() (context.Channel, error) {
		count++
		if count == invocations {
			return nil, errors.New("Mock failure")
		}
		return mocks.NewMockChannel(channelID)
	}

	return channelProvider
}
