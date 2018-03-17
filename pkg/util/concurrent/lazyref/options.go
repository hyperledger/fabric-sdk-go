/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lazyref

import (
	"time"
)

// Opt is a reference option
type Opt func(ref *Reference)

// WithIdleExpiration sets the idle-time expiration for the reference.
// The reference is expired after not being accessed for the given duration.
func WithIdleExpiration(expiration time.Duration) Opt {
	return func(ref *Reference) {
		ref.expirationProvider = NewSimpleExpirationProvider(expiration)
		ref.expiryType = LastAccessed
	}
}

// WithAbsoluteExpiration sets the expiration time for the reference.
// It will expire after this time period, regardless of whether or not
// it has been recently accessed.
func WithAbsoluteExpiration(expiration time.Duration) Opt {
	return func(ref *Reference) {
		ref.expirationProvider = NewSimpleExpirationProvider(expiration)
		ref.expiryType = LastInitialized
	}
}

// WithExpirationProvider sets the expiration provider, which determines
// the expiration time of the reference
func WithExpirationProvider(expirationProvider ExpirationProvider, expiryType ExpirationType) Opt {
	return func(ref *Reference) {
		ref.expirationProvider = expirationProvider
		ref.expiryType = expiryType
	}
}

// WithFinalizer sets a finalizer function that is called when the
// reference is closed or if it expires
func WithFinalizer(finalizer Finalizer) Opt {
	return func(ref *Reference) {
		ref.finalizer = finalizer
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
func WithRefreshInterval(initialInit, refreshPeriod time.Duration) Opt {
	return func(ref *Reference) {
		ref.expirationHandler = ref.refreshValue
		ref.expiryType = Refreshing
		ref.expirationProvider = NewSimpleExpirationProvider(refreshPeriod)
		ref.initialInit = initialInit
	}
}
