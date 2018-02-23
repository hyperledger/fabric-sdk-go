/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package staticselection

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabric_sdk_go")

// SelectionProvider implements selection provider
type SelectionProvider struct {
	config core.Config
}

// New returns static selection provider
func New(config core.Config) (*SelectionProvider, error) {
	return &SelectionProvider{config: config}, nil
}

// selectionService implements static selection service
type selectionService struct {
}

// NewSelectionService creates a static selection service
func (p *SelectionProvider) NewSelectionService(channelID string) (fab.SelectionService, error) {
	return &selectionService{}, nil
}

func (s *selectionService) GetEndorsersForChaincode(channelPeers []fab.Peer,
	chaincodeIDs ...string) ([]fab.Peer, error) {

	if len(chaincodeIDs) == 0 {
		return nil, errors.New("no chaincode IDs provided")
	}

	return channelPeers, nil
}
