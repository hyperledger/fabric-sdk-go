/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmt

import fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"

// TargetFilter allows for filtering target peers
type TargetFilter interface {
	// Accept returns true if peer should be included in the list of target peers
	Accept(peer fab.Peer) bool
}

// JoinChannelOpts contains options for peers joining channel
type JoinChannelOpts struct {
	Targets      []fab.Peer   // target peers
	TargetFilter TargetFilter // peer filter
}

// ResourceMgmtClient is responsible for managing resources: peers joining channels, and installing and instantiating chaincodes(TODO).
type ResourceMgmtClient interface {

	// JoinChannel allows for peers to join existing channel
	JoinChannel(channelID string) error

	//JoinChannelWithOpts allows for customizing set of peers about to join the channel (specific peers/filtered peers)
	JoinChannelWithOpts(channelID string, opts JoinChannelOpts) error
}
