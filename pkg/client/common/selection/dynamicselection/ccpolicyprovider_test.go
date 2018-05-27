/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicselection

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	peer1 = mocks.NewMockPeer("p1", "peer1.example.com:9999")
	peer2 = mocks.NewMockPeer("p2", "peer2.example.com:9999")
)

func TestCCPolicyProvider(t *testing.T) {
	context := mocks.NewMockContext(
		mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
	)

	// All good
	ccPolicyProvider, err := newCCPolicyProvider(context, mocks.NewMockDiscoveryService(nil, peer1, peer2), "mychannel")
	require.NoErrorf(t, err, "Failed to setup cc policy provider")
	require.NotNilf(t, ccPolicyProvider, "Policy provider is nil")

	// Empty chaincode ID
	_, err = ccPolicyProvider.GetChaincodePolicy("")
	assert.Errorf(t, err, "Should have failed to retrieve chaincode policy for empty chaincode id")

	// Non-existent chaincode ID
	_, err = ccPolicyProvider.GetChaincodePolicy("abc")
	assert.Errorf(t, err, "Should have failed to retrieve non-existent cc policy")
}

func TestCCPolicyProviderNegative(t *testing.T) {
	context := mocks.NewMockContext(
		mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
	)

	// Invalid channelID
	ccPolicyProvider, err := newCCPolicyProvider(context, mocks.NewMockDiscoveryService(nil, peer1, peer2), "")
	require.Errorf(t, err, "Expected error for invalid channel ID")
	require.Nilf(t, ccPolicyProvider, "Expected policy provider to be nil")
}
