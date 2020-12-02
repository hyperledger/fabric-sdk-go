/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	reqContext "context"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-protos-go/gossip"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/test/mockfab"
	"github.com/stretchr/testify/require"
)

const (
	normalTimeout = 5 * time.Second
)

// TestNewPeerWithCertNoTLS tests that a peer can be constructed without using a cert
func TestNewPeerWithCertNoTLS(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mockfab.DefaultMockConfig(mockCtrl)

	url := "http://example.com"

	p, err := New(config, WithURL(url))

	if err != nil {
		t.Fatal("Expected peer to be constructed")
	}

	if p.URL() != url {
		t.Fatal("Unexpected peer URL")
	}
}

// TestNewPeerTLSFromCert tests that a peer can be constructed using a cert
func TestNewPeerTLSFromCert(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockfab.DefaultMockConfig(mockCtrl)

	url := "grpcs://0.0.0.0:1234"

	// TODO - test actual parameters and test server name override
	_, err := New(config, WithURL(url), WithTLSCert(mockfab.GoodCert))

	if err != nil {
		t.Fatal("Expected peer to be constructed")
	}
}

// TestNewPeerWithCertBadParams tests that bad parameters causes an expected failure
func TestNewPeerWithCertBadParams(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockfab.DefaultMockConfig(mockCtrl)

	_, err := New(config)

	if err == nil {
		t.Fatal("Peer should not be constructed - bad params")
	}
}

// TestNewPeerTLSFromCertBad tests that bad parameters causes an expected failure
func TestNewPeerTLSFromCertBad(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockfab.BadTLSClientMockConfig(mockCtrl)

	url := "grpcs://0.0.0.0:1234"
	_, err := New(config, WithURL(url))

	if err == nil {
		t.Fatal("Expected peer construction to fail")
	}
}

func TestMSPIDs(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockfab.DefaultMockConfig(mockCtrl)

	testMSP := "orgN"
	peer, err := New(config, WithURL(peer1URL), WithMSPID(testMSP))

	if err != nil {
		t.Fatalf("Failed to create NewPeer error(%s)", err)
	}

	if peer.MSPID() != testMSP {
		t.Fatal("Unexpected peer msp id")
	}
}

// Test that peer is proxy for proposal processor interface
func TestProposalProcessorSendProposal(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	proc := mockfab.NewMockProposalProcessor(mockCtrl)

	tp := mockProcessProposalRequest()
	tpr := fab.TransactionProposalResponse{Endorser: "example.com", Status: 99, ProposalResponse: nil}

	proc.EXPECT().ProcessTransactionProposal(gomock.Any(), tp).Return(&tpr, nil)

	p := Peer{processor: proc}
	ctx, cancel := reqContext.WithTimeout(reqContext.Background(), normalTimeout)
	defer cancel()
	tpr1, err := p.ProcessTransactionProposal(ctx, tp)

	if err != nil || !reflect.DeepEqual(&tpr, tpr1) {
		t.Fatal("Peer didn't proxy proposal processing")
	}
}

func TestPeersToTxnProcessors(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockfab.DefaultMockConfig(mockCtrl)

	peer1, err := New(config, WithURL(peer1URL))

	if err != nil {
		t.Fatalf("Failed to create NewPeer error(%s)", err)
	}

	peer2, err := New(config, WithURL(peer2URL))

	if err != nil {
		t.Fatalf("Failed to create NewPeer error(%s)", err)
	}

	peers := []fab.Peer{peer1, peer2}
	processors := PeersToTxnProcessors(peers)

	for i := range peers {
		if !reflect.DeepEqual(peers[i], processors[i]) {
			t.Fatal("Peer to Processors mismatch")
		}
	}
}

func TestInterfaces(t *testing.T) {
	var apiPeer fab.Peer
	var peer Peer

	apiPeer = &peer
	if apiPeer == nil {
		t.Fatal("this shouldn't happen.")
	}
}

func TestWithServerName(t *testing.T) {
	option := WithServerName("name")
	if option == nil {
		t.Fatal("Failed to get option for server name.")
	}
}

func TestPeerOptions(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	grpcOpts := make(map[string]interface{})
	grpcOpts["fail-fast"] = true
	grpcOpts["keep-alive-time"] = 1 * time.Second
	grpcOpts["keep-alive-timeout"] = 2 * time.Second
	grpcOpts["keep-alive-permit"] = false
	grpcOpts["ssl-target-name-override"] = "mnq"
	grpcOpts["allow-insecure"] = true
	config := mockfab.DefaultMockConfig(mockCtrl)

	peerConfig := fab.PeerConfig{
		URL:         "abc.com",
		GRPCOptions: grpcOpts,
	}

	chaincodes := []*gossip.Chaincode{
		{
			Name:    "cc1",
			Version: "v1",
		},
	}

	networkPeer := &fab.NetworkPeer{
		PeerConfig: peerConfig,
		MSPID:      "Org1MSP",
		Properties: fab.Properties{
			fab.PropertyLedgerHeight: uint64(100),
			fab.PropertyLeftChannel:  true,
			fab.PropertyChaincodes:   chaincodes,
		},
	}
	//from config with grpc
	peer, err := New(config, FromPeerConfig(networkPeer))
	require.NoError(t, err)
	require.NotNil(t, peer)
	require.NotEmpty(t, peer.Properties())
	require.Equal(t, uint64(100), peer.Properties()[fab.PropertyLedgerHeight])
	require.Equal(t, true, peer.Properties()[fab.PropertyLeftChannel])
	require.Equal(t, chaincodes, peer.Properties()[fab.PropertyChaincodes])

	//with peer processor
	_, err = New(config, WithPeerProcessor(nil))
	if err == nil {
		t.Fatal("Expected 'Failed to create new peer WithPeerProcessor ((target is required))")
	}

	//with peer processor
	_, err = New(config, WithServerName("server-name"))
	if err == nil {
		t.Fatal("Expected 'Failed to create new peer WithServerName ((target is required))")
	}
}

// TestNewPeerSecured validates that insecure option
func TestNewPeerSecured(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockfab.DefaultMockConfig(mockCtrl)

	url := "grpc://0.0.0.0:1234"

	conn, err := New(config, WithURL(url), WithInsecure())
	if err != nil {
		t.Fatal("Peer conn should be constructed")
	}

	if !conn.inSecure {
		t.Fatal("Expected insecure to be true")
	}

}
