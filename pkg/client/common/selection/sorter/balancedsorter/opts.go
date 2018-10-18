/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package balancedsorter

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/balancer"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
)

type params struct {
	balancer balancer.Balancer
}

func defaultParams() *params {
	return &params{
		balancer: balancer.RoundRobin(),
	}
}

// WithBalancer sets the balancing strategy to use to load balance between the peers
func WithBalancer(value balancer.Balancer) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(balancerSetter); ok {
			setter.SetBalancer(value)
		}
	}
}

type balancerSetter interface {
	SetBalancer(value balancer.Balancer)
}

func (p *params) SetBalancer(value balancer.Balancer) {
	logger.Debugf("Balancer: %#v", value)
	p.balancer = value
}
