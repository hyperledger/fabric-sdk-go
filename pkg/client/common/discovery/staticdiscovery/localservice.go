/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package staticdiscovery

import (
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

type localDiscoveryService struct {
	config fab.EndpointConfig
	peers  []fab.Peer
	mspID  string
}

// Initialize initializes the service with local context
func (ds *localDiscoveryService) Initialize(ctx contextAPI.Local) error {
	ds.mspID = ctx.Identifier().MSPID
	return nil
}

// GetPeers is used to get local peers
func (ds *localDiscoveryService) GetPeers() ([]fab.Peer, error) {
	var peers []fab.Peer
	for _, p := range ds.peers {
		if p.MSPID() == ds.mspID {
			peers = append(peers, p)
		}
	}
	return peers, nil
}
