/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package membership

import (
	"crypto/sha256"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazycache"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazyref"

	"github.com/pkg/errors"
)

// CacheKey membership reference cache key
type CacheKey interface {
	lazycache.Key
	Context() Context
	ChannelID() string
	ChConfigRef() *lazyref.Reference
}

// CacheKey holds a key for the cache
type cacheKey struct {
	key         string
	context     Context
	channelID   string
	chConfigRef *lazyref.Reference
}

// NewCacheKey returns a new CacheKey
func NewCacheKey(context Context, chConfigRef *lazyref.Reference, channelID string) (CacheKey, error) {
	h := sha256.New()
	hash := h.Sum([]byte(channelID))

	return &cacheKey{
		key:         string(hash),
		context:     context,
		chConfigRef: chConfigRef,
		channelID:   channelID,
	}, nil
}

// NewRefCache a cache of membership references that refreshed with the
// given interval
func NewRefCache(refresh time.Duration) *lazycache.Cache {
	initializer := func(key lazycache.Key) (interface{}, error) {
		ck, ok := key.(CacheKey)
		if !ok {
			return nil, errors.New("unexpected cache key")
		}
		return NewRef(refresh, ck.Context(), ck.ChConfigRef()), nil
	}

	return lazycache.New("Membership_Cache", initializer)
}

// String returns the key as a string
func (k *cacheKey) String() string {
	return k.key
}

// Context returns the context
func (k *cacheKey) Context() Context {
	return k.context
}

// ChannelID returns the channelID
func (k *cacheKey) ChannelID() string {
	return k.channelID
}

// ChConfigRef returns the channel config reference
func (k *cacheKey) ChConfigRef() *lazyref.Reference {
	return k.chConfigRef
}
