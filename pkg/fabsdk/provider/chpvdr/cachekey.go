/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chpvdr

import (
	"crypto/sha256"
	"strconv"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// ctxtCacheKey is a lazy cache key for the context cache
type ctxtCacheKey struct {
	key     string
	context fab.ClientContext
}

// newCtxtCacheKey returns a new cacheKey
func newCtxtCacheKey(ctx fab.ClientContext) (*ctxtCacheKey, error) {
	identity, err := ctx.Serialize()
	if err != nil {
		return nil, err
	}

	h := sha256.New()
	if _, err := h.Write(identity); err != nil {
		return nil, err
	}

	hash := h.Sum(nil)

	return &ctxtCacheKey{
		key:     string(hash),
		context: ctx,
	}, nil
}

// String returns the key as a string
func (k *ctxtCacheKey) String() string {
	return k.key
}

// cacheKey holds a key for the provider cache
type cacheKey struct {
	channelConfig fab.ChannelCfg
}

// newCacheKey returns a new cacheKey
func newCacheKey(chConfig fab.ChannelCfg) *cacheKey {
	return &cacheKey{
		channelConfig: chConfig,
	}
}

// String returns the key as a string
func (k *cacheKey) String() string {
	return k.channelConfig.ID()
}

// eventCacheKey holds a key for the provider cache
type eventCacheKey struct {
	key           string
	channelConfig fab.ChannelCfg
	opts          []options.Opt
}

// newEventCacheKey returns a new eventCacheKey
func newEventCacheKey(chConfig fab.ChannelCfg, opts ...options.Opt) (*eventCacheKey, error) {
	params := defaultParams()
	options.Apply(params, opts)

	h := sha256.New()
	if _, err := h.Write([]byte(params.getOptKey())); err != nil {
		return nil, err
	}
	hash := h.Sum([]byte(chConfig.ID()))

	return &eventCacheKey{
		channelConfig: chConfig,
		key:           string(hash),
		opts:          opts,
	}, nil
}

// Opts returns the options to use for creating events service
func (k *eventCacheKey) Opts() []options.Opt {
	return k.opts
}

// String returns the key as a string
func (k *eventCacheKey) String() string {
	return k.key
}

type params struct {
	permitBlockEvents bool
}

func defaultParams() *params {
	return &params{}
}

func (p *params) PermitBlockEvents() {
	p.permitBlockEvents = true
}

func (p *params) getOptKey() string {
	//	Construct opts portion
	optKey := "blockEvents:" + strconv.FormatBool(p.permitBlockEvents)
	return optKey
}
