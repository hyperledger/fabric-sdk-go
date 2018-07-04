/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lazycache

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazyref"
)

// refOptCheck is used to test whether any of the lazyref options have been passed in
type refOptCheck struct {
	useRef bool
}

func (p *refOptCheck) SetIdleExpiration(expiration time.Duration) {
	p.useRef = true
}

func (p *refOptCheck) SetAbsoluteExpiration(expiration time.Duration) {
	p.useRef = true
}

func (p *refOptCheck) SetExpirationProvider(expirationProvider lazyref.ExpirationProvider, expiryType lazyref.ExpirationType) {
	p.useRef = true
}

func (p *refOptCheck) SetFinalizer(value lazyref.Finalizer) {
	p.useRef = true
}

func (p *refOptCheck) SetRefreshInterval(initialInit, refreshPeriod time.Duration) {
	p.useRef = true
}
