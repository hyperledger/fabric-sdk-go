/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"encoding/pem"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig/mocks"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	mock_fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

// TestNewPeerWithCertNoTLS tests that a peer can be constructed without using a cert
func TestDeprecatedNewPeerWithCertNoTLS(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	url := "http://example.com"
	config := mock_apiconfig.NewMockConfig(mockCtrl)
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5)

	p, err := NewPeer(url, config)

	if err != nil {
		t.Fatalf("Expected peer to be constructed")
	}

	if p.URL() != url {
		t.Fatalf("Unexpected peer URL")
	}
}

// TestNewPeerTLSFromCert tests that a peer can be constructed using a cert
func TestDeprecatedNewPeerTLSFromCert(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	tlsConfig := apiconfig.TLSConfig{Path: "../../../test/fixtures/fabricca/tls/ca/ca_root.pem"}
	cert, err := tlsConfig.TLSCert()

	if err != nil {
		t.Fatalf("Failed to load cert: %v", err)
	}

	config := mock_apiconfig.DefaultMockConfig(mockCtrl)
	config.EXPECT().TLSCACertPool(cert).Return(mock_apiconfig.CertPool, nil).AnyTimes()

	url := "grpcs://0.0.0.0:1234"

	// TODO - test actual parameters and test server name override
	_, err = NewPeerTLSFromCert(url, "../../../test/fixtures/fabricca/tls/ca/ca_root.pem", "", config)

	if err != nil {
		t.Fatalf("Expected peer to be constructed, failed with error: %v", err)
	}
}

// TestNewPeerWithCertBadParams tests that bad parameters causes an expected failure
func TestDeprecatedNewPeerWithCertBadParams(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)

	_, err := NewPeer("", config)

	if err == nil {
		t.Fatalf("Peer should not be constructed - bad params")
	}
}

// TestNewPeerTLSFromCertBad tests that bad parameters causes an expected failure
func TestDeprecatedNewPeerTLSFromCertBad(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.DefaultMockConfig(mockCtrl)

	url := "grpcs://0.0.0.0:1234"
	_, err := NewPeerTLSFromCert(url, "", "", config)

	if err == nil {
		t.Fatalf("Expected peer construction to fail")
	}
}

// TestEnrollmentCert tests the enrollment certificate getter/setters
func TestDeprecatedEnrollmentCert(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5)

	peer, err := NewPeer(peer1URL, config)
	if err != nil {
		t.Fatalf("Failed to create NewPeerWithCert error(%v)", err)
	}

	if peer.EnrollmentCertificate() != nil {
		t.Fatalf("Expected empty peer enrollment certs")
	}

	pemBuffer, err := ioutil.ReadFile(ecertPath)
	if err != nil {
		t.Fatalf("ecert fixture missing")
	}

	cert, _ := pem.Decode(pemBuffer)
	peer.SetEnrollmentCertificate(cert)

	fetchedCert := peer.EnrollmentCertificate()
	if !reflect.DeepEqual(cert, fetchedCert) {
		t.Fatalf("Enrollment certificate mismatch")
	}
}

// TestRoles tests the roles certificate getter/setters
func TestDeprecatedRoles(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5)

	peer, err := NewPeer(peer1URL, config)
	if err != nil {
		t.Fatalf("Failed to create NewPeer error(%v)", err)
	}

	if len(peer.Roles()) != 0 {
		t.Fatalf("Expected empty peer roles")
	}

	roles := []string{"role1", "role2"}
	peer.SetRoles(roles)

	fetchedRoles := peer.Roles()
	if !reflect.DeepEqual(roles, fetchedRoles) {
		t.Fatalf("Unexpected roles")
	}
}

// TestRoles tests the name certificate getter/setters

func TestDeprecatedNames(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5)

	peer, err := NewPeer(peer1URL, config)
	if err != nil {
		t.Fatalf("Failed to create NewPeer error(%v)", err)
	}

	if peer.Name() != "" {
		t.Fatalf("Expected empty peer name")
	}

	const peerName = "I am Peer"
	peer.SetName(peerName)

	if peer.Name() != peerName {
		t.Fatalf("Unexpected peer name")
	}
}

// Test that peer is proxy for proposal processor interface
func TestDeprecatedProposalProcessorSendProposal(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	proc := mock_fab.NewMockProposalProcessor(mockCtrl)

	tp := mockTransactionProposal()
	tpr := fab.TransactionProposalResponse{Endorser: "example.com", Status: 99, Proposal: tp, ProposalResponse: nil}

	proc.EXPECT().ProcessTransactionProposal(tp).Return(tpr, nil)

	p := Peer{processor: proc, name: "", roles: nil}
	tpr1, err := p.ProcessTransactionProposal(tp)

	if err != nil || !reflect.DeepEqual(tpr, tpr1) {
		t.Fatalf("Peer didn't proxy proposal processing")
	}
}

func TestDeprecatedPeersToTxnProcessors(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5)

	peer1, err := NewPeer(peer1URL, config)
	if err != nil {
		t.Fatalf("Failed to create NewPeer error(%v)", err)
	}

	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5)
	peer2, err := NewPeer(peer2URL, config)
	if err != nil {
		t.Fatalf("Failed to create NewPeer error(%v)", err)
	}

	peers := []fab.Peer{peer1, peer2}
	processors := PeersToTxnProcessors(peers)

	for i := range peers {
		if !reflect.DeepEqual(peers[i], processors[i]) {
			t.Fatalf("Peer to Processors mismatch")
		}
	}
}

func TestDeprecatedInterfaces(t *testing.T) {
	var apiPeer fab.Peer
	var peer Peer

	apiPeer = &peer
	if apiPeer == nil {
		t.Fatalf("this shouldn't happen.")
	}
}

func TestNewPeerFromConfig(t *testing.T) {

	grpcOpts := make(map[string]interface{})

	peerConfig := apiconfig.PeerConfig{
		URL:         "abc.com",
		GRPCOptions: grpcOpts,
	}

	networkPeer := &apiconfig.NetworkPeer{
		PeerConfig: peerConfig,
		MspID:      "Org1MSP",
	}
	_, err := NewPeerFromConfig(networkPeer, mocks.NewMockConfig())
	if err != nil {
		t.Fatalf("Failed to create new peer from config: %v", err)
	}

}
