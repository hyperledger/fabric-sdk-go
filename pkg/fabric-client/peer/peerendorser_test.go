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
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig/mocks"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	pb "github.com/hyperledger/fabric/protos/peer"
)

const (
	ecertPath   = "../../../test/fixtures/tls/fabricca/client/client_client1.pem"
	peer1URL    = "localhost:7050"
	peer2URL    = "localhost:7054"
	peerURLBad  = "localhost:9999"
	testAddress = "0.0.0.0:0"
)

// TestNewPeerEndorserTLS validates that a client configured with TLS
// creates the correct dial options.
func TestNewPeerEndorserTLS(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)

	url := "0.0.0.0:1234"
	certPool := x509.NewCertPool()

	config.EXPECT().IsTLSEnabled().Return(true)
	config.EXPECT().TLSCACertPool("cert").Return(certPool, nil)
	config.EXPECT().TLSCACertPool("").Return(certPool, nil)
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5)

	conn, err := newPeerEndorser(url, "cert", "", true, config)
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
	config := mock_apiconfig.NewMockConfig(mockCtrl)

	url := "0.0.0.0:1234"
	certPool := x509.NewCertPool()

	config.EXPECT().IsTLSEnabled().Return(true)
	config.EXPECT().TLSCACertPool("").Return(certPool, nil)
	config.EXPECT().TLSCACertPool("cert").Return(certPool, fmt.Errorf("ohoh"))
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5)

	_, err := newPeerEndorser(url, "cert", "", true, config)
	if err == nil {
		t.Fatalf("Peer conn construction should have failed")
	}
}

// TestNewPeerEndorserNoTLS validates that a client configured without TLS
// creates the correct dial options.
func TestNewPeerEndorserNoTLS(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)

	url := "0.0.0.0:1234"
	config.EXPECT().IsTLSEnabled().Return(false)
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5)

	conn, err := newPeerEndorser(url, "", "", true, config)
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
	config := mock_apiconfig.NewMockConfig(mockCtrl)

	url := "0.0.0.0:1234"
	config.EXPECT().IsTLSEnabled().Return(false)
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5)

	conn, err := newPeerEndorser(url, "", "", true, config)
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
	config := mock_apiconfig.NewMockConfig(mockCtrl)

	url := "0.0.0.0:1234"
	config.EXPECT().IsTLSEnabled().Return(false)
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5)

	conn, err := newPeerEndorser(url, "", "", false, config)
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
	config := mock_apiconfig.NewMockConfig(mockCtrl)

	url := ""
	_, err := newPeerEndorser(url, "", "", true, config)
	if err == nil {
		t.Fatalf("Peer conn should not be constructed - bad params")
	}
}

// TestNewPeerEndorserTLSBad validates that a client configured without
// the cert pool fails
func TestNewPeerEndorserTLSBad(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)

	url := "0.0.0.0:1234"
	config.EXPECT().IsTLSEnabled().Return(true)
	config.EXPECT().TLSCACertPool("").Return(x509.NewCertPool(), nil)
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5)

	_, err := newPeerEndorser(url, "", "", true, config)
	if err == nil {
		t.Fatalf("Peer conn should not be constructed - bad cert pool")
	}
}

// TestProcessProposalBadDial validates that a down
// endorser fails gracefully.
func TestProcessProposalBadDial(t *testing.T) {
	_, err := testProcessProposal(t, testAddress)
	if err == nil {
		t.Fatalf("Process proposal should have failed")
	}
}

// TestProcessProposalGoodDial validates that an up
// endorser connects.
func TestProcessProposalGoodDial(t *testing.T) {
	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	_, addr := startEndorserServer(t, grpcServer)

	_, err := testProcessProposal(t, addr)
	if err != nil {
		t.Fatalf("Process proposal failed (%v)", err)
	}
}

func testProcessProposal(t *testing.T, url string) (apitxn.TransactionProposalResult, error) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)

	config.EXPECT().IsTLSEnabled().Return(false)
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5)

	conn, err := newPeerEndorser(url, "", "", true, config)
	if err != nil {
		t.Fatalf("Peer conn construction error (%v)", err)
	}

	return conn.ProcessTransactionProposal(mockTransactionProposal())
}

func mockTransactionProposal() apitxn.TransactionProposal {
	return apitxn.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}
}

func startEndorserServer(t *testing.T, grpcServer *grpc.Server) (*mocks.MockEndorserServer, string) {
	lis, err := net.Listen("tcp", testAddress)
	addr := lis.Addr().String()

	endorserServer := &mocks.MockEndorserServer{}
	pb.RegisterEndorserServer(grpcServer, endorserServer)
	if err != nil {
		fmt.Printf("Error starting test server %s", err)
		t.FailNow()
	}
	fmt.Printf("Starting test server on %s\n", addr)
	go grpcServer.Serve(lis)
	return endorserServer, addr
}

func TestTransactionProposalError(t *testing.T) {
	var mockError error

	proposal := mockTransactionProposal()
	mockError = TransactionProposalError{
		Endorser: "1.2.3.4:1000",
		Proposal: mockTransactionProposal(),
		Err:      fmt.Errorf("error"),
	}

	errText := mockError.Error()
	mockText := fmt.Sprintf("Transaction processor (1.2.3.4:1000) returned error 'error' for proposal: %v", proposal)

	if errText != mockText {
		t.Fatalf("Unexpected error")
	}
}
