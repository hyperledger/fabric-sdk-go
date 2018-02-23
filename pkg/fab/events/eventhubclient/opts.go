/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package eventhubclient

import "time"

type params struct {
	respTimeout time.Duration
}

func defaultParams() *params {
	return &params{
		respTimeout: 5 * time.Second,
	}
}

func (p *params) SetResponseTimeout(value time.Duration) {
	logger.Debugf("ResponseTimeout: %s", value)
	p.respTimeout = value
}
