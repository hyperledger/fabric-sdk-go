/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package deliverclient

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
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
	}
}

func (p *params) SetResponseTimeout(value time.Duration) {
	logger.Debugf("ResponseTimeout: %s", value)
	p.respTimeout = value
}

func (p *params) SetSnapshot(value fab.EventSnapshot) error {
	logger.Debugf("EventSnapshot.LastBlockReceived: %d", value.LastBlockReceived)
	p.SetSeekType(seek.FromBlock)
	// Set 'from block' as the last block received. We may get a duplicate block but, if we
	// ask for the next block and there are no more blocks on the channel, then we'll get an
	// error from the deliver service.
	// TODO: The client should be enhanced to handle this situation more gracefully. It should first
	// try LastBlockReceived+1 and then LastBlockReceived (if an error is received from the deliver server).
	p.SetFromBlock(value.LastBlockReceived())
	return nil
}
