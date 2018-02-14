/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig/mocks"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	mock_fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient/mocks"
)

// TestNewPeerWithCertNoTLS tests that a peer can be constructed without using a cert
func TestNewPeerWithCertNoTLS(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mock_apiconfig.DefaultMockConfig(mockCtrl)

	url := "http://example.com"

	p, err := New(config, WithURL(url))

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

	config := mock_apiconfig.DefaultMockConfig(mockCtrl)

	url := "grpcs://0.0.0.0:1234"

	// TODO - test actual parameters and test server name override
	_, err := New(config, WithURL(url), WithTLSCert(mock_apiconfig.GoodCert))

	if err != nil {
		t.Fatalf("Expected peer to be constructed")
	}
}

// TestNewPeerWithCertBadParams tests that bad parameters causes an expected failure
func TestNewPeerWithCertBadParams(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mock_apiconfig.DefaultMockConfig(mockCtrl)

	_, err := New(config)

	if err == nil {
		t.Fatalf("Peer should not be constructed - bad params")
	}
}

// TestNewPeerTLSFromCertBad tests that bad parameters causes an expected failure
func TestNewPeerTLSFromCertBad(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mock_apiconfig.DefaultMockConfig(mockCtrl)

	url := "grpcs://0.0.0.0:1234"
	_, err := New(config, WithURL(url))

	if err == nil {
		t.Fatalf("Expected peer construction to fail")
	}
}

// TestEnrollmentCert tests the enrollment certificate getter/setters
func TestEnrollmentCert(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mock_apiconfig.DefaultMockConfig(mockCtrl)

	peer, err := New(config, WithURL(peer1URL))
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

	config := mock_apiconfig.DefaultMockConfig(mockCtrl)

	peer, err := New(config, WithURL(peer1URL))
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

	config := mock_apiconfig.DefaultMockConfig(mockCtrl)

	peer, err := New(config, WithURL(peer1URL))
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

	config := mock_apiconfig.DefaultMockConfig(mockCtrl)

	peer, err := New(config, WithURL(peer1URL))

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

func TestPeersToTxnProcessors(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mock_apiconfig.DefaultMockConfig(mockCtrl)

	peer1, err := New(config, WithURL(peer1URL))

	if err != nil {
		t.Fatalf("Failed to create NewPeer error(%v)", err)
	}

	peer2, err := New(config, WithURL(peer2URL))

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

func TestWithServerName(t *testing.T) {
	option := WithServerName("name")
	if option == nil {
		t.Fatalf("Failed to get option for server name.")
	}
	fmt.Printf("%v\n", &option)
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
	config := mock_apiconfig.DefaultMockConfig(mockCtrl)

	tlsConfig := apiconfig.TLSConfig{
		Path: "abc.com",
		Pem:  "",
	}
	peerConfig := apiconfig.PeerConfig{
		URL:         "abc.com",
		GRPCOptions: grpcOpts,
		TLSCACerts:  tlsConfig,
	}

	networkPeer := &apiconfig.NetworkPeer{
		PeerConfig: peerConfig,
		MspID:      "Org1MSP",
	}
	//from config with grpc
	_, err := New(config, FromPeerConfig(networkPeer))
	if err != nil {
		t.Fatalf("Failed to create new peer FromPeerConfig (%v)", err)
	}

	//with peer processor
	_, err = New(config, WithPeerProcessor(nil))
	if err == nil {
		t.Fatalf("Expected 'Failed to create new peer WithPeerProcessor ((target is required))")
	}

	//with peer processor
	_, err = New(config, WithServerName("server-name"))
	if err == nil {
		t.Fatalf("Expected 'Failed to create new peer WithServerName ((target is required))")
	}

}
