/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicdiscovery

import (
	"time"

	coptions "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

type options struct {
	refreshInterval time.Duration
	responseTimeout time.Duration
	errHandler      fab.ErrorHandler
}

// WithRefreshInterval sets the interval in which the
// peer cache is refreshed
func WithRefreshInterval(value time.Duration) coptions.Opt {
	return func(p coptions.Params) {
		logger.Debug("Checking refreshIntervalSetter")
		if setter, ok := p.(refreshIntervalSetter); ok {
			setter.SetDiscoveryRefreshInterval(value)
		}
	}
}

// WithResponseTimeout sets the Discover service response timeout
func WithResponseTimeout(value time.Duration) coptions.Opt {
	return func(p coptions.Params) {
		logger.Debug("Checking responseTimeoutSetter")
		if setter, ok := p.(responseTimeoutSetter); ok {
			setter.SetDiscoveryResponseTimeout(value)
		}
	}
}

// WithErrorHandler sets the error handler
func WithErrorHandler(value fab.ErrorHandler) coptions.Opt {
	return func(p coptions.Params) {
		logger.Debugf("Checking errHandlerSetter")
		if setter, ok := p.(errHandlerSetter); ok {
			logger.Debugf("... setting error handler")
			setter.SetErrorHandler(value)
		}
	}
}

type refreshIntervalSetter interface {
	SetDiscoveryRefreshInterval(value time.Duration)
}

type responseTimeoutSetter interface {
	SetDiscoveryResponseTimeout(value time.Duration)
}

type errHandlerSetter interface {
	SetErrorHandler(value fab.ErrorHandler)
}

func (o *options) SetDiscoveryRefreshInterval(value time.Duration) {
	logger.Debugf("RefreshInterval: %s", value)
	o.refreshInterval = value
}

func (o *options) SetDiscoveryResponseTimeout(value time.Duration) {
	logger.Debugf("ResponseTimeout: %s", value)
	o.responseTimeout = value
}

func (o *options) SetErrorHandler(value fab.ErrorHandler) {
	logger.Debugf("ErrorHandler: %+v", value)
	o.errHandler = value
}
