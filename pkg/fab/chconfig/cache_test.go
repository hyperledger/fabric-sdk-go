/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chconfig

import (
	"fmt"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/stretchr/testify/assert"

	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
)

const (
	badProviderErrMessage = "bad provider"
)

func TestChannelConfigCache(t *testing.T) {
	user := mspmocks.NewMockSigningIdentity("user", "user")
	clientCtx := mocks.NewMockContext(user)

	cache := NewRefCache(time.Millisecond * 10)
	assert.NotNil(t, cache)

	key, err := NewCacheKey(clientCtx, mockProvider, "test")
	assert.Nil(t, err)
	assert.NotNil(t, key)

	r, err := cache.Get(key)
	assert.Nil(t, err)
	assert.NotNil(t, r)
}

func TestChannelConfigCacheBad(t *testing.T) {
	user := mspmocks.NewMockSigningIdentity("user", "user")
	clientCtx := mocks.NewMockContext(user)

	cache := NewRefCache(time.Millisecond * 10)
	assert.NotNil(t, cache)

	r, err := cache.Get(&badKey{s: "test"})
	assert.NotNil(t, err)
	assert.Equal(t, "unexpected cache key", err.Error())
	assert.Nil(t, r)

	key, err := NewCacheKey(clientCtx, badProvider, "test")
	assert.Nil(t, err)
	assert.NotNil(t, key)

	cache = NewRefCache(time.Millisecond * 10)
	assert.NotNil(t, cache)

	r, err = cache.Get(key)
	assert.NotNil(t, r)
	assert.Nil(t, err)

	c, err := r.(*Ref).Get()
	assert.NotNil(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), badProviderErrMessage)
}

type badKey struct {
	s string
}

func (b *badKey) String() string {
	return b.s
}

func mockProvider(channelID string) (fab.ChannelConfig, error) {
	return mocks.NewMockChannelConfig(nil, channelID)
}

func badProvider(channelID string) (fab.ChannelConfig, error) {
	return nil, fmt.Errorf(badProviderErrMessage)
}
