/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pgresolver

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// PeerGroupResolver resolves a group of peers that would (exactly) satisfy
// a chaincode's endorsement policy.
type PeerGroupResolver interface {
	// Resolve returns a PeerGroup ensuring that all of the peers in the group are
	// in the given set of available peers.
	Resolve(peers []fab.Peer) (PeerGroup, error)
}

// LoadBalancePolicy is used to pick a peer group from a given set of peer groups
type LoadBalancePolicy interface {
	// Choose returns one of the peer groups from the given set of peer groups.
	// This method should never return nil but may return a PeerGroup that contains no peers.
	Choose(peerGroups []PeerGroup) PeerGroup
}
