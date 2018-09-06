/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockheightsorter

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/balancer"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
)

const (
	defaultBlockHeightLagThreshold = 5

	// Disable disables choosing by block height threshold meaning that any peer has a likelihood of being chosen
	Disable = -1
)

type params struct {
	blockHeightLagThreshold int
	balancer                balancer.Balancer
}

func defaultParams() *params {
	return &params{
		blockHeightLagThreshold: defaultBlockHeightLagThreshold,
		balancer:                balancer.RoundRobin(),
	}
}

// WithBlockHeightLagThreshold is the number of blocks from the highest block of a group of peers
// that a peer can lag behind and still be considered to be up-to-date. These peers are sorted using
// the given Balancer. If a peer's block height falls behind this "lag" threshold then it will be
// demoted to a lower priority list of peers which is sorted according to block height.
//
// If set to 0 then only the most up-to-date peers are considered.
// If set to -1 then all peers (regardless of block height) will be load-balanced using the provided balancer.
func WithBlockHeightLagThreshold(value int) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(blockHeightLagThresholdSetter); ok {
			setter.SetBlockHeightLagThreshold(value)
		}
	}
}

// WithBalancer sets the balancing strategy to load balance (sort) the peers.
func WithBalancer(value balancer.Balancer) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(balancerSetter); ok {
			setter.SetBalancer(value)
		}
	}
}

type blockHeightLagThresholdSetter interface {
	SetBlockHeightLagThreshold(value int)
}

func (p *params) SetBlockHeightLagThreshold(value int) {
	logger.Debugf("BlockHeightLagThreshold: %d", value)
	p.blockHeightLagThreshold = value
}

type balancerSetter interface {
	SetBalancer(value balancer.Balancer)
}

func (p *params) SetBalancer(value balancer.Balancer) {
	logger.Debugf("Balancer: %#v", value)
	p.balancer = value
}
