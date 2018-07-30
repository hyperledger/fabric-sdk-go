/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

// The Peer class represents a peer in the target blockchain network to which
// HFC sends endorsement proposals or query requests.
type Peer interface {
	ProposalProcessor
	// MSPID gets the Peer mspID.
	MSPID() string

	//URL gets the peer address
	URL() string

	// TODO: Roles, Name, EnrollmentCertificate (if needed)
}

// PeerState provides state information about the Peer
type PeerState interface {
	BlockHeight() uint64
}
