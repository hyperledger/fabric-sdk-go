// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabpvdr

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazycache"
)

// SetChannelConfig allows setting channel configuration.
// This method is intended to enable tests and should not be called.
func (f *InfraProvider) SetChannelConfig(cfg fab.ChannelCfg) {
	f.chCfgCache = newMockCache(cfg)
}

type mockCache struct {
	cfg fab.ChannelCfg
}

func newMockCache(cfg fab.ChannelCfg) cache {
	return &mockCache{cfg: cfg}
}

func (m *mockCache) Get(lazycache.Key) (interface{}, error) {
	return m.cfg, nil
}
func (m *mockCache) Close() {
}
