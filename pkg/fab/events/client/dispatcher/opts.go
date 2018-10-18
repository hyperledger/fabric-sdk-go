/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/lbp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/peerresolver"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/peerresolver/balanced"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/peerresolver/minblockheight"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/peerresolver/preferorg"
)

const (
	defaultPeerMonitorPeriod = 5 * time.Second
)

type params struct {
	peerMonitorPeriod    time.Duration
	peerResolverProvider peerresolver.Provider
}

func defaultParams(context context.Client, channelID string) *params {
	policy := context.EndpointConfig().ChannelConfig(channelID).Policies.EventService

	peerMonitorPeriod := policy.PeerMonitorPeriod

	// Set the peer monitor period to 0 (disabled) if explicitly configured to be disabled or
	// if the resolver is Balanced (since there's no need for a peer monitor for Balanced strategy)
	if policy.PeerMonitor == fab.Disabled || policy.ResolverStrategy == fab.BalancedStrategy {
		peerMonitorPeriod = 0
	} else if peerMonitorPeriod <= 0 {
		logger.Warnf("Invalid PeerMonitorPeriod: %s. Using default: %s.", peerMonitorPeriod, defaultPeerMonitorPeriod)
		peerMonitorPeriod = defaultPeerMonitorPeriod
	}

	return &params{
		peerMonitorPeriod:    peerMonitorPeriod,
		peerResolverProvider: getPeerResolver(policy),
	}
}

// WithLoadBalancePolicy sets the load-balance policy to use when
// choosing an event endpoint from a set of endpoints
func WithLoadBalancePolicy(value lbp.LoadBalancePolicy) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(loadBalancePolicySetter); ok {
			setter.SetLoadBalancePolicy(value)
		}
	}
}

// WithPeerMonitorPeriod is the period with which the connected peer is monitored
// to see whether or not it should be disconnected.
func WithPeerMonitorPeriod(value time.Duration) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(peerMonitorPeriodSetter); ok {
			setter.SetPeerMonitorPeriod(value)
		}
	}
}

// WithPeerResolver sets the peer resolver that chooses the peer from a discovered list of peers.
func WithPeerResolver(value peerresolver.Provider) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(peerResolverSetter); ok {
			setter.SetPeerResolver(value)
		}
	}
}

type loadBalancePolicySetter interface {
	SetLoadBalancePolicy(value lbp.LoadBalancePolicy)
}

type peerMonitorPeriodSetter interface {
	SetPeerMonitorPeriod(value time.Duration)
}

func (p *params) SetPeerMonitorPeriod(value time.Duration) {
	logger.Debugf("PeerMonitorPeriod: %s", value)
	p.peerMonitorPeriod = value
}

type peerResolverSetter interface {
	SetPeerResolver(value peerresolver.Provider)
}

func (p *params) SetPeerResolver(value peerresolver.Provider) {
	logger.Debugf("PeerResolver: %#v", value)
	p.peerResolverProvider = value
}

func getPeerResolver(policy fab.EventServicePolicy) peerresolver.Provider {
	switch policy.ResolverStrategy {
	case fab.PreferOrgStrategy:
		logger.Debugf("Using prefer-org peer resolver")
		return preferorg.NewResolver()
	case fab.MinBlockHeightStrategy:
		logger.Debugf("Using min-block-height peer resolver")
		return minblockheight.NewResolver()
	case fab.BalancedStrategy:
		logger.Debugf("Using balanced peer resolver")
		return balanced.NewResolver()
	default:
		logger.Debugf("Resolver strategy not specified. Using prefer-org peer resolver.")
		return preferorg.NewResolver()
	}
}
