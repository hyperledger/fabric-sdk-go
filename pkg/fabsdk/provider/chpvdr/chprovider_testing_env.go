// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chpvdr

import "github.com/hyperledger/fabric-sdk-go/api/apifabclient"

// SetChannelConfig allows setting channel configuration.
// This method is intended to enable tests and should not be called.
func (cp *ChannelProvider) SetChannelConfig(cfg apifabclient.ChannelCfg) {
	cp.chCfgMap.Store(cfg.Name(), cfg)
}
