/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
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
	peers   []fab.Peer
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

	discovery, err := ctx.ChannelService().Discovery()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create discovery service")
	}

	peers, err := discovery.GetPeers()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to discover peers")
	}

	n.peers = peers

	return &n, nil
}

// Name is the name of the network (also known as channel name)
func (n *Network) Name() string {
	return n.name
}

// GetContract returns instance of a smart contract on the current network.
func (n *Network) GetContract(chaincodeID string) *Contract {
	return newContract(n, chaincodeID, "")
}
