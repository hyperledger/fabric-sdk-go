// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chpvdr

import (
	"sync"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/chconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazycache"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazyref"
	"github.com/pkg/errors"
)

// SetChannelConfig allows setting channel configuration.
// This method is intended to enable tests and should not be called.
func (f *ChannelProvider) SetChannelConfig(cfg fab.ChannelCfg) {
	if _, ok := f.chCfgCache.(*chCfgCache); !ok {
		f.chCfgCache = newMockChCfgCache(cfg)
	} else {
		f.chCfgCache.(*chCfgCache).Put(cfg)
	}
}

type chCfgCache struct {
	cfgMap sync.Map
}

func newMockChCfgCache(cfg fab.ChannelCfg) *chCfgCache {
	c := &chCfgCache{}
	c.cfgMap.Store(cfg.ID(), newChCfgRef(cfg))
	return c
}

func newChCfgRef(cfg fab.ChannelCfg) *chconfig.Ref {
	r := &chconfig.Ref{}
	r.Reference = lazyref.New(func() (interface{}, error) {
		return cfg, nil
	})
	return r
}

// Get mock channel config reference
func (m *chCfgCache) Get(k lazycache.Key) (interface{}, error) {
	cfg, ok := m.cfgMap.Load(k.(chconfig.CacheKey).ChannelID())
	if !ok {
		return nil, errors.New("Channel config not found in cache")
	}
	return cfg, nil
}

// Close not implemented
func (m *chCfgCache) Close() {
}

// Put channel config reference into mock cache
func (m *chCfgCache) Put(cfg fab.ChannelCfg) {
	m.cfgMap.Store(cfg.ID(), newChCfgRef(cfg))
}
