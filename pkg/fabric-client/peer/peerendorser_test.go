/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"crypto/x509"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/api"
	"github.com/hyperledger/fabric-sdk-go/api/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	pb "github.com/hyperledger/fabric/protos/peer"
)

const (
	ecertPath   = "../../../test/fixtures/tls/fabricca/client/client_client1.pem"
	peer1URL    = "localhost:7050"
	peer2URL    = "localhost:7054"
	peerURLBad  = "localhost:9999"
	testAddress = "0.0.0.0:5244"
)

// TestNewPeerEndorserTLS validates that a client configured with TLS
// creates the correct dial options.
func TestNewPeerEndorserTLS(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_api.NewMockConfig(mockCtrl)

	url := "0.0.0.0:1234"
	certPool := x509.NewCertPool()

	config.EXPECT().IsTLSEnabled().Return(true)
	config.EXPECT().TLSCACertPool("cert").Return(certPool, nil)

	conn, err := newPeerEndorser(url, "cert", "", connTimeout, true, config)
	if err != nil {
		t.Fatalf("Peer conn should be constructed")
	}

	optInsecure := reflect.ValueOf(grpc.WithInsecure())

	for _, opt := range conn.grpcDialOption {
		optr := reflect.ValueOf(opt)
		if optr.Pointer() == optInsecure.Pointer() {
			t.Fatalf("TLS enabled - insecure not allowed")
		}
	}
}

// TestNewPeerEndorserTLSBadPool validates that a client configured with TLS
// with a bad cert pool fails gracefully.
func TestNewPeerEndorserTLSBadPool(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_api.NewMockConfig(mockCtrl)

	url := "0.0.0.0:1234"
	certPool := x509.NewCertPool()

	config.EXPECT().IsTLSEnabled().Return(true)
	config.EXPECT().TLSCACertPool("cert").Return(certPool, fmt.Errorf("ohoh"))

	_, err := newPeerEndorser(url, "cert", "", connTimeout, true, config)
	if err == nil {
		t.Fatalf("Peer conn construction should have failed")
	}
}

// TestNewPeerEndorserNoTLS validates that a client configured without TLS
// creates the correct dial options.
func TestNewPeerEndorserNoTLS(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_api.NewMockConfig(mockCtrl)

	url := "0.0.0.0:1234"
	config.EXPECT().IsTLSEnabled().Return(false)

	conn, err := newPeerEndorser(url, "", "", connTimeout, true, config)
	if err != nil {
		t.Fatalf("Peer conn should be constructed")
	}

	optInsecure := reflect.ValueOf(grpc.WithInsecure())
	optInsecureFound := false

	for _, opt := range conn.grpcDialOption {
		optr := reflect.ValueOf(opt)
		if optr.Pointer() == optInsecure.Pointer() {
			optInsecureFound = true
		}
	}

	if !optInsecureFound {
		t.Fatalf("Expected insecure to be found")
	}
}

// TestNewPeerEndorserBlocking validates that a client configured with blocking
// creates the correct dial options.
func TestNewPeerEndorserBlocking(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_api.NewMockConfig(mockCtrl)

	url := "0.0.0.0:1234"
	config.EXPECT().IsTLSEnabled().Return(false)

	conn, err := newPeerEndorser(url, "", "", connTimeout, true, config)
	if err != nil {
		t.Fatalf("Peer conn should be constructed")
	}

	optBlocking := reflect.ValueOf(grpc.WithBlock())
	optBlockingFound := false

	for _, opt := range conn.grpcDialOption {
		optr := reflect.ValueOf(opt)
		if optr.Pointer() == optBlocking.Pointer() {
			optBlockingFound = true
		}
	}

	if !optBlockingFound {
		t.Fatalf("Expected blocking to be found")
	}
}

// TestNewPeerEndorserNonBlocking validates that a client configured without blocking
// creates the correct dial options.
func TestNewPeerEndorserNonBlocking(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_api.NewMockConfig(mockCtrl)

	url := "0.0.0.0:1234"
	config.EXPECT().IsTLSEnabled().Return(false)

	conn, err := newPeerEndorser(url, "", "", connTimeout, false, config)
	if err != nil {
		t.Fatalf("Peer conn should be constructed")
	}

	optBlocking := reflect.ValueOf(grpc.WithBlock())

	for _, opt := range conn.grpcDialOption {
		optr := reflect.ValueOf(opt)
		if optr.Pointer() == optBlocking.Pointer() {
			t.Fatalf("Blocking opt found when not expected")
		}
	}
}

// TestNewPeerEndorserBadParams validates that a client configured without
// params fails
func TestNewPeerEndorserBadParams(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_api.NewMockConfig(mockCtrl)

	url := ""
	_, err := newPeerEndorser(url, "", "", connTimeout, true, config)
	if err == nil {
		t.Fatalf("Peer conn should not be constructed - bad params")
	}
}

// TestNewPeerEndorserTLSBad validates that a client configured without
// the cert pool fails
func TestNewPeerEndorserTLSBad(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_api.NewMockConfig(mockCtrl)

	url := "0.0.0.0:1234"
	config.EXPECT().IsTLSEnabled().Return(true)

	_, err := newPeerEndorser(url, "", "", connTimeout, true, config)
	if err == nil {
		t.Fatalf("Peer conn should not be constructed - bad cert pool")
	}
}

// TestProcessProposalBadDial validates that a down
// endorser fails gracefully.
func TestProcessProposalBadDial(t *testing.T) {
	_, err := testProcessProposal(t, time.Millisecond*10)
	if err == nil {
		t.Fatalf("Process proposal should have failed")
	}
}

// TestProcessProposalGoodDial validates that an up
// endorser connects.
func TestProcessProposalGoodDial(t *testing.T) {
	startEndorserServer(t)

	_, err := testProcessProposal(t, connTimeout)
	if err != nil {
		t.Fatalf("Process proposal failed (%v)", err)
	}
}

func testProcessProposal(t *testing.T, to time.Duration) (*api.TransactionProposalResponse, error) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_api.NewMockConfig(mockCtrl)

	url := testAddress
	config.EXPECT().IsTLSEnabled().Return(false)

	conn, err := newPeerEndorser(url, "", "", to, true, config)
	if err != nil {
		t.Fatalf("Peer conn construction error (%v)", err)
	}

	return conn.ProcessProposal(mockTransactionProposal())
}

func mockTransactionProposal() *api.TransactionProposal {
	return &api.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}
}

// TODO: this function is duplicated.
func startEndorserServer(t *testing.T) *mocks.MockEndorserServer {
	grpcServer := grpc.NewServer()
	lis, err := net.Listen("tcp", testAddress)
	endorserServer := &mocks.MockEndorserServer{}
	pb.RegisterEndorserServer(grpcServer, endorserServer)
	if err != nil {
		fmt.Printf("Error starting test server %s", err)
		t.FailNow()
	}
	fmt.Printf("Starting test server\n")
	go grpcServer.Serve(lis)
	return endorserServer
}
