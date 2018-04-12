/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabpvdr

import (
	"crypto/sha256"
	"strconv"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// CacheKey holds a key for the provider cache
type CacheKey struct {
	key      string
	context  fab.ClientContext
	chConfig fab.ChannelCfg
	opts     []options.Opt
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

// NewCacheKey returns a new CacheKey
func NewCacheKey(ctx fab.ClientContext, chConfig fab.ChannelCfg, opts ...options.Opt) (*CacheKey, error) {
	identity, err := ctx.Serialize()
	if err != nil {
		return nil, err
	}

	params := defaultParams()
	options.Apply(params, opts)

	h := sha256.New()
	h.Write(append(identity, []byte(params.getOptKey())...)) // nolint
	hash := h.Sum([]byte(chConfig.ID()))

	return &CacheKey{
		key:      string(hash),
		context:  ctx,
		chConfig: chConfig,
		opts:     opts,
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

// Opts returns the options to use for creating events service
func (k *CacheKey) Opts() []options.Opt {
	return k.opts
}
