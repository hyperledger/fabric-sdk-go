/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicdiscovery

import (
	"time"

	coptions "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
)

type options struct {
	refreshInterval time.Duration
	responseTimeout time.Duration
}

// WithRefreshInterval sets the interval in which the
// peer cache is refreshed
func WithRefreshInterval(value time.Duration) coptions.Opt {
	return func(p coptions.Params) {
		logger.Debug("Checking refreshIntervalSetter")
		if setter, ok := p.(refreshIntervalSetter); ok {
			setter.SetRefreshInterval(value)
		}
	}
}

// WithResponseTimeout sets the Discover service response timeout
func WithResponseTimeout(value time.Duration) coptions.Opt {
	return func(p coptions.Params) {
		logger.Debug("Checking responseTimeoutSetter")
		if setter, ok := p.(responseTimeoutSetter); ok {
			setter.SetResponseTimeout(value)
		}
	}
}

type refreshIntervalSetter interface {
	SetRefreshInterval(value time.Duration)
}

type responseTimeoutSetter interface {
	SetResponseTimeout(value time.Duration)
}

func (o *options) SetRefreshInterval(value time.Duration) {
	logger.Debugf("RefreshInterval: %s", value)
	o.refreshInterval = value
}

func (o *options) SetResponseTimeout(value time.Duration) {
	logger.Debugf("ResponseTimeout: %s", value)
	o.responseTimeout = value
}
