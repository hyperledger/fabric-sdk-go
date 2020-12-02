/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discovery

import (
	"testing"

	"github.com/hyperledger/fabric-protos-go/gossip"
	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	gprotoext "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/gossip/protoext"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery/mocks"
	"github.com/stretchr/testify/require"
)

func TestGetProperties(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		chaincodes := []*gossip.Chaincode{
			{
				Name:    "cc1",
				Version: "v1",
			},
		}

		endpoint := &discclient.Peer{
			StateInfoMessage: newStateInfoMessage(&mocks.MockDiscoveryPeerEndpoint{
				LedgerHeight: 1001,
				Chaincodes:   chaincodes,
				LeftChannel:  true,
			}),
		}

		properties := GetProperties(endpoint)
		require.NotEmpty(t, properties)
		require.Equal(t, uint64(1001), properties[fab.PropertyLedgerHeight])
		require.Equal(t, chaincodes, properties[fab.PropertyChaincodes])
		require.Equal(t, true, properties[fab.PropertyLeftChannel])
	})

	t.Run("Nil state info message", func(t *testing.T) {
		properties := GetProperties(&discclient.Peer{})
		require.Empty(t, properties)
	})

	t.Run("Nil properties in state info message", func(t *testing.T) {
		endpoint := &discclient.Peer{
			StateInfoMessage: &gprotoext.SignedGossipMessage{
				GossipMessage: &gossip.GossipMessage{
					Content: &gossip.GossipMessage_StateInfo{
						StateInfo: &gossip.StateInfo{},
					},
				},
			},
		}

		properties := GetProperties(endpoint)
		require.Empty(t, properties)
	})
}
