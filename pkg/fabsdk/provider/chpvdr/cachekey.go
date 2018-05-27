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

// cacheKey holds a key for the provider cache
type cacheKey struct {
	key           string
	context       fab.ClientContext
	channelConfig fab.ChannelCfg
}

// newCacheKey returns a new cacheKey
func newCacheKey(ctx fab.ClientContext, chConfig fab.ChannelCfg, opts ...options.Opt) (*cacheKey, error) {
	identity, err := ctx.Serialize()
	if err != nil {
		return nil, err
	}

	params := defaultParams()
	options.Apply(params, opts)

	h := sha256.New()
	h.Write(append(identity, []byte(params.getOptKey())...)) // nolint
	hash := h.Sum([]byte(chConfig.ID()))

	return &cacheKey{
		key:           string(hash),
		context:       ctx,
		channelConfig: chConfig,
	}, nil
}

// String returns the key as a string
func (k *cacheKey) String() string {
	return k.key
}

// cacheKey holds a key for the provider cache
type eventCacheKey struct {
	cacheKey
	opts []options.Opt
}

// newEventCacheKey returns a new eventCacheKey
func newEventCacheKey(ctx fab.ClientContext, chConfig fab.ChannelCfg, opts ...options.Opt) (*eventCacheKey, error) {
	identity, err := ctx.Serialize()
	if err != nil {
		return nil, err
	}

	params := defaultParams()
	options.Apply(params, opts)

	h := sha256.New()
	h.Write(append(identity, []byte(params.getOptKey())...)) // nolint
	hash := h.Sum([]byte(chConfig.ID()))

	return &eventCacheKey{
		cacheKey: cacheKey{
			key:           string(hash),
			context:       ctx,
			channelConfig: chConfig,
		},
		opts: opts,
	}, nil
}

// Opts returns the options to use for creating events service
func (k *eventCacheKey) Opts() []options.Opt {
	return k.opts
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
