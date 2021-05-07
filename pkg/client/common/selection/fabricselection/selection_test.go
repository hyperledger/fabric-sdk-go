// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricselection

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hyperledger/fabric-protos-go/gossip"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/balancer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/sorter/blockheightsorter"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	fab "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery"
	discmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	channelID = "testchannel"
	cc1       = "cc1"
	cc1Col1   = "cc1col1"
	cc1Col2   = "cc1col2"
	cc2       = "cc2"
	cc2Col1   = "cc2col1"

	mspID1       = "Org1MSP"
	peer1Org1URL = "peer1.org1.com:9999"
	peer2Org1URL = "peer2.org1.com:9999"

	mspID2       = "Org2MSP"
	peer1Org2URL = "peer1.org2.com:9999"
	peer2Org2URL = "peer2.org2.com:9999"

	mspID3       = "Org3MSP"
	peer1Org3URL = "peer1.org3.com:9999"
	peer2Org3URL = "peer2.org3.com:9999"
	peer3Org3URL = "peer3.org3.com:9999"
)

var (
	peer1Org1 = mocks.NewMockPeer("p11", peer1Org1URL)
	peer2Org1 = mocks.NewMockPeer("p12", peer2Org1URL)
	peer1Org2 = mocks.NewMockPeer("p21", peer1Org2URL)
	peer2Org2 = mocks.NewMockPeer("p22", peer2Org2URL)
	peer1Org3 = mocks.NewMockPeer("p31", peer1Org3URL)
	peer2Org3 = mocks.NewMockPeer("p32", peer2Org3URL)
	peer3Org3 = mocks.NewMockPeer("p33", peer3Org3URL)

	channelPeers = []fab.ChannelPeer{
		{NetworkPeer: newPeerConfig(peer1Org1URL, mspID1)},
		{NetworkPeer: newPeerConfig(peer2Org1URL, mspID1)},
		{NetworkPeer: newPeerConfig(peer1Org2URL, mspID2)},
		{NetworkPeer: newPeerConfig(peer2Org2URL, mspID2)},
		{NetworkPeer: newPeerConfig(peer1Org3URL, mspID3)},
		{NetworkPeer: newPeerConfig(peer2Org3URL, mspID3)},
	}

	peer1Org1Endpoint = &discmocks.MockDiscoveryPeerEndpoint{
		MSPID:        mspID1,
		Endpoint:     peer1Org1URL,
		LedgerHeight: 1000,
	}
	peer2Org1Endpoint = &discmocks.MockDiscoveryPeerEndpoint{
		MSPID:        mspID1,
		Endpoint:     peer2Org1URL,
		LedgerHeight: 1001,
	}
	peer1Org2Endpoint = &discmocks.MockDiscoveryPeerEndpoint{
		MSPID:        mspID2,
		Endpoint:     peer1Org2URL,
		LedgerHeight: 1002,
	}
	peer2Org2Endpoint = &discmocks.MockDiscoveryPeerEndpoint{
		MSPID:        mspID2,
		Endpoint:     peer2Org2URL,
		LedgerHeight: 1003,
	}
	peer1Org3Endpoint = &discmocks.MockDiscoveryPeerEndpoint{
		MSPID:        mspID3,
		Endpoint:     peer1Org3URL,
		LedgerHeight: 1004,
	}
	peer2Org3Endpoint = &discmocks.MockDiscoveryPeerEndpoint{
		MSPID:        mspID3,
		Endpoint:     peer2Org3URL,
		LedgerHeight: 1005,
	}
	//this peer is not part of EndpointConfig, but it should be treated equally and not ignored
	peer3Org3Endpoint = &discmocks.MockDiscoveryPeerEndpoint{
		MSPID:        mspID3,
		Endpoint:     peer3Org3URL,
		LedgerHeight: 1006,
	}

	cc1ChaincodeCall = &fab.ChaincodeCall{
		ID:          cc1,
		Collections: []string{cc1Col1, cc1Col2},
	}
	cc2ChaincodeCall = &fab.ChaincodeCall{
		ID:          cc2,
		Collections: []string{cc2Col1},
	}
)

func TestSelection(t *testing.T) {
	ctx := mocks.NewMockContext(mspmocks.NewMockSigningIdentity("test", mspID1))
	config := &config{
		EndpointConfig: mocks.NewMockEndpointConfig(),
		peers:          channelPeers,
	}
	ctx.SetEndpointConfig(config)

	discClient := discovery.NewMockDiscoveryClient()

	SetClientProvider(func(ctx contextAPI.Client) (DiscoveryClient, error) {
		return discClient, nil
	})

	var service *Service

	errHandler := func(ctxt fab.ClientContext, channelID string, err error) {
		derr, ok := errors.Cause(err).(DiscoveryError)
		if ok && derr.IsAccessDenied() {
			service.Close()
		}
	}

	service, err := New(
		ctx, channelID,
		mocks.NewMockDiscoveryService(nil, peer1Org1, peer2Org1, peer1Org2, peer2Org2, peer1Org3, peer2Org3, peer1Org3),
		WithRefreshInterval(5*time.Millisecond),
		WithResponseTimeout(100*time.Millisecond),
		WithErrorHandler(errHandler),
	)
	require.NoError(t, err)
	defer service.Close()

	t.Run("Error", func(t *testing.T) {
		// Error condition
		discClient.SetResponses(
			&discovery.MockDiscoverEndpointResponse{
				PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{},
				Error:         fmt.Errorf("simulated response error"),
			},
		)
		testSelectionError(t, service, "error getting channel response for channel [testchannel]: no successful response received from any peer: simulated response error")
	})

	t.Run("Transient error on one target", func(t *testing.T) {
		discClient.SetResponses(
			&discovery.MockDiscoverEndpointResponse{
				EndorsersErr: fmt.Errorf("no endorsement combination can be satisfied"), // Transient
			},
			&discovery.MockDiscoverEndpointResponse{
				Error: fmt.Errorf("some discovery error"), // Non-transient
			},
		)
		testSelectionError(t, service, "no endorsement combination can be satisfied")
	})

	t.Run("Success on one target", func(t *testing.T) {
		chaincodes := []*gossip.Chaincode{
			{
				Name:    cc1,
				Version: "v1",
			},
		}

		discClient.SetResponses(
			&discovery.MockDiscoverEndpointResponse{
				EndorsersErr: fmt.Errorf("no endorsement combination can be satisfied"),
			},
			&discovery.MockDiscoverEndpointResponse{
				Error: fmt.Errorf("some discovery error"),
			},
			&discovery.MockDiscoverEndpointResponse{
				PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{
					{
						MSPID:        mspID1,
						Endpoint:     peer1Org1URL,
						LedgerHeight: 10,
						Chaincodes:   chaincodes,
					},
				},
			},
		)
		endorsers, err := service.GetEndorsersForChaincode([]*fab.ChaincodeCall{{ID: cc1}})
		require.NoError(t, err)
		require.Len(t, endorsers, 1)

		endorser := endorsers[0]
		require.NotEmpty(t, endorser.Properties())
		require.Equal(t, uint64(10), endorser.Properties()[fab.PropertyLedgerHeight])
		require.Equal(t, false, endorser.Properties()[fab.PropertyLeftChannel])
		require.Equal(t, chaincodes, endorser.Properties()[fab.PropertyChaincodes])
	})

	t.Run("CCtoCC", func(t *testing.T) {
		discClient.SetResponses(
			&discovery.MockDiscoverEndpointResponse{
				PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{
					peer2Org1Endpoint, peer2Org3Endpoint, peer2Org2Endpoint,
					peer1Org1Endpoint, peer1Org2Endpoint, peer1Org3Endpoint,
				},
			},
		)

		// Wait for cache to refresh
		time.Sleep(20 * time.Millisecond)
		testSelectionCCtoCC(t, service)
	})

	t.Run("Peer Filter", func(t *testing.T) {
		endorsers, err := service.GetEndorsersForChaincode([]*fab.ChaincodeCall{{ID: cc1}},
			options.WithPeerFilter(func(peer fab.Peer) bool {
				return peer.(fab.PeerState).BlockHeight() > 1001
			}),
		)

		assert.NoError(t, err)
		assert.Equalf(t, 4, len(endorsers), "Expecting 4 endorser")

		// Ensure the endorsers all have a block height > 1001
		for _, endorser := range endorsers {
			blockHeight := endorser.(fab.PeerState).BlockHeight()
			assert.Truef(t, blockHeight > 1001, "Expecting block height to be > 1001")
		}
	})

	t.Run("Default Peer Filter", func(t *testing.T) {
		var prev fab.Peer
		for i := 0; i < 6; i++ {
			endorsers, err := service.GetEndorsersForChaincode([]*fab.ChaincodeCall{{ID: cc1}})
			assert.NoError(t, err)
			assert.Equalf(t, 6, len(endorsers), "Expecting 6 endorser")

			// Ensure that we get a different endorser as the first peer each time GetEndorsersForChaincode is called in
			// order to know that the default balancer (round-robin) is working.
			if prev != nil {
				require.NotEqual(t, prev, endorsers[0])
			}
			prev = endorsers[0]
		}
	})

	t.Run("Block Height Sorter Round Robin", func(t *testing.T) {
		testSelectionDistribution(t, service, balancer.RoundRobin(), 0)
	})

	t.Run("Block Height Sorter Random", func(t *testing.T) {
		testSelectionDistribution(t, service, balancer.Random(), 50)
	})

	t.Run("Priority Selector", func(t *testing.T) {
		testSelectionPrioritySelector(t, service)
	})

	serviceNoErrHandling, err := New(
		ctx, channelID,
		mocks.NewMockDiscoveryService(nil, peer1Org1, peer2Org1, peer1Org2, peer2Org2, peer1Org3, peer2Org3),
		WithRefreshInterval(5*time.Millisecond),
		WithResponseTimeout(100*time.Millisecond),
	)
	require.NoError(t, err)
	defer serviceNoErrHandling.Close()

	t.Run("Fatal Error, some error that returned from fab/discovery client ", func(t *testing.T) {
		discClient.SetResponses(
			&discovery.MockDiscoverEndpointResponse{
				PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{},
				Error:         fmt.Errorf("some err happened"),
			},
		)
		// Wait for cache to refresh
		time.Sleep(20 * time.Millisecond)
		testSelectionError(t, serviceNoErrHandling, "some err happened")
	})

	t.Run("Fatal Error Transient error", func(t *testing.T) {
		discClient.SetResponses(
			&discovery.MockDiscoverEndpointResponse{
				PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{
					{MSPID: "someMSPId"},
				},
			},
		)
		// Wait for cache to refresh
		time.Sleep(20 * time.Millisecond)
		endorsers, err := serviceNoErrHandling.GetEndorsersForChaincode([]*fab.ChaincodeCall{{ID: "notInstalledToAnyPeer"}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no endorsement combination can be satisfied")
		assert.Contains(t, err.Error(), "Discovery status Code: (11) UNKNOWN")
		assert.Equal(t, 0, len(endorsers))
	})

	t.Run("Fatal Error Transient error, cc not installed to peers", func(t *testing.T) {
		discClient.SetResponses(
			&discovery.MockDiscoverEndpointResponse{
				PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{
					{MSPID: "someMSPId"},
				},
			},
		)
		// Wait for cache to refresh
		time.Sleep(20 * time.Millisecond)
		endorsers, err := serviceNoErrHandling.GetEndorsersForChaincode([]*fab.ChaincodeCall{{ID: "notInstalledToAnyPeer"}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no endorsement combination can be satisfied")
		assert.Contains(t, err.Error(), "Discovery status Code: (11) UNKNOWN")
		assert.Equal(t, 0, len(endorsers))
	})

	t.Run("Fatal Error Access denied on all peers", func(t *testing.T) {
		discClient.SetResponses(
			&discovery.MockDiscoverEndpointResponse{
				PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{},
				EndorsersErr:  errors.New(AccessDenied),
			},
		)
		// Wait for cache to refresh
		time.Sleep(20 * time.Millisecond)
		endorsers, err := serviceNoErrHandling.GetEndorsersForChaincode([]*fab.ChaincodeCall{{ID: cc1}})
		require.Error(t, err)
		fmt.Println(err)
		assert.Contains(t, err.Error(), AccessDenied)
		assert.Equal(t, 0, len(endorsers))
	})

	t.Run("peer which was received from DS, but absent in EndpointConfig, isn't ignored by selection filter", func(t *testing.T) {
		svc, err := New(
			ctx, channelID,
			mocks.NewMockDiscoveryService(nil, peer3Org3),
			WithRefreshInterval(5*time.Millisecond),
			WithResponseTimeout(100*time.Millisecond),
		)
		require.NoError(t, err)
		defer serviceNoErrHandling.Close()

		discClient.SetResponses(
			&discovery.MockDiscoverEndpointResponse{
				PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{peer3Org3Endpoint},
			},
		)
		// Wait for cache to refresh
		time.Sleep(20 * time.Millisecond)
		endorsers, err := svc.GetEndorsersForChaincode([]*fab.ChaincodeCall{{ID: cc1}})
		require.NoError(t, err)
		assert.Len(t, endorsers, 1)
		assert.Equal(t, endorsers[0].URL(), peer3Org3Endpoint.Endpoint)
	})
}

func TestWithDiscoveryFilter(t *testing.T) {
	ctx := mocks.NewMockContext(mspmocks.NewMockSigningIdentity("test", mspID1))
	config := &config{
		EndpointConfig: mocks.NewMockEndpointConfig(),
		peers:          channelPeers,
	}
	ctx.SetEndpointConfig(config)

	discClient := discovery.NewMockDiscoveryClient()
	SetClientProvider(func(ctx contextAPI.Client) (DiscoveryClient, error) {
		return discClient, nil
	})

	discClient.SetResponses(
		&discovery.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{
				peer2Org1Endpoint, peer2Org3Endpoint, peer2Org2Endpoint,
				peer1Org1Endpoint, peer1Org2Endpoint, peer1Org3Endpoint,
			},
		},
	)

	t.Run("Error", func(t *testing.T) {
		expectedDiscoveryErrMsg := "simulated discovery service error"
		service, err := New(
			ctx, channelID,
			mocks.NewMockDiscoveryService(fmt.Errorf(expectedDiscoveryErrMsg)),
			WithRefreshInterval(500*time.Millisecond),
			WithResponseTimeout(2*time.Second),
		)
		require.NoError(t, err)
		defer service.Close()

		_, err = service.GetEndorsersForChaincode([]*fab.ChaincodeCall{cc1ChaincodeCall})
		assert.Truef(t, strings.Contains(err.Error(), expectedDiscoveryErrMsg), "expected error due to discovery error")
	})

	t.Run("Peers Down", func(t *testing.T) {
		service, err := New(
			ctx, channelID,
			mocks.NewMockDiscoveryService(nil, peer1Org1, peer2Org1, peer2Org2, peer2Org3),
			WithRefreshInterval(500*time.Millisecond),
			WithResponseTimeout(2*time.Second),
		)
		require.NoError(t, err)
		defer service.Close()

		endorsers, err := service.GetEndorsersForChaincode([]*fab.ChaincodeCall{cc1ChaincodeCall})
		assert.NoError(t, err)
		assert.Equalf(t, 4, len(endorsers), "Expecting 4 endorser")
	})

	t.Run("Peer Filter", func(t *testing.T) {
		service, err := New(
			ctx, channelID,
			mocks.NewMockDiscoveryService(nil, peer1Org1, peer2Org1, peer2Org2, peer2Org3),
			WithRefreshInterval(500*time.Millisecond),
			WithResponseTimeout(2*time.Second),
		)
		require.NoError(t, err)
		defer service.Close()

		endorsers, err := service.GetEndorsersForChaincode([]*fab.ChaincodeCall{cc1ChaincodeCall},
			options.WithPeerFilter(func(peer fab.Peer) bool {
				return peer.(fab.PeerState).BlockHeight() > 1001
			}))
		assert.NoError(t, err)
		assert.Equalf(t, 2, len(endorsers), "Expecting 2 endorser but got")
	})
}

func testSelectionError(t *testing.T, service *Service, expectedErrMsg string) {
	endorsers, err := service.GetEndorsersForChaincode([]*fab.ChaincodeCall{{ID: cc1}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), expectedErrMsg)
	assert.Equal(t, 0, len(endorsers))
}

func testSelectionCCtoCC(t *testing.T, service *Service) {
	endorsers, err := service.GetEndorsersForChaincode([]*fab.ChaincodeCall{cc1ChaincodeCall, cc2ChaincodeCall})
	assert.NoError(t, err)
	assert.Equalf(t, 6, len(endorsers), "Expecting 6 endorser")
}

func testSelectionDistribution(t *testing.T, service *Service, balancer balancer.Balancer, tolerance int) {
	iterations := 1000

	for threshold := 5; threshold >= 0; threshold-- {
		sorter := blockheightsorter.New(
			blockheightsorter.WithBlockHeightLagThreshold(threshold),
			blockheightsorter.WithBalancer(balancer),
		)

		expectedMin := iterations/(threshold+1) - tolerance
		count := make(map[string]int)

		for i := 0; i < iterations; i++ {
			endorsers, err := service.GetEndorsersForChaincode(
				[]*fab.ChaincodeCall{{ID: cc1}},
				options.WithPeerSorter(sorter),
			)

			assert.NoError(t, err)
			assert.Equalf(t, 6, len(endorsers), "Expecting 6 endorser")

			endorser := endorsers[0]
			count[endorser.URL()] = count[endorser.URL()] + 1
		}

		for url, c := range count {
			assert.Truef(t, c >= expectedMin, "Expecting peer [%s] to have been called at least %d times but got only %d", url, expectedMin, c)
		}
	}
}

func testSelectionPrioritySelector(t *testing.T, service *Service) {
	endorsers, err := service.GetEndorsersForChaincode([]*fab.ChaincodeCall{{ID: cc1}},
		options.WithPrioritySelector(func(peer1, peer2 fab.Peer) int {
			// Return peers in alphabetical order
			if peer1.URL() < peer2.URL() {
				return -1
			}
			if peer1.URL() > peer2.URL() {
				return 1
			}
			return 0
		}),
	)

	assert.NoError(t, err)
	assert.Equalf(t, 6, len(endorsers), "Expecting 6 endorser")

	var lastURL string
	for _, endorser := range endorsers {
		if lastURL != "" {
			assert.Truef(t, endorser.URL() <= lastURL, "Expecting endorsers in alphabetical order")
		}
		lastURL = endorser.URL()
	}
}

type config struct {
	fab.EndpointConfig
	peers []fab.ChannelPeer
}

func (c *config) ChannelPeers(name string) []fab.ChannelPeer {
	return c.peers
}

func (c *config) PeerConfig(nameOrURL string) (*fab.PeerConfig, bool) {
	for _, peer := range c.peers {
		if peer.URL == nameOrURL {
			return &peer.NetworkPeer.PeerConfig, true
		}
	}
	return nil, false
}

func newPeerConfig(url, mspID string) fab.NetworkPeer {
	return fab.NetworkPeer{
		PeerConfig: fab.PeerConfig{
			URL: url,
		},
		MSPID: mspID,
	}
}
