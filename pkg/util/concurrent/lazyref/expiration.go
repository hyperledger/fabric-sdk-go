/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lazyref

import "time"

// NewSimpleExpirationProvider returns an expiration provider
// that sets the given expiration period
func NewSimpleExpirationProvider(expiry time.Duration) ExpirationProvider {
	return func() time.Duration {
		return expiry
	}
}

// NewGraduatingExpirationProvider returns an expiration provider
// that has an initial expiration and then expires in graduated increments
// with a maximum expiration time.
func NewGraduatingExpirationProvider(initialExpiry, increments, maxExpiry time.Duration) ExpirationProvider {
	var iteration uint32
	return func() time.Duration {
		expiry := initialExpiry + time.Duration(iteration)*increments
		if expiry > maxExpiry {
			return maxExpiry
		}
		iteration++
		return expiry
	}
}
