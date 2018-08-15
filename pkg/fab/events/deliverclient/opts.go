/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package deliverclient

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
)

type params struct {
	connProvider api.ConnectionProvider
	seekType     seek.Type
	fromBlock    uint64
	respTimeout  time.Duration
}

func defaultParams() *params {
	return &params{
		connProvider: deliverFilteredProvider,
		respTimeout:  5 * time.Second,
	}
}

// WithSeekType specifies the point from which block events are to be received.
func WithSeekType(value seek.Type) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(seekTypeSetter); ok {
			setter.SetSeekType(value)
		}
	}
}

// WithBlockNum specifies the block number from which events are to be received.
// Note that this option is only valid if SeekType is set to SeekFrom.
func WithBlockNum(value uint64) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(fromBlockSetter); ok {
			setter.SetFromBlock(value)
		}
	}
}

type seekTypeSetter interface {
	SetSeekType(value seek.Type)
}

type fromBlockSetter interface {
	SetFromBlock(value uint64)
}

func (p *params) PermitBlockEvents() {
	logger.Debug("PermitBlockEvents")
	p.connProvider = deliverProvider
}

// SetConnectionProvider is only used in unit tests
func (p *params) SetConnectionProvider(connProvider api.ConnectionProvider) {
	logger.Debugf("ConnectionProvider: %#v", connProvider)
	p.connProvider = connProvider
}

func (p *params) SetFromBlock(value uint64) {
	logger.Debugf("FromBlock: %d", value)
	p.fromBlock = value
}

func (p *params) SetSeekType(value seek.Type) {
	logger.Debugf("SeekType: %s", value)
	if value != "" {
		p.seekType = value
	} else {
		logger.Warnf("SeekType must not be empty. Defaulting to %s", p.seekType)
	}
}

func (p *params) SetResponseTimeout(value time.Duration) {
	logger.Debugf("ResponseTimeout: %s", value)
	p.respTimeout = value
}
