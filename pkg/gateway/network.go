/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
)

// A Network object represents the set of peers in a Fabric network (channel).
// Applications should get a Network instance from a Gateway using the GetNetwork method.
type Network struct {
	name    string
	gateway *Gateway
	client  *channel.Client
	event   *event.Client
}

func newNetwork(gateway *Gateway, channelProvider context.ChannelProvider) (*Network, error) {
	n := Network{
		gateway: gateway,
	}

	// Channel client is used to query and execute transactions
	client, err := channel.New(channelProvider)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create new channel client")
	}

	n.client = client

	ctx, err := channelProvider()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create new channel context")
	}

	n.name = ctx.ChannelID()

	n.event, err = event.New(channelProvider, event.WithBlockEvents())
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create new event client")
	}

	return &n, nil
}

// Name is the name of the network (also known as channel name)
func (n *Network) Name() string {
	return n.name
}

// GetContract returns instance of a smart contract on the current network.
//  Parameters:
//  name is the name of the smart contract
//
//  Returns:
//  A Contract object representing the smart contract
func (n *Network) GetContract(chaincodeID string) *Contract {
	return newContract(n, chaincodeID, "")
}

// RegisterBlockEvent registers for block events. Unregister must be called when the registration is no longer needed.
//  Returns:
//  the registration and a channel that is used to receive events. The channel is closed when Unregister is called.
func (n *Network) RegisterBlockEvent() (fab.Registration, <-chan *fab.BlockEvent, error) {
	return n.event.RegisterBlockEvent()
}

// RegisterFilteredBlockEvent registers for filtered block events. Unregister must be called when the registration is no longer needed.
//  Returns:
//  the registration and a channel that is used to receive events. The channel is closed when Unregister is called.
func (n *Network) RegisterFilteredBlockEvent() (fab.Registration, <-chan *fab.FilteredBlockEvent, error) {
	return n.event.RegisterFilteredBlockEvent()
}

// Unregister removes the given registration and closes the event channel.
//  Parameters:
//  registration is the registration handle that was returned from RegisterBlockEvent method
func (n *Network) Unregister(registration fab.Registration) {
	n.event.Unregister(registration)
}
