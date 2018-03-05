/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
)

type params struct {
	eventConsumerBufferSize uint
	eventConsumerTimeout    time.Duration
}

func defaultParams() *params {
	return &params{
		eventConsumerBufferSize: 100,
		eventConsumerTimeout:    500 * time.Millisecond,
	}
}

// WithEventConsumerBufferSize sets the size of the registered consumer's event channel.
func WithEventConsumerBufferSize(value uint) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(eventConsumerBufferSizeSetter); ok {
			setter.SetEventConsumerBufferSize(value)
		}
	}
}

// WithEventConsumerTimeout is the timeout when sending events to a registered consumer.
// If < 0, if buffer full, unblocks immediately and does not send.
// If 0, if buffer full, will block and guarantee the event will be sent out.
// If > 0, if buffer full, blocks util timeout.
func WithEventConsumerTimeout(value time.Duration) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(eventEventConsumerTimeoutSetter); ok {
			setter.SetEventConsumerTimeout(value)
		}
	}
}

type eventConsumerBufferSizeSetter interface {
	SetEventConsumerBufferSize(value uint)
}

type eventEventConsumerTimeoutSetter interface {
	SetEventConsumerTimeout(value time.Duration)
}

func (p *params) SetEventConsumerBufferSize(value uint) {
	logger.Debugf("EventConsumerBufferSize: %d", value)
	p.eventConsumerBufferSize = value
}

func (p *params) SetEventConsumerTimeout(value time.Duration) {
	logger.Debugf("EventConsumerTimeout: %s", value)
	p.eventConsumerTimeout = value
}
