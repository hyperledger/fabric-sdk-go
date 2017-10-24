/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

// SelectionProvider is used to select peers for endorsement
type SelectionProvider interface {
	NewSelectionService(channelID string) (SelectionService, error)
}

// SelectionService selects peers for endorsement and commit events
type SelectionService interface {
	// GetEndorsersForChaincode returns a set of peers that should satisfy the endorsement
	// policies of all of the given chaincodes
	GetEndorsersForChaincode(channelPeers []Peer, chaincodeIDs ...string) ([]Peer, error)
}
