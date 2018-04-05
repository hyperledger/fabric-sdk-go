/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package retry

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/stretchr/testify/assert"
)

func TestInvokeSuccess(t *testing.T) {
	r := New(Opts{
		Attempts:       3,
		BackoffFactor:  2,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     1 * time.Second,
	})

	attempt := 0
	expectedResp := "invoked"
	invoker := NewInvoker(r)
	resp, err := invoker.Invoke(
		func() (interface{}, error) {
			attempt++
			if attempt == 1 {
				return nil, status.New(status.EndorserClientStatus, status.EndorsementMismatch.ToInt32(), "", nil)
			}
			return expectedResp, nil
		},
	)

	assert.NoError(t, err, "Not expecting error")
	assert.Equal(t, expectedResp, resp)
	assert.Equal(t, 2, attempt)
}

func TestInvokeError(t *testing.T) {
	r := New(Opts{
		Attempts:       3,
		BackoffFactor:  2,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     1 * time.Second,
	})

	attempt := 0
	expectedResp := "invoked"
	firstErr := status.New(status.EndorserClientStatus, status.EndorsementMismatch.ToInt32(), "", nil)
	exepectedErr := status.New(status.ChaincodeStatus, int32(500), "", nil)
	invoker := NewInvoker(r)
	resp, err := invoker.Invoke(
		func() (interface{}, error) {
			attempt++
			if attempt == 1 {
				return nil, firstErr
			}
			if attempt == 2 {
				return nil, exepectedErr
			}
			return expectedResp, nil
		},
	)

	assert.EqualError(t, err, exepectedErr.Error())
	assert.Nil(t, resp)
	assert.Equal(t, 2, attempt)
}

func TestInvokeWithBeforeRetry(t *testing.T) {
	r := New(Opts{
		Attempts:       3,
		BackoffFactor:  2,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     1 * time.Second,
	})

	beforeRetryHandlerCalled := 0
	attempt := 0
	expectedResp := "invoked"
	invoker := NewInvoker(r, WithBeforeRetry(
		func(err error) {
			beforeRetryHandlerCalled++
		},
	))
	resp, err := invoker.Invoke(
		func() (interface{}, error) {
			attempt++
			if attempt == 1 {
				return nil, status.New(status.EndorserClientStatus, status.EndorsementMismatch.ToInt32(), "", nil)
			}
			return expectedResp, nil
		},
	)

	assert.NoError(t, err, "Not expecting error")
	assert.Equal(t, expectedResp, resp)
	assert.Equal(t, 2, attempt)
	assert.Equal(t, 1, beforeRetryHandlerCalled)
}
