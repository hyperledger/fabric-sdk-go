/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricselection

import (
	"context"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/stretchr/testify/require"
)

func TestPeerEndpointValue(t *testing.T) {
	const mspID = "org1"
	const url = "peer1.org1.com"

	t.Run("Success", func(t *testing.T) {
		ep := &peerEndpointValue{
			mspID: mspID,
			url:   url,
			properties: fab.Properties{
				fab.PropertyLedgerHeight: uint64(1001),
			},
		}

		require.Equal(t, mspID, ep.MSPID())
		require.Equal(t, url, ep.URL())
		require.NotEmpty(t, ep.Properties())
		require.Equal(t, uint64(1001), ep.Properties()[fab.PropertyLedgerHeight])
		require.Equal(t, uint64(1001), ep.BlockHeight())
		require.Panics(t, func() {
			_, err := ep.ProcessTransactionProposal(context.TODO(), fab.ProcessProposalRequest{})
			require.NoError(t, err)
		})
	})

	t.Run("No ledger height property", func(t *testing.T) {
		ep := &peerEndpointValue{
			mspID:      mspID,
			url:        url,
			properties: fab.Properties{},
		}

		require.Equal(t, mspID, ep.MSPID())
		require.Equal(t, url, ep.URL())
		require.Empty(t, ep.Properties())
		require.Equal(t, uint64(0), ep.BlockHeight())
	})
}
