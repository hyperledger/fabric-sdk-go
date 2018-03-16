/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package staticselection

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	copts "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

const loggerModule = "fabsdk/client"

var logger = logging.NewLogger(loggerModule)

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
	discoveryService fab.DiscoveryService
}

// CreateSelectionService creates a static selection service
func (p *SelectionProvider) CreateSelectionService(channelID string) (fab.SelectionService, error) {
	return &selectionService{}, nil
}

func (s *selectionService) Initialize(context contextAPI.Channel) error {
	s.discoveryService = context.DiscoveryService()
	return nil
}

func (s *selectionService) GetEndorsersForChaincode(chaincodeIDs []string, opts ...copts.Opt) ([]fab.Peer, error) {
	params := options.NewParams(opts)

	channelPeers, err := s.discoveryService.GetPeers()
	if err != nil {
		logger.Errorf("Error retrieving peers from discovery service: %s", err)
		return nil, nil
	}

	// Apply peer filter if provided
	if params.PeerFilter != nil {
		var peers []fab.Peer
		for _, peer := range channelPeers {
			if params.PeerFilter(peer) {
				peers = append(peers, peer)
			}
		}
		channelPeers = peers
	}

	if logging.IsEnabledFor(loggerModule, logging.DEBUG) {
		str := ""
		for i, peer := range channelPeers {
			str += peer.URL()
			if i+1 < len(channelPeers) {
				str += ","
			}
		}
		logger.Debugf("Available peers:\n%s\n", str)
	}

	return channelPeers, nil
}
