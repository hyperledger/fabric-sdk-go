/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chconfig

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazycache"
	"github.com/pkg/errors"
)

// Provider provides ChannelConfig
type Provider func(channelID string) (fab.ChannelConfig, error)

// CacheKey channel config reference cache key
type CacheKey interface {
	lazycache.Key
	Context() fab.ClientContext
	ChannelID() string
	Provider() Provider
}

// CacheKey holds a key for the provider cache
type cacheKey struct {
	key       string
	channelID string
	context   fab.ClientContext
	pvdr      Provider
}

// NewCacheKey returns a new CacheKey
func NewCacheKey(ctx fab.ClientContext, pvdr Provider, channelID string) (CacheKey, error) {
	return &cacheKey{
		key:       channelID,
		channelID: channelID,
		context:   ctx,
		pvdr:      pvdr,
	}, nil
}

// NewRefCache a cache of channel config references that refreshed with the
// given interval
func NewRefCache(opts ...options.Opt) *lazycache.Cache {
	initializer := func(key lazycache.Key) (interface{}, error) {
		ck, ok := key.(CacheKey)
		if !ok {
			return nil, errors.New("unexpected cache key")
		}
		return NewRef(ck.Context(), ck.Provider(), ck.ChannelID(), opts...), nil
	}

	return lazycache.New("Channel_Cfg_Cache", initializer)
}

// String returns the key as a string
func (k *cacheKey) String() string {
	return k.key
}

// Context returns the Context
func (k *cacheKey) Context() fab.ClientContext {
	return k.context
}

// ChannelID returns the channelID
func (k *cacheKey) ChannelID() string {
	return k.channelID
}

// Provider channel configuration provider
func (k *cacheKey) Provider() Provider {
	return k.pvdr
}
