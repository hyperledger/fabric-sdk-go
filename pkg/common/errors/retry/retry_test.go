/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package retry

import (
	"fmt"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/stretchr/testify/assert"
)

func TestRetryRequired(t *testing.T) {
	attempts := 3
	transientErr := status.New(status.EndorserClientStatus,
		status.EndorsementMismatch.ToInt32(), "", nil)
	nonTransientErr := status.New(status.EndorserServerStatus,
		int32(common.Status_BAD_REQUEST), "", nil)
	unknownErr := fmt.Errorf("Unknown")

	r := New(Opts{
		Attempts:       attempts,
		BackoffFactor:  2,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     1 * time.Second,
	})
	for i := 1; i <= attempts; i++ {
		assert.True(t, r.Required(transientErr), "Expected retry to be required on transient error")
	}
	assert.False(t, r.Required(transientErr), "Expected retry to not be required after exhausting attempts")
	r = WithDefaults()
	assert.False(t, r.Required(nonTransientErr), "Expected retry to not be required on non-transient error")
	r = WithAttempts(2)
	assert.False(t, r.Required(unknownErr), "Expected retry to not be required on unknown error")
}

func TestBackoffPeriod(t *testing.T) {
	testAttempts := 10
	testBackoffFactor := 3.34
	testInitialBackoff := 2 * time.Second
	floatInitBackoff := float64(testInitialBackoff)
	testMaxBackoff := 30 * time.Second
	r := New(Opts{
		Attempts:       testAttempts,
		BackoffFactor:  testBackoffFactor,
		InitialBackoff: testInitialBackoff,
		MaxBackoff:     testMaxBackoff,
	})
	i := r.(*impl)
	assert.Equal(t, testInitialBackoff, i.backoffPeriod(), "Expected initial backoff on first attempt")
	i.retries = 1
	assert.Equal(t, time.Duration(floatInitBackoff*testBackoffFactor), i.backoffPeriod(),
		"Expected initial backoff multiplied by backoff factor on second attempt")
	i.retries = 2
	assert.Equal(t, time.Duration(floatInitBackoff*testBackoffFactor*testBackoffFactor),
		i.backoffPeriod(), "Expected exponential backoff")
	i.retries = 3
	assert.Equal(t, testMaxBackoff, i.backoffPeriod(), "Expected max backoff")
}
