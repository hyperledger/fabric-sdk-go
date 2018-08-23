/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/stretchr/testify/assert"
)

func TestDefaultChannelWithDefaultChannelConfiguredAndNoMatchers(t *testing.T) {

	// Default channel and no channel matchers test
	defaultChannelBackend, err := config.FromFile(configEmbeddedUsersTestFilePath)()
	assert.Nil(t, err, "Failed to get backend")

	endpointConfig, err := ConfigFromBackend(defaultChannelBackend...)
	assert.Nil(t, err, "Failed to get endpoint config from backend")
	assert.NotNil(t, endpointConfig, "expected valid endpointconfig")

	//When channel is not defined it should take values from "_default"
	chConfig, ok := endpointConfig.ChannelConfig("test")
	assert.True(t, ok)
	assert.NotNil(t, chConfig)
	assert.Equal(t, 1, chConfig.Policies.QueryChannelConfig.MinResponses)
	assert.Equal(t, 3, chConfig.Policies.QueryChannelConfig.MaxTargets)
	assert.Equal(t, 1, chConfig.Policies.Discovery.MinResponses)
	assert.Equal(t, 3, chConfig.Policies.Discovery.MaxTargets)

	//When channel is not defined it should take channel peers from "_default"
	chPeers, ok := endpointConfig.ChannelPeers("test")
	assert.True(t, ok)
	assert.NotNil(t, chPeers)
	assert.Equal(t, 1, len(chPeers))

	//When channel is not defined it should take channel orderers from "_default"
	chOrderers, ok := endpointConfig.ChannelOrderers("test")
	assert.True(t, ok)
	assert.NotNil(t, chOrderers)
	assert.Equal(t, 1, len(chOrderers))
}

func TestDefaultChannelWithDefaultChannelConfiguredAndChannelMatchers(t *testing.T) {

	// Default channel and channel matchers test
	endpointConfig, err := ConfigFromBackend(getMatcherConfig())
	assert.Nil(t, err, "Failed to get endpoint config from backend")
	assert.NotNil(t, endpointConfig, "expected valid endpointconfig")

	//When channel is not defined and it fails matchers it should take values from "_default"
	chConfig, ok := endpointConfig.ChannelConfig("test")
	assert.True(t, ok)
	assert.NotNil(t, chConfig)
	assert.Equal(t, 1, chConfig.Policies.QueryChannelConfig.MinResponses)
	assert.Equal(t, 3, chConfig.Policies.QueryChannelConfig.MaxTargets)
	assert.Equal(t, 1, chConfig.Policies.Discovery.MinResponses)
	assert.Equal(t, 3, chConfig.Policies.Discovery.MaxTargets)

	//When channel is not defined and it passes matchers it should take values from matched channel
	chConfig, ok = endpointConfig.ChannelConfig("sampleachannel")
	assert.True(t, ok)
	assert.NotNil(t, chConfig)
	// Discovery comes from 'ch1' channel
	assert.Equal(t, 1, chConfig.Policies.Discovery.MinResponses)
	assert.Equal(t, 4, chConfig.Policies.Discovery.MaxTargets)
	// QueryChannelConfig policy contains defaults since not defined in 'ch1'
	assert.Equal(t, 1, chConfig.Policies.QueryChannelConfig.MinResponses)
	assert.Equal(t, 3, chConfig.Policies.QueryChannelConfig.MaxTargets)
}

func TestDefaultChannelWithNoDefaultChannelConfiguredAndNoMatchers(t *testing.T) {

	// Test no default channel + no channel matchers
	endpointConfig, err := ConfigFromBackend(configBackend)
	assert.Nil(t, err, "Failed to get endpoint config from backend")
	assert.NotNil(t, endpointConfig, "expected valid endpointconfig")

	// Channel 'test' is not configured and since there's no default channel in config all tests should 'fail'
	chConfig, ok := endpointConfig.ChannelConfig("test")
	assert.False(t, ok)
	assert.Nil(t, chConfig)

	chPeers, ok := endpointConfig.ChannelPeers("test")
	assert.False(t, ok)
	assert.Nil(t, chPeers)

	chOrderers, ok := endpointConfig.ChannelOrderers("test")
	assert.False(t, ok)
	assert.Nil(t, chOrderers)
}

func TestDefaultChannelWithNoDefaultChannelConfiguredAndWithMatchers(t *testing.T) {

	// Test no default channel + channel matchers
	backends, err := getBackendsFromFiles(sampleMatchersOverrideAll, configTestFilePath)
	assert.Nil(t, err, "not supposed to get error")
	assert.Equal(t, 2, len(backends))

	endpointConfig, err := ConfigFromBackend(backends...)
	assert.Nil(t, err, "not supposed to get error")
	assert.NotNil(t, endpointConfig)

	// Channel 'abc' is not configured and since there's no channel match and no default channel in config all tests should 'fail'
	chConfig, ok := endpointConfig.ChannelConfig("abc")
	assert.False(t, ok)
	assert.Nil(t, chConfig)

	chPeers, ok := endpointConfig.ChannelPeers("abc")
	assert.False(t, ok)
	assert.Nil(t, chPeers)

	chOrderers, ok := endpointConfig.ChannelOrderers("abc")
	assert.False(t, ok)
	assert.Nil(t, chOrderers)

	// Channel 'testXYZchannel' is not configured and however there is channel match so no test should fail
	chConfig, ok = endpointConfig.ChannelConfig("testXYZchannel")
	assert.True(t, ok, "supposed to find channel config")
	assert.Equal(t, 1, len(chConfig.Orderers))
	assert.Equal(t, 1, len(chConfig.Peers))
	assert.Equal(t, 1, chConfig.Policies.QueryChannelConfig.MinResponses)
	assert.Equal(t, 1, chConfig.Policies.QueryChannelConfig.MaxTargets)
	assert.Equal(t, 1, chConfig.Policies.Discovery.MinResponses)
	assert.Equal(t, 1, chConfig.Policies.Discovery.MaxTargets)

	chPeers, ok = endpointConfig.ChannelPeers("testXYZchannel")
	assert.True(t, ok)
	assert.NotNil(t, chPeers)
	assert.Equal(t, 1, len(chPeers))

	chOrderers, ok = endpointConfig.ChannelOrderers("testXYZchannel")
	assert.True(t, ok)
	assert.NotNil(t, chOrderers)
	assert.Equal(t, 1, len(chOrderers))

}

func TestMissingDiscoveryPolicesInfo(t *testing.T) {

	// Default channel and no channel matchers test
	defaultChannelBackend, err := config.FromFile(configEmbeddedUsersTestFilePath)()
	assert.Nil(t, err, "Failed to get backend")

	endpointConfig, err := ConfigFromBackend(defaultChannelBackend...)
	assert.Nil(t, err, "Failed to get endpoint config from backend")
	assert.NotNil(t, endpointConfig, "expected valid endpointconfig")

	//When channel is not defined it should take values from "_default"
	defChConfig, ok := endpointConfig.ChannelConfig("test")
	assert.True(t, ok)
	assert.NotNil(t, defChConfig)

	chConfig, ok := endpointConfig.ChannelConfig("orgchannel")
	assert.True(t, ok)
	assert.NotNil(t, chConfig)

	// Org channel is missing polices in config (should be equal to default channel)
	assert.Equal(t, defChConfig.Policies.Discovery.MaxTargets, chConfig.Policies.Discovery.MaxTargets)
	assert.Equal(t, defChConfig.Policies.Discovery.MinResponses, chConfig.Policies.Discovery.MinResponses)
	assert.Equal(t, defChConfig.Policies.Discovery.RetryOpts.Attempts, chConfig.Policies.Discovery.RetryOpts.Attempts)
	assert.Equal(t, defChConfig.Policies.Discovery.RetryOpts.InitialBackoff, chConfig.Policies.Discovery.RetryOpts.InitialBackoff)
	assert.Equal(t, defChConfig.Policies.Discovery.RetryOpts.BackoffFactor, chConfig.Policies.Discovery.RetryOpts.BackoffFactor)
	assert.Equal(t, defChConfig.Policies.Discovery.RetryOpts.MaxBackoff, chConfig.Policies.Discovery.RetryOpts.MaxBackoff)

	assert.Equal(t, defChConfig.Policies.QueryChannelConfig.MaxTargets, chConfig.Policies.QueryChannelConfig.MaxTargets)
	assert.Equal(t, defChConfig.Policies.QueryChannelConfig.MinResponses, chConfig.Policies.QueryChannelConfig.MinResponses)
	assert.Equal(t, defChConfig.Policies.QueryChannelConfig.RetryOpts.Attempts, chConfig.Policies.QueryChannelConfig.RetryOpts.Attempts)
	assert.Equal(t, defChConfig.Policies.QueryChannelConfig.RetryOpts.InitialBackoff, chConfig.Policies.QueryChannelConfig.RetryOpts.InitialBackoff)
	assert.Equal(t, defChConfig.Policies.QueryChannelConfig.RetryOpts.BackoffFactor, chConfig.Policies.QueryChannelConfig.RetryOpts.BackoffFactor)
	assert.Equal(t, defChConfig.Policies.QueryChannelConfig.RetryOpts.MaxBackoff, chConfig.Policies.QueryChannelConfig.RetryOpts.MaxBackoff)
}

func TestMissingPartialChannelPoliciesInfo(t *testing.T) {

	// Default channel and no channel matchers test
	defaultChannelBackend, err := config.FromFile(configEmbeddedUsersTestFilePath)()
	assert.Nil(t, err, "Failed to get backend")

	endpointConfig, err := ConfigFromBackend(defaultChannelBackend...)
	assert.Nil(t, err, "Failed to get endpoint config from backend")
	assert.NotNil(t, endpointConfig, "expected valid endpointconfig")

	//When channel is not defined it should take values from "_default"
	defChConfig, ok := endpointConfig.ChannelConfig("test")
	assert.True(t, ok)
	assert.NotNil(t, defChConfig)

	chConfig, ok := endpointConfig.ChannelConfig("mychannel")
	assert.True(t, ok)
	assert.NotNil(t, chConfig)

	// My channel is missing max targets and min responses for discovery policy in config (should be equal to default channel)
	assert.Equal(t, defChConfig.Policies.Discovery.MaxTargets, chConfig.Policies.Discovery.MaxTargets)
	assert.Equal(t, defChConfig.Policies.Discovery.MinResponses, chConfig.Policies.Discovery.MinResponses)
	// Retry opts are defined except for initial backoff
	assert.Equal(t, 4, chConfig.Policies.Discovery.RetryOpts.Attempts)
	assert.Equal(t, defChConfig.Policies.Discovery.RetryOpts.InitialBackoff, chConfig.Policies.Discovery.RetryOpts.InitialBackoff)
	assert.Equal(t, 8.0, chConfig.Policies.Discovery.RetryOpts.BackoffFactor)
	assert.Equal(t, "8s", chConfig.Policies.Discovery.RetryOpts.MaxBackoff.String())

	// My channel is missing RetryOpts for channel config policy in config (should be equal to default channel)
	assert.Equal(t, 8, chConfig.Policies.QueryChannelConfig.MaxTargets)
	assert.Equal(t, 8, chConfig.Policies.QueryChannelConfig.MinResponses)
	// Retry opts initial backoff and attempts are defined
	assert.Equal(t, 5, chConfig.Policies.QueryChannelConfig.RetryOpts.Attempts)
	assert.Equal(t, "5s", chConfig.Policies.QueryChannelConfig.RetryOpts.InitialBackoff.String())
	assert.Equal(t, defChConfig.Policies.QueryChannelConfig.RetryOpts.BackoffFactor, chConfig.Policies.QueryChannelConfig.RetryOpts.BackoffFactor)
	assert.Equal(t, defChConfig.Policies.QueryChannelConfig.RetryOpts.MaxBackoff, chConfig.Policies.QueryChannelConfig.RetryOpts.MaxBackoff)
}
