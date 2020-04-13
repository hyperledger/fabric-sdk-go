/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicdiscovery

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/multi"
	coptions "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	reqContext "github.com/hyperledger/fabric-sdk-go/pkg/context"
	fabdiscovery "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery"
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
	if ctx, ok := s.context().(contextAPI.Local); ok {
		return ctx
	}
	return nil
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

	targets, err := s.getTargets(ctx)
	if err != nil {
		return nil, err
	}

	reqCtx, cancel := reqContext.NewRequest(ctx, reqContext.WithTimeout(s.responseTimeout))
	defer cancel()

	req := fabdiscovery.NewRequest().AddLocalPeersQuery()
	responsesCh, err := s.discoveryClient().Send(reqCtx, req, targets...)

	if err != nil {
		return nil, errors.Wrap(err, "error calling discover service send")
	}

	var respErrors []error

	for resp := range responsesCh {
		endpoints, err := resp.ForLocal().Peers()

		if err == nil {
			//got successful response, cancel all outstanding requests to other targets
			cancel()

			return s.filterLocalMSP(asPeers(ctx, endpoints)), nil
		}

		respErrors = append(respErrors, newDiscoveryError(err, resp.Target()))
	}

	return nil, errors.Wrap(multi.New(respErrors...), "no successful response received from any peer")
}

func (s *LocalService) getTargets(ctx contextAPI.Client) ([]fab.PeerConfig, error) {
	peers := ctx.EndpointConfig().NetworkPeers()
	mspID := ctx.Identifier().MSPID
	var targets []fab.PeerConfig
	for _, p := range peers {
		// Need to go to a peer with the local MSPID, otherwise the request will be rejected
		if p.MSPID == mspID {
			targets = append(targets, p.PeerConfig)
		}
	}

	if len(targets) == 0 {
		return nil, errors.Errorf("no bootstrap peers configured for MSP [%s]", mspID)
	}
	return targets, nil
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
