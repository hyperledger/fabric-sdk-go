/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabpvdr

import (
	"crypto/sha256"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// CacheKey holds a key for the provider cache
type CacheKey struct {
	key      string
	context  fab.ClientContext
	chConfig fab.ChannelCfg
}

// NewCacheKey returns a new CacheKey
func NewCacheKey(ctx fab.ClientContext, chConfig fab.ChannelCfg) (*CacheKey, error) {
	identity, err := ctx.Serialize()
	if err != nil {
		return nil, err
	}

	h := sha256.New()
	h.Write(identity)
	hash := h.Sum([]byte(chConfig.ID()))

	return &CacheKey{
		key:      string(hash),
		context:  ctx,
		chConfig: chConfig,
	}, nil
}

// String returns the key as a string
func (k *CacheKey) String() string {
	return k.key
}

// Context returns the Context
func (k *CacheKey) Context() fab.ClientContext {
	return k.context
}

// ChannelConfig returns the channel configuration
func (k *CacheKey) ChannelConfig() fab.ChannelCfg {
	return k.chConfig
}
