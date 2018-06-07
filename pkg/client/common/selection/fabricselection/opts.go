/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricselection

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"

	coptions "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
)

type params struct {
	refreshInterval time.Duration
	responseTimeout time.Duration
	retryOpts       retry.Opts
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

// WithRetryOpts sets retry options for retries on transient errors
// from the Discovery Server
func WithRetryOpts(value retry.Opts) coptions.Opt {
	return func(p coptions.Params) {
		logger.Debug("Checking retryOptsSetter")
		if setter, ok := p.(retryOptsSetter); ok {
			setter.SetRetryOpts(value)
		}
	}
}

type refreshIntervalSetter interface {
	SetRefreshInterval(value time.Duration)
}

type responseTimeoutSetter interface {
	SetResponseTimeout(value time.Duration)
}

type retryOptsSetter interface {
	SetRetryOpts(value retry.Opts)
}

func (o *params) SetRefreshInterval(value time.Duration) {
	logger.Debugf("RefreshInterval: %s", value)
	o.refreshInterval = value
}

func (o *params) SetResponseTimeout(value time.Duration) {
	logger.Debugf("ResponseTimeout: %s", value)
	o.responseTimeout = value
}

func (o *params) SetRetryOpts(value retry.Opts) {
	logger.Debugf("RetryOpts: %#v", value)
	o.retryOpts = value
}
