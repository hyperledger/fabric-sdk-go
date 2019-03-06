/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricselection

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	coptions "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

type params struct {
	refreshInterval time.Duration
	responseTimeout time.Duration
	retryOpts       retry.Opts
	errHandler      fab.ErrorHandler
}

// WithRefreshInterval sets the interval in which the
// peer cache is refreshed
func WithRefreshInterval(value time.Duration) coptions.Opt {
	return func(p coptions.Params) {
		logger.Debug("Checking refreshIntervalSetter")
		if setter, ok := p.(refreshIntervalSetter); ok {
			setter.SetSelectionRefreshInterval(value)
		}
	}
}

// WithResponseTimeout sets the Discover service response timeout
func WithResponseTimeout(value time.Duration) coptions.Opt {
	return func(p coptions.Params) {
		logger.Debug("Checking responseTimeoutSetter")
		if setter, ok := p.(responseTimeoutSetter); ok {
			setter.SetSelectionResponseTimeout(value)
		}
	}
}

// WithRetryOpts sets retry options for retries on transient errors
// from the Discovery Server
func WithRetryOpts(value retry.Opts) coptions.Opt {
	return func(p coptions.Params) {
		logger.Debug("Checking retryOptsSetter")
		if setter, ok := p.(retryOptsSetter); ok {
			setter.SetSelectionRetryOpts(value)
		}
	}
}

// WithErrorHandler sets the error handler
func WithErrorHandler(value fab.ErrorHandler) coptions.Opt {
	return func(p coptions.Params) {
		logger.Debug("Checking errHandlerSetter")
		if setter, ok := p.(errHandlerSetter); ok {
			setter.SetErrorHandler(value)
		}
	}
}

type refreshIntervalSetter interface {
	SetSelectionRefreshInterval(value time.Duration)
}

type responseTimeoutSetter interface {
	SetSelectionResponseTimeout(value time.Duration)
}

type retryOptsSetter interface {
	SetSelectionRetryOpts(value retry.Opts)
}

type errHandlerSetter interface {
	SetErrorHandler(value fab.ErrorHandler)
}

func (o *params) SetSelectionRefreshInterval(value time.Duration) {
	logger.Debugf("RefreshInterval: %s", value)
	o.refreshInterval = value
}

func (o *params) SetSelectionResponseTimeout(value time.Duration) {
	logger.Debugf("ResponseTimeout: %s", value)
	o.responseTimeout = value
}

func (o *params) SetSelectionRetryOpts(value retry.Opts) {
	logger.Debugf("RetryOpts: %#v", value)
	o.retryOpts = value
}

func (o *params) SetErrorHandler(value fab.ErrorHandler) {
	logger.Debugf("ErrorHandler: %+v", value)
	o.errHandler = value
}
