/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package retry

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
)

var logger = logging.NewLogger("fabsdk/common")

// Invocation is the function to be invoked.
type Invocation func() (interface{}, error)

// BeforeRetryHandler is a function that's invoked before
// a retry attempt.
type BeforeRetryHandler func(error)

// RetryableInvoker manages invocations that could return
// errors and retries the invocation on transient errors.
type RetryableInvoker struct {
	handler     Handler
	beforeRetry BeforeRetryHandler
}

// InvokerOpt is an invoker option
type InvokerOpt func(invoker *RetryableInvoker)

// WithBeforeRetry specifies a function to call before a retry attempt
func WithBeforeRetry(beforeRetry BeforeRetryHandler) InvokerOpt {
	return func(invoker *RetryableInvoker) {
		invoker.beforeRetry = beforeRetry
	}
}

// NewInvoker creates a new RetryableInvoker
func NewInvoker(handler Handler, opts ...InvokerOpt) *RetryableInvoker {
	invoker := &RetryableInvoker{
		handler: handler,
	}
	for _, opt := range opts {
		opt(invoker)
	}
	return invoker
}

// Invoke invokes the given function and performs retries according
// to the retry options.
func (ri *RetryableInvoker) Invoke(invocation Invocation) (interface{}, error) {
	attemptNum := 0
	var lastErr error

	for {
		attemptNum++
		if attemptNum > 1 {
			logger.Debugf("Retry attempt #%d on error [%s]", attemptNum, lastErr)
		}

		retval, err := invocation()
		if err == nil {
			if attemptNum > 1 {
				logger.Debugf("Success on attempt #%d after error [%s]", attemptNum, lastErr)
			}
			return retval, nil
		}

		logger.Debugf("Failed with err [%s] on attempt #%d. Checking if retry is warranted...", err, attemptNum)
		if !ri.resolveRetry(err) {
			if lastErr != nil && lastErr.Error() != err.Error() {
				logger.Debugf("... retry for err [%s] is NOT warranted after %d attempt(s). Previous error [%s]", err, attemptNum, lastErr)
			} else {
				logger.Debugf("... retry for err [%s] is NOT warranted after %d attempt(s).", err, attemptNum)
			}
			return nil, err
		}
		logger.Debugf("... retry for err [%s] is warranted", err)
		lastErr = err
	}
}

func (ri *RetryableInvoker) resolveRetry(err error) bool {
	errs, ok := err.(multi.Errors)
	if !ok {
		errs = append(errs, err)
	}
	for _, e := range errs {
		if ri.handler.Required(e) {
			logger.Debugf("Retrying on error %s", e)
			if ri.beforeRetry != nil {
				ri.beforeRetry(err)
			}
			return true
		}
	}
	return false
}
