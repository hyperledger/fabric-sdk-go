/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig/mocks"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/mocks"
)

// TestNewPeerWithCertNoTLS tests that a peer can be constructed without using a cert
func TestNewPeerWithCertNoTLS(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)
	config.EXPECT().IsTLSEnabled().Return(false)
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5)

	url := "http://example.com"
	p, err := NewPeer("http://example.com", config)

	if err != nil {
		t.Fatalf("Expected peer to be constructed")
	}

	if p.URL() != url {
		t.Fatalf("Unexpected peer URL")
	}
}

// TestNewPeerTLSFromCert tests that a peer can be constructed using a cert
func TestNewPeerTLSFromCert(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)

	certPool := x509.NewCertPool()
	config.EXPECT().IsTLSEnabled().Return(true)
	config.EXPECT().TLSCACertPool("cert").Return(certPool, nil)
	config.EXPECT().TLSCACertPool("").Return(certPool, nil)
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5)

	url := "0.0.0.0:1234"
	// TODO - test actual parameters and test server name override
	_, err := NewPeerTLSFromCert(url, "cert", "", config)

	if err != nil {
		t.Fatalf("Expected peer to be constructed")
	}
}

// TestNewPeerWithCertBadParams tests that bad parameters causes an expected failure
func TestNewPeerWithCertBadParams(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)

	_, err := NewPeer("", config)

	if err == nil {
		t.Fatalf("Peer should not be constructed - bad params")
	}
}

// TestNewPeerTLSFromCertBad tests that bad parameters causes an expected failure
func TestNewPeerTLSFromCertBad(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)

	config.EXPECT().IsTLSEnabled().Return(true)
	config.EXPECT().TLSCACertPool("").Return(x509.NewCertPool(), nil)
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5)

	url := "0.0.0.0:1234"
	_, err := NewPeerTLSFromCert(url, "", "", config)

	if err == nil {
		t.Fatalf("Expected peer construction to fail")
	}
}

// TestEnrollmentCert tests the enrollment certificate getter/setters
func TestEnrollmentCert(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)
	config.EXPECT().IsTLSEnabled().Return(false)
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
func TestRoles(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)
	config.EXPECT().IsTLSEnabled().Return(false)
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

func TestNames(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)
	config.EXPECT().IsTLSEnabled().Return(false)
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

func TestMSPIDs(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)
	config.EXPECT().IsTLSEnabled().Return(false)
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5)

	peer, err := NewPeer(peer1URL, config)
	if err != nil {
		t.Fatalf("Failed to create NewPeer error(%v)", err)
	}
	testMSP := "orgN"
	peer.SetMSPID(testMSP)

	if peer.MSPID() != testMSP {
		t.Fatalf("Unexpected peer msp id")
	}
}

// Test that peer is proxy for proposal processor interface
func TestProposalProcessorSendProposal(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	proc := mock_apitxn.NewMockProposalProcessor(mockCtrl)

	tp := mockTransactionProposal()
	tpr := apitxn.TransactionProposalResult{Endorser: "example.com", Status: 99, Proposal: tp, ProposalResponse: nil}

	proc.EXPECT().ProcessTransactionProposal(tp).Return(tpr, nil)

	p := Peer{processor: proc, name: "", roles: nil}
	tpr1, err := p.ProcessTransactionProposal(tp)

	if err != nil || !reflect.DeepEqual(tpr, tpr1) {
		t.Fatalf("Peer didn't proxy proposal processing")
	}
}

// TODO: Placeholder functions
func TestPlaceholders(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)
	config.EXPECT().IsTLSEnabled().Return(false)
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5)

	peer, err := NewPeer(peer1URL, config)
	if err != nil {
		t.Fatalf("Failed to create NewPeer error(%v)", err)
	}

	_, err = peer.AddListener("", nil, nil)
	if err != nil {
		t.Fatalf("Placeholder function has been implemented - update tests")
	}

	_, err = peer.RemoveListener("")
	if err != nil {
		t.Fatalf("Placeholder function has been implemented - update tests")
	}

	_, err = peer.IsEventListened("", nil)
	if err != nil {
		t.Fatalf("Placeholder function has been implemented - update tests")
	}
}

func TestPeersToTxnProcessors(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)
	config.EXPECT().IsTLSEnabled().Return(false)
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5)

	peer1, err := NewPeer(peer1URL, config)
	if err != nil {
		t.Fatalf("Failed to create NewPeer error(%v)", err)
	}

	config.EXPECT().IsTLSEnabled().Return(false)
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

func TestInterfaces(t *testing.T) {
	var apiPeer fab.Peer
	var peer Peer

	apiPeer = &peer
	if apiPeer == nil {
		t.Fatalf("this shouldn't happen.")
	}
}
