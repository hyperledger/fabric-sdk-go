/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chconfig

import (
	"time"

	coptions "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

const (
	defaultRefreshInterval = time.Second * 90
)

type params struct {
	refreshInterval time.Duration
	errHandler      fab.ErrorHandler
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
	SetChConfigRefreshInterval(value time.Duration)
}

type errHandlerSetter interface {
	SetErrorHandler(value fab.ErrorHandler)
}

func (o *params) SetChConfigRefreshInterval(value time.Duration) {
	logger.Debugf("RefreshInterval: %s", value)
	o.refreshInterval = value
}

func (o *params) SetErrorHandler(value fab.ErrorHandler) {
	logger.Debugf("ErrorHandler: %+v", value)
	o.errHandler = value
}
