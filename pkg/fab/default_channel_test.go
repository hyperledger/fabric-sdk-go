/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/stretchr/testify/assert"
)

func TestDefaultChannelWithDefaultChannelConfiguredAndNoMatchers(t *testing.T) {

	// Default channel and no channel matchers test
	configPath := filepath.Join(getConfigPath(), configEmbeddedUsersTestFile)
	defaultChannelBackend, err := config.FromFile(configPath)()
	assert.Nil(t, err, "Failed to get backend")

	endpointConfig, err := ConfigFromBackend(defaultChannelBackend...)
	assert.Nil(t, err, "Failed to get endpoint config from backend")
	assert.NotNil(t, endpointConfig, "expected valid endpointconfig")

	//When channel is not defined it should take values from "_default"
	chConfig := endpointConfig.ChannelConfig("test")
	assert.NotNil(t, chConfig)
	assert.Equal(t, 1, chConfig.Policies.QueryChannelConfig.MinResponses)
	assert.Equal(t, 3, chConfig.Policies.QueryChannelConfig.MaxTargets)
	assert.Equal(t, 1, chConfig.Policies.Discovery.MinResponses)
	assert.Equal(t, 3, chConfig.Policies.Discovery.MaxTargets)
	assert.Equal(t, 2, chConfig.Policies.Discovery.RetryOpts.Attempts)
	assert.Equal(t, 2*time.Second, chConfig.Policies.Discovery.RetryOpts.InitialBackoff)
	assert.Equal(t, 7*time.Second, chConfig.Policies.Discovery.RetryOpts.MaxBackoff)
	assert.Equal(t, 2, len(chConfig.Policies.Discovery.RetryOpts.RetryableCodes))

	eventPolicies := chConfig.Policies.EventService
	assert.Equalf(t, fab.BalancedStrategy, eventPolicies.ResolverStrategy, "Unexpected value for ResolverStrategy")
	assert.Equal(t, fab.RoundRobin, eventPolicies.Balancer, "Unexpected value for Balancer")
	assert.Equal(t, 3, eventPolicies.BlockHeightLagThreshold, "Unexpected value for BlockHeightLagThreshold")
	assert.Equal(t, 7, eventPolicies.ReconnectBlockHeightLagThreshold, "Unexpected value for ReconnectBlockHeightLagThreshold")
	assert.Equal(t, 8*time.Second, eventPolicies.PeerMonitorPeriod, "Unexpected value for PeerMonitorPeriod")

	//When channel is not defined it should take channel peers from "_default"
	chPeers := endpointConfig.ChannelPeers("test")
	assert.NotNil(t, chPeers)
	assert.Equal(t, 1, len(chPeers))

	//When channel is not defined it should take channel orderers from "_default"
	chOrderers := endpointConfig.ChannelOrderers("test")
	assert.Equal(t, 1, len(chOrderers))
}

func TestDefaultChannelWithDefaultChannelConfiguredAndChannelMatchers(t *testing.T) {

	// Default channel and channel matchers test
	endpointConfig, err := ConfigFromBackend(getMatcherConfig())
	assert.Nil(t, err, "Failed to get endpoint config from backend")
	assert.NotNil(t, endpointConfig, "expected valid endpointconfig")

	//When channel is not defined and it fails matchers it should take values from "_default"
	chConfig := endpointConfig.ChannelConfig("test")
	assert.NotNil(t, chConfig)
	assert.Equal(t, 1, chConfig.Policies.QueryChannelConfig.MinResponses)
	assert.Equal(t, 3, chConfig.Policies.QueryChannelConfig.MaxTargets)
	assert.Equal(t, 1, chConfig.Policies.Discovery.MinResponses)
	assert.Equal(t, 3, chConfig.Policies.Discovery.MaxTargets)

	//When channel is not defined and it passes matchers it should take values from matched channel
	chConfig = endpointConfig.ChannelConfig("sampleachannel")
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

	// Channel 'test' is not configured and since there's no default channel in config
	// we should be using the hard-coded default config
	chConfig := endpointConfig.ChannelConfig("test")
	assert.NotNil(t, chConfig)

	chPeers := endpointConfig.ChannelPeers("test")
	assert.Empty(t, len(chPeers))

	chOrderers := endpointConfig.ChannelOrderers("test")
	assert.Empty(t, chOrderers)
}

func TestDefaultChannelWithNoDefaultChannelConfiguredAndWithMatchers(t *testing.T) {

	// Test no default channel + channel matchers
	matcherPath := filepath.Join(getConfigPath(), matchersDir, sampleMatchersOverrideAll)
	configPath := filepath.Join(getConfigPath(), configTestFile)
	backends, err := getBackendsFromFiles(matcherPath, configPath)
	assert.Nil(t, err, "not supposed to get error")
	assert.Equal(t, 2, len(backends))

	endpointConfig, err := ConfigFromBackend(backends...)
	assert.Nil(t, err, "not supposed to get error")
	assert.NotNil(t, endpointConfig)

	// Channel 'abc' is not configured and since there's no default channel in config
	// we should be using the hard-coded default config
	chConfig := endpointConfig.ChannelConfig("abc")
	assert.NotNil(t, chConfig)

	chPeers := endpointConfig.ChannelPeers("abc")
	assert.Empty(t, chPeers)

	chOrderers := endpointConfig.ChannelOrderers("abc")
	assert.Empty(t, chOrderers)

	// Channel 'testXYZchannel' is not configured and however there is channel match so no test should fail
	chConfig = endpointConfig.ChannelConfig("testXYZchannel")
	assert.Equal(t, 1, len(chConfig.Orderers))
	assert.Equal(t, 1, len(chConfig.Peers))
	assert.Equal(t, 1, chConfig.Policies.QueryChannelConfig.MinResponses)
	assert.Equal(t, 2, chConfig.Policies.QueryChannelConfig.MaxTargets)
	assert.Equal(t, 1, chConfig.Policies.Discovery.MinResponses)
	assert.Equal(t, 2, chConfig.Policies.Discovery.MaxTargets)

	chPeers = endpointConfig.ChannelPeers("testXYZchannel")
	assert.Equal(t, 1, len(chPeers))

	chOrderers = endpointConfig.ChannelOrderers("testXYZchannel")
	assert.Equal(t, 1, len(chOrderers))

}

func TestMissingPolicesInfo(t *testing.T) {

	// Default channel and no channel matchers test
	configPath := filepath.Join(getConfigPath(), configEmbeddedUsersTestFile)
	defaultChannelBackend, err := config.FromFile(configPath)()
	assert.Nil(t, err, "Failed to get backend")

	endpointConfig, err := ConfigFromBackend(defaultChannelBackend...)
	assert.Nil(t, err, "Failed to get endpoint config from backend")
	assert.NotNil(t, endpointConfig, "expected valid endpointconfig")

	//When channel is not defined it should take values from "_default"
	defChConfig := endpointConfig.ChannelConfig("test")
	assert.NotNil(t, defChConfig)

	chConfig := endpointConfig.ChannelConfig("orgchannel")
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

	assert.Equal(t, defChConfig.Policies.EventService.Balancer, chConfig.Policies.EventService.Balancer)
	assert.Equal(t, defChConfig.Policies.EventService.MinBlockHeightResolverMode, chConfig.Policies.EventService.MinBlockHeightResolverMode)
	assert.Equal(t, defChConfig.Policies.EventService.BlockHeightLagThreshold, chConfig.Policies.EventService.BlockHeightLagThreshold)
	assert.Equal(t, defChConfig.Policies.EventService.PeerMonitorPeriod, chConfig.Policies.EventService.PeerMonitorPeriod)
	assert.Equal(t, defChConfig.Policies.EventService.PeerMonitor, chConfig.Policies.EventService.PeerMonitor)
	assert.Equal(t, defChConfig.Policies.EventService.ReconnectBlockHeightLagThreshold, chConfig.Policies.EventService.ReconnectBlockHeightLagThreshold)
	assert.Equal(t, defChConfig.Policies.EventService.ResolverStrategy, chConfig.Policies.EventService.ResolverStrategy)

	chConfig = endpointConfig.ChannelConfig("mychannel")
	assert.NotNil(t, chConfig)

	assert.Equal(t, fab.ResolveLatest, chConfig.Policies.EventService.MinBlockHeightResolverMode)
	assert.Equal(t, fab.Disabled, chConfig.Policies.EventService.PeerMonitor)
	assert.Equal(t, defChConfig.Policies.EventService.Balancer, chConfig.Policies.EventService.Balancer)
	assert.Equal(t, defChConfig.Policies.EventService.PeerMonitorPeriod, chConfig.Policies.EventService.PeerMonitorPeriod)
	assert.Equal(t, defChConfig.Policies.EventService.ReconnectBlockHeightLagThreshold, chConfig.Policies.EventService.ReconnectBlockHeightLagThreshold)
	assert.Equal(t, defChConfig.Policies.EventService.ResolverStrategy, chConfig.Policies.EventService.ResolverStrategy)
}

func TestMissingPartialChannelPoliciesInfo(t *testing.T) {

	// Default channel and no channel matchers test
	configPath := filepath.Join(getConfigPath(), configEmbeddedUsersTestFile)
	defaultChannelBackend, err := config.FromFile(configPath)()
	assert.Nil(t, err, "Failed to get backend")

	endpointConfig, err := ConfigFromBackend(defaultChannelBackend...)
	assert.Nil(t, err, "Failed to get endpoint config from backend")
	assert.NotNil(t, endpointConfig, "expected valid endpointconfig")

	//When channel is not defined it should take values from "_default"
	defChConfig := endpointConfig.ChannelConfig("test")
	assert.NotNil(t, defChConfig)

	chConfig := endpointConfig.ChannelConfig("mychannel")
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

func TestMissingPeersInfo(t *testing.T) {

	// Default channel and no channel matchers test
	configPath := filepath.Join(getConfigPath(), configEmbeddedUsersTestFile)
	defaultChannelBackend, err := config.FromFile(configPath)()
	assert.Nil(t, err, "Failed to get backend")

	endpointConfig, err := ConfigFromBackend(defaultChannelBackend...)
	assert.Nil(t, err, "Failed to get endpoint config from backend")
	assert.NotNil(t, endpointConfig, "expected valid endpointconfig")

	defChConfig := endpointConfig.ChannelConfig("test")
	assert.NotNil(t, defChConfig)

	//If peers are not defined for channel then peers should be filled in from "_default" channel
	chConfig := endpointConfig.ChannelConfig("nopeers")
	assert.NotNil(t, chConfig)
	assert.Equal(t, 1, len(chConfig.Orderers))
	assert.Equal(t, 1, len(chConfig.Peers))

	// 'nopeers' channel is missing polices in config (should be equal to default channel)
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

	// Since peers are not defined channel peers should equal peers from "_default"
	chPeers := endpointConfig.ChannelPeers("nopeers")
	assert.Equal(t, 1, len(chPeers))
	assert.True(t, strings.Contains(chPeers[0].URL, "peer0.org2.example.com"))

	// Orderer is defined for channel, verify orderer
	chOrderers := endpointConfig.ChannelOrderers("nopeers")
	assert.Equal(t, 1, len(chOrderers))
	assert.True(t, strings.Contains(chOrderers[0].URL, "orderer2.example.com"))

}

func TestMissingOrderersInfo(t *testing.T) {

	// Default channel and no channel matchers test
	configPath := filepath.Join(getConfigPath(), configEmbeddedUsersTestFile)
	defaultChannelBackend, err := config.FromFile(configPath)()
	assert.Nil(t, err, "Failed to get backend")

	endpointConfig, err := ConfigFromBackend(defaultChannelBackend...)
	assert.Nil(t, err, "Failed to get endpoint config from backend")
	assert.NotNil(t, endpointConfig, "expected valid endpointconfig")

	defChConfig := endpointConfig.ChannelConfig("test")
	assert.NotNil(t, defChConfig)

	//If orderers are not defined for channel then orderers should be filled in from "_default" channel
	chConfig := endpointConfig.ChannelConfig("noorderers")
	assert.NotNil(t, chConfig)
	assert.Equal(t, 1, len(chConfig.Orderers))
	assert.True(t, strings.Contains(chConfig.Orderers[0], "orderer.example.com"))
	assert.Equal(t, 2, len(chConfig.Peers))

	// Channel peers are defined, verify
	chPeers := endpointConfig.ChannelPeers("noorderers")
	assert.Equal(t, 2, len(chPeers))

	//Verify channel orderers are from "_default"
	chOrderers := endpointConfig.ChannelOrderers("noorderers")
	assert.Equal(t, 1, len(chOrderers))
	assert.True(t, strings.Contains(chOrderers[0].URL, "orderer.example.com"))

}
