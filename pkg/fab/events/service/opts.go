/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package service

type params struct {
	eventConsumerBufferSize uint
}

func defaultParams() *params {
	return &params{
		eventConsumerBufferSize: 100,
	}
}

func (p *params) SetEventConsumerBufferSize(value uint) {
	logger.Debugf("EventConsumerBufferSize: %d", value)
	p.eventConsumerBufferSize = value
}
