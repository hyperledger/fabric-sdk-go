/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lazyref

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
)

// WithIdleExpiration sets the idle-time expiration for the reference.
// The reference is expired after not being accessed for the given duration.
func WithIdleExpiration(value time.Duration) options.Opt {
	return func(p options.Params) {
		logger.Debug("Checking idleExpirationSetter")
		if setter, ok := p.(idleExpirationSetter); ok {
			setter.SetIdleExpiration(value)
		}
	}
}

// WithAbsoluteExpiration sets the expiration time for the reference.
// It will expire after this time period, regardless of whether or not
// it has been recently accessed.
func WithAbsoluteExpiration(value time.Duration) options.Opt {
	return func(p options.Params) {
		logger.Debug("Checking absoluteExpirationSetter")
		if setter, ok := p.(absoluteExpirationSetter); ok {
			setter.SetAbsoluteExpiration(value)
		}
	}
}

// WithExpirationProvider sets the expiration provider, which determines
// the expiration time of the reference
func WithExpirationProvider(expirationProvider ExpirationProvider, expiryType ExpirationType) options.Opt {
	return func(p options.Params) {
		logger.Debug("Checking expirationProviderSetter")
		if setter, ok := p.(expirationProviderSetter); ok {
			setter.SetExpirationProvider(expirationProvider, expiryType)
		}
	}
}

// WithFinalizer sets a finalizer function that is called when the
// reference is closed or if it expires
func WithFinalizer(finalizer Finalizer) options.Opt {
	return func(p options.Params) {
		logger.Debug("Checking finalizerSetter")
		if setter, ok := p.(finalizerSetter); ok {
			setter.SetFinalizer(finalizer)
		}
	}
}

const (
	// InitOnFirstAccess specifies that the reference should be initialized the first time it is accessed
	InitOnFirstAccess time.Duration = time.Duration(-1)

	// InitImmediately specifies that the reference should be initialized immediately after it is created
	InitImmediately time.Duration = time.Duration(0)
)

// WithRefreshInterval specifies that the reference should be proactively refreshed.
// Argument, initialInit, if greater than or equal to 0, indicates that the reference
// should be initialized after this duration. If less than 0, the reference will be
// initialized on first access.
// Argument, refreshPeriod, is the period at which the reference will be refreshed.
// Note that the Finalizer will not be invoked each time the value is refreshed.
func WithRefreshInterval(initialInit, refreshPeriod time.Duration) options.Opt {
	return func(p options.Params) {
		logger.Debug("Checking refreshIntervalSetter")
		if setter, ok := p.(refreshIntervalSetter); ok {
			setter.SetRefreshInterval(initialInit, refreshPeriod)
		}
	}
}

type idleExpirationSetter interface {
	SetIdleExpiration(expiration time.Duration)
}

type absoluteExpirationSetter interface {
	SetAbsoluteExpiration(expiration time.Duration)
}

type expirationProviderSetter interface {
	SetExpirationProvider(expirationProvider ExpirationProvider, expiryType ExpirationType)
}

type finalizerSetter interface {
	SetFinalizer(value Finalizer)
}

type refreshIntervalSetter interface {
	SetRefreshInterval(initialInit, refreshPeriod time.Duration)
}

type params struct {
	initialInit        time.Duration
	finalizer          Finalizer
	expirationProvider ExpirationProvider
	expiryType         ExpirationType
}

func (p *params) SetIdleExpiration(expiration time.Duration) {
	p.expirationProvider = NewSimpleExpirationProvider(expiration)
	p.expiryType = LastAccessed
}

func (p *params) SetAbsoluteExpiration(expiration time.Duration) {
	p.expirationProvider = NewSimpleExpirationProvider(expiration)
	p.expiryType = LastInitialized
}

func (p *params) SetExpirationProvider(expirationProvider ExpirationProvider, expiryType ExpirationType) {
	p.expirationProvider = expirationProvider
	p.expiryType = expiryType
}

func (p *params) SetFinalizer(value Finalizer) {
	p.finalizer = value
}

func (p *params) SetRefreshInterval(initialInit, refreshPeriod time.Duration) {
	p.expiryType = Refreshing
	p.expirationProvider = NewSimpleExpirationProvider(refreshPeriod)
	p.initialInit = initialInit
}
