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

	// Properties returns properties of the peer.
	Properties() Properties

	// TODO: Name, EnrollmentCertificate (if needed)
}

// PeerState provides state information about the Peer
type PeerState interface {
	BlockHeight() uint64
}

// Properties defines the properties of a peer
type Properties map[Property]interface{}

// Property is the key into the Properties map
type Property = string

// Following is a well-known list of properties of a peer, although this list may be extended.
const (
	// PropertyChaincodes defines the chaincodes that are deployed on the peer. Value type:[]*github.com/hyperledger/fabric-protos-go/gossip.Chaincode
	PropertyChaincodes Property = "Chaincodes"
	// PropertyLedgerHeight defines the ledger height property. Value type: uint64
	PropertyLedgerHeight Property = "LedgerHeight"
	// PropertyLeftChannel defines the "left-channel" property which indicates whether the peer left the channel. Value type: bool
	PropertyLeftChannel Property = "LeftChannel"
)
