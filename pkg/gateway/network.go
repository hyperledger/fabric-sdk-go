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

type network struct {
	name    string
	gateway *gateway
	client  *channel.Client
	peers   []fab.Peer
}

func newNetwork(gateway *gateway, channelProvider context.ChannelProvider) (*network, error) {
	n := network{
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

func (n *network) GetName() string {
	return n.name
}

func (n *network) GetContract(chaincodeID string) Contract {
	return newContract(n, chaincodeID, "")
}
