/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chconfig

import (
	"time"

	coptions "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
)

const (
	defaultRefreshInterval = time.Second * 90
)

type params struct {
	refreshInterval time.Duration
}

func newDefaultParams() *params {
	return &params{
		refreshInterval: defaultRefreshInterval,
	}
}

// WithRefreshInterval sets the interval in which the
// channel config cache is refreshed
func WithRefreshInterval(value time.Duration) coptions.Opt {
	return func(p coptions.Params) {
		if setter, ok := p.(refreshIntervalSetter); ok {
			setter.SetChConfigRefreshInterval(value)
		}
	}
}

type refreshIntervalSetter interface {
	SetChConfigRefreshInterval(value time.Duration)
}

func (o *params) SetChConfigRefreshInterval(value time.Duration) {
	logger.Debugf("RefreshInterval: %s", value)
	o.refreshInterval = value
}
