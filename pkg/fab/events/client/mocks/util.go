/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
)

// NewMockConfig returns a mock endpoint config with the given event service policy for the given channel
func NewMockConfig(channelID string, policy fab.EventServicePolicy) *fabmocks.MockConfig {
	config := &fabmocks.MockConfig{}
	config.SetCustomChannelConfig(channelID, &fab.ChannelEndpointConfig{
		Policies: fab.ChannelPolicies{
			EventService: policy,
		},
	})
	return config
}
