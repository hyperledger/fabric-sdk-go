/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package filter provides common filters (e.g. Endpoint)
package filter

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
)

// EndpointType represents endpoint type
type EndpointType int32

// Endpoint types
const (
	ChaincodeQuery EndpointType = iota
	EndorsingPeer
	LedgerQuery
	EventSource
)

// NewEndpointFilter creates a new endpoint filter that is based on configuration.
// If channel peer is not configured it will be selected by default.
func NewEndpointFilter(ctx context.Channel, et EndpointType) *EndpointFilter {

	// Retrieve channel peers
	chPeers := ctx.EndpointConfig().ChannelPeers(ctx.ChannelID())
	return &EndpointFilter{endpointType: et, ctx: ctx, chPeers: chPeers}

}

// EndpointFilter filters based on endpoint config options
type EndpointFilter struct {
	endpointType EndpointType
	ctx          context.Channel
	chPeers      []fab.ChannelPeer // configured channel peers
}

// Accept returns false if this peer is to be excluded from the target list
func (f *EndpointFilter) Accept(peer fab.Peer) bool {

	peerConfig, err := comm.SearchPeerConfigFromURL(f.ctx.EndpointConfig(), peer.URL())
	if err != nil {
		return true
	}

	chPeer := f.getChannelPeer(peerConfig)
	if chPeer == nil {
		return true
	}

	switch t := f.endpointType; t {
	case ChaincodeQuery:
		return chPeer.ChaincodeQuery
	case EndorsingPeer:
		return chPeer.EndorsingPeer
	case LedgerQuery:
		return chPeer.LedgerQuery
	case EventSource:
		return chPeer.EventSource
	}

	return true
}

func (f *EndpointFilter) getChannelPeer(peerConfig *fab.PeerConfig) *fab.ChannelPeer {
	for _, chpeer := range f.chPeers {
		if chpeer.URL == peerConfig.URL {
			return &fab.ChannelPeer{
				PeerChannelConfig: chpeer.PeerChannelConfig,
				NetworkPeer:       chpeer.NetworkPeer,
			}
		}
	}
	return nil
}
