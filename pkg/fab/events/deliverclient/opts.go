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
	connProvider      api.ConnectionProvider
	permitBlockEvents bool
	seekType          seek.Type
	fromBlock         uint64
	respTimeout       time.Duration
}

func defaultParams() *params {
	return &params{
		connProvider: deliverFilteredProvider,
		seekType:     seek.Newest,
		respTimeout:  5 * time.Second,
	}
}

// WithBlockEvents indicates that block events are to be received.
// Note that the caller must have sufficient privileges for this option.
func WithBlockEvents() options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(connectionProviderSetter); ok {
			setter.SetConnectionProvider(deliverProvider, true)
		}
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

// withConnectionProvider is used only for testing
func withConnectionProvider(connProvider api.ConnectionProvider, permitBlockEvents bool) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(connectionProviderSetter); ok {
			setter.SetConnectionProvider(connProvider, permitBlockEvents)
		}
	}
}

type connectionProviderSetter interface {
	SetConnectionProvider(value api.ConnectionProvider, permitBlockEvents bool)
}

type seekTypeSetter interface {
	SetSeekType(value seek.Type)
}

type fromBlockSetter interface {
	SetFromBlock(value uint64)
}

func (p *params) SetConnectionProvider(connProvider api.ConnectionProvider, permitBlockEvents bool) {
	logger.Debugf("ConnectionProvider: %#v, PermitBlockEvents: %t", connProvider, permitBlockEvents)
	p.connProvider = connProvider
	p.permitBlockEvents = permitBlockEvents
}

func (p *params) SetFromBlock(value uint64) {
	logger.Debugf("FromBlock: %d", value)
	p.fromBlock = value
}

func (p *params) SetSeekType(value seek.Type) {
	logger.Debugf("SeekType: %s", value)
	p.seekType = value
}

func (p *params) SetResponseTimeout(value time.Duration) {
	logger.Debugf("ResponseTimeout: %s", value)
	p.respTimeout = value
}
