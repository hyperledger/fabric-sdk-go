/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicdiscovery

import (
	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	coptions "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	reqContext "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/pkg/errors"
)

// LocalService implements a dynamic Discovery Service that queries
// Fabric's Discovery service for the peers that are in the local MSP.
type LocalService struct {
	*service
	mspID string
}

// newLocalService creates a Local Discovery Service to query the list of member peers in the local MSP.
func newLocalService(config fab.EndpointConfig, mspID string, opts ...coptions.Opt) *LocalService {
	logger.Debug("Creating new local discovery service")

	s := &LocalService{mspID: mspID}
	s.service = newService(config, s.queryPeers, opts...)
	return s
}

// Initialize initializes the service with local context
func (s *LocalService) Initialize(ctx contextAPI.Local) error {
	if ctx.Identifier().MSPID != s.mspID {
		return errors.Errorf("expecting context for MSP [%s] but got [%s]", s.mspID, ctx.Identifier().MSPID)
	}
	return s.service.initialize(ctx)
}

// Close releases resources
func (s *LocalService) Close() {
	logger.Debugf("Closing local discovery service for MSP [%s]", s.mspID)
	s.service.Close()
}

func (s *LocalService) localContext() contextAPI.Local {
	return s.context().(contextAPI.Local)
}

func (s *LocalService) queryPeers() ([]fab.Peer, error) {
	peers, err := s.doQueryPeers()

	if err != nil && s.ErrHandler != nil {
		logger.Debugf("Got error from discovery query: %s. Invoking error handler", err)
		s.ErrHandler(s.ctx, "", err)
	}

	return peers, err
}

func (s *LocalService) doQueryPeers() ([]fab.Peer, error) {
	logger.Debug("Refreshing local peers from discovery service...")

	ctx := s.localContext()
	if ctx == nil {
		return nil, errors.Errorf("the service has not been initialized")
	}

	target, err := s.getTarget(ctx)
	if err != nil {
		return nil, err
	}

	reqCtx, cancel := reqContext.NewRequest(ctx, reqContext.WithTimeout(s.responseTimeout))
	defer cancel()

	req := discclient.NewRequest().AddLocalPeersQuery()
	responses, err := s.discoveryClient().Send(reqCtx, req, *target)
	if err != nil {
		return nil, errors.Wrap(err, "error calling discover service send")
	}
	if len(responses) == 0 {
		return nil, errors.Wrap(err, "expecting 1 response from discover service send but got none")
	}

	response := responses[0]
	endpoints, err := response.ForLocal().Peers()
	if err != nil {
		return nil, DiscoveryError(err)
	}

	return s.filterLocalMSP(asPeers(ctx, endpoints)), nil
}

func (s *LocalService) getTarget(ctx contextAPI.Client) (*fab.PeerConfig, error) {
	peers := ctx.EndpointConfig().NetworkPeers()
	mspID := ctx.Identifier().MSPID
	for _, p := range peers {
		// Need to go to a peer with the local MSPID, otherwise the request will be rejected
		if p.MSPID == mspID {
			return &p.PeerConfig, nil
		}
	}
	return nil, errors.Errorf("no bootstrap peers configured for MSP [%s]", mspID)
}

// Even though the local peer query should only return peers in the local
// MSP, this function double checks and logs a warning if this is not the case.
func (s *LocalService) filterLocalMSP(peers []fab.Peer) []fab.Peer {
	localMSPID := s.ctx.Identifier().MSPID
	var filteredPeers []fab.Peer
	for _, p := range peers {
		if p.MSPID() != localMSPID {
			logger.Debugf("Peer [%s] is not part of the local MSP [%s] but in MSP [%s]", p.URL(), localMSPID, p.MSPID())
		} else {
			filteredPeers = append(filteredPeers, p)
		}
	}
	return filteredPeers
}
