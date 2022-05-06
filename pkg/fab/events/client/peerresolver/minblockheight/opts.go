/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package minblockheight

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/lbp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/peerresolver"
)

const (
	defaultBlockHeightLagThreshold          = 5
	defaultReconnectBlockHeightLagThreshold = 10
)

type params struct {
	blockHeightLagThreshold          int
	reconnectBlockHeightLagThreshold int
	minBlockHeight                   uint64
	loadBalancePolicy                lbp.LoadBalancePolicy
}

func defaultParams(context context.Client, channelID string) *params {
	policy := context.EndpointConfig().ChannelConfig(channelID).Policies.EventService

	return &params{
		blockHeightLagThreshold:          getBlockHeightLagThreshold(policy),
		reconnectBlockHeightLagThreshold: getReconnectBlockHeightLagThreshold(policy),
		loadBalancePolicy:                peerresolver.GetBalancer(policy),
	}
}

// WithBlockHeightLagThreshold sets the block height lag threshold. If a peer is lagging behind
// the most up-to-date peer by more than the given number of blocks then it will be excluded.
// If set to 0 then only the most up-to-date peers are considered.
// If set to -1 then all peers (regardless of block height) are considered for selection.
func WithBlockHeightLagThreshold(value int) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(blockHeightLagThresholdSetter); ok {
			setter.SetBlockHeightLagThreshold(value)
		}
	}
}

// WithReconnectBlockHeightThreshold indicates that the event client is to disconnect from the peer if the peer's
// block height falls too far behind the other peers. If the connected peer lags more than the given number of blocks
// then the client will disconnect from that peer and reconnect to another peer at a more acceptable block height.
// If set to 0 then this feature is disabled.
// NOTE: Setting this value too low may cause the event client to disconnect/reconnect too frequently, thereby affecting
// performance.
func WithReconnectBlockHeightThreshold(value int) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(reconnectBlockHeightLagThresholdSetter); ok {
			setter.SetReconnectBlockHeightLagThreshold(value)
		}
	}
}

func (p *params) SetLoadBalancePolicy(value lbp.LoadBalancePolicy) {
	logger.Debugf("LoadBalancePolicy: %#v", value)
	p.loadBalancePolicy = value
}

type blockHeightLagThresholdSetter interface {
	SetBlockHeightLagThreshold(value int)
}

func (p *params) SetBlockHeightLagThreshold(value int) {
	logger.Debugf("BlockHeightLagThreshold: %d", value)
	p.blockHeightLagThreshold = value
}

type reconnectBlockHeightLagThresholdSetter interface {
	SetReconnectBlockHeightLagThreshold(value int)
}

func (p *params) SetReconnectBlockHeightLagThreshold(value int) {
	logger.Debugf("ReconnectBlockHeightLagThreshold: %d", value)
	p.reconnectBlockHeightLagThreshold = value
}

func (p *params) SetFromBlock(value uint64) {
	logger.Debugf("FromBlock: %d", value)
	p.minBlockHeight = value + 1
}

func (p *params) SetSnapshot(value fab.EventSnapshot) error {
	logger.Debugf("SetSnapshot.FromBlock: %d", value)
	p.minBlockHeight = value.LastBlockReceived() + 1
	return nil
}

func getBlockHeightLagThreshold(policy fab.EventServicePolicy) int {
	var threshold int

	switch policy.MinBlockHeightResolverMode {
	case fab.ResolveLatest:
		threshold = 0
	case fab.ResolveByThreshold:
		threshold = policy.BlockHeightLagThreshold
	default:
		logger.Warnf("Invalid MinBlockHeightResolverMode: [%s]. Using default: [%s]", policy.MinBlockHeightResolverMode, fab.ResolveByThreshold)
		threshold = policy.BlockHeightLagThreshold
		if threshold <= 0 {
			logger.Warnf("Invalid BlockHeightLagThreshold: %d. Using default: %d", threshold, defaultBlockHeightLagThreshold)
			threshold = defaultBlockHeightLagThreshold
		}
	}

	return threshold
}

func getReconnectBlockHeightLagThreshold(policy fab.EventServicePolicy) int {
	threshold := policy.ReconnectBlockHeightLagThreshold
	if threshold <= 0 {
		logger.Warnf("Invalid ReconnectBlockHeightLagThreshold: %d. Using default: %d", threshold, defaultReconnectBlockHeightLagThreshold)
		threshold = defaultReconnectBlockHeightLagThreshold
	}
	return threshold
}
