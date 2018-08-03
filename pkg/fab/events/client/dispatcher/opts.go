/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/lbp"
)

type params struct {
	loadBalancePolicy                lbp.LoadBalancePolicy
	blockHeightMonitorPeriod         time.Duration
	blockHeightLagThreshold          int
	reconnectBlockHeightLagThreshold int
}

func defaultParams(config fab.EventServiceConfig) *params {
	return &params{
		loadBalancePolicy:                lbp.NewRoundRobin(),
		blockHeightMonitorPeriod:         config.BlockHeightMonitorPeriod(),
		blockHeightLagThreshold:          config.BlockHeightLagThreshold(),
		reconnectBlockHeightLagThreshold: config.ReconnectBlockHeightLagThreshold(),
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

// WithBlockHeightMonitorPeriod is the period in which the connected peer's block height is monitored.
func WithBlockHeightMonitorPeriod(value time.Duration) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(blockHeightMonitorPeriodSetter); ok {
			setter.SetBlockHeightMonitorPeriod(value)
		}
	}
}

type loadBalancePolicySetter interface {
	SetLoadBalancePolicy(value lbp.LoadBalancePolicy)
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

type blockHeightMonitorPeriodSetter interface {
	SetBlockHeightMonitorPeriod(value time.Duration)
}

func (p *params) SetBlockHeightMonitorPeriod(value time.Duration) {
	logger.Debugf("BlockHeightMonitorPeriod: %s", value)
	p.blockHeightMonitorPeriod = value
}
