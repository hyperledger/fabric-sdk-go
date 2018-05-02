/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package retry provides retransmission capabilities to fabric-sdk-go.
// The only user interaction with this package is expected to be with the
// defaults defined below.
// They can be used in conjunction with the WithRetry setting offered by certain
// clients in the SDK:
// https://godoc.org/github.com/hyperledger/fabric-sdk-go/pkg/client/channel#WithRetry
// https://godoc.org/github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt#WithRetry
package retry

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
)

// Opts defines the retry parameters
type Opts struct {
	// Attempts the number retry attempts
	Attempts int
	// InitialBackoff the backoff interval for the first retry attempt
	InitialBackoff time.Duration
	// MaxBackoff the maximum backoff interval for any retry attempt
	MaxBackoff time.Duration
	// BackoffFactor the factor by which the InitialBackoff is exponentially
	// incremented for consecutive retry attempts.
	// For example, a backoff factor of 2.5 will result in a backoff of
	// InitialBackoff * 2.5 * 2.5 on the second attempt.
	BackoffFactor float64
	// RetryableCodes defines the status codes, mapped by group, returned by fabric-sdk-go
	// that warrant a retry. This will default to retry.DefaultRetryableCodes.
	RetryableCodes map[status.Group][]status.Code
}

// Handler retry handler interface decides whether a retry is required for the given
// error
type Handler interface {
	Required(err error) bool
}

// impl retry Handler implementation
type impl struct {
	opts    Opts
	retries int
}

// New retry Handler with the given opts
func New(opts Opts) Handler {
	if len(opts.RetryableCodes) == 0 {
		opts.RetryableCodes = DefaultRetryableCodes
	}
	return &impl{opts: opts}
}

// WithDefaults new retry Handler with default opts
func WithDefaults() Handler {
	return &impl{opts: DefaultOpts}
}

// WithAttempts new retry Handler with given attempts. Other opts are set to default.
func WithAttempts(attempts int) Handler {
	opts := DefaultOpts
	opts.Attempts = attempts
	return &impl{opts: opts}
}

// Required determines if retry is required for the given error
// Note: backoffs are implemented behind this interface
func (i *impl) Required(err error) bool {
	if i.retries == i.opts.Attempts {
		return false
	}

	s, ok := status.FromError(err)
	if ok && i.isRetryable(s.Group, s.Code) {
		time.Sleep(i.backoffPeriod())
		i.retries++
		return true
	}

	return false
}

// backoffPeriod calculates the backoff duration based on the provided opts
func (i *impl) backoffPeriod() time.Duration {
	backoff, max := float64(i.opts.InitialBackoff), float64(i.opts.MaxBackoff)
	for j := 0; j < i.retries && backoff < max; j++ {
		backoff *= i.opts.BackoffFactor
	}
	if backoff > max {
		backoff = max
	}

	return time.Duration(backoff)
}

// isRetryable determines if the given status is configured to be retryable
func (i *impl) isRetryable(g status.Group, c int32) bool {
	for group, codes := range i.opts.RetryableCodes {
		if g != group {
			continue
		}
		for _, code := range codes {
			if status.Code(c) == code {
				return true
			}
		}
	}
	return false
}
