/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"google.golang.org/grpc"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig/mocks"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

const (
	ecertPath   = "../../../test/fixtures/fabricca/tls/certs/client/client_client1.pem"
	peer1URL    = "localhost:7050"
	peer2URL    = "localhost:7054"
	peerURLBad  = "localhost:9999"
	testAddress = "127.0.0.1:0"
)

// TestNewPeerEndorserTLS validates that a client configured with TLS
// creates the correct dial options.
func TestNewPeerEndorserTLS(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)

	url := "grpcs://0.0.0.0:1234"
	certPool := x509.NewCertPool()

	config.EXPECT().TLSCACertPool("cert").Return(certPool, nil)
	config.EXPECT().TLSCACertPool("").Return(certPool, nil)
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5)
	config.EXPECT().TLSClientCerts().Return([]tls.Certificate{}, nil)

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

func TestNewPeerEndorserMutualTLS(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)

	//mutualTLSCerts := apiconfig.MutualTLSConfig{
	//	Client: struct {
	//		KeyPem   string
	//		Keyfile  string
	//		CertPem  string
	//		Certfile string
	//	}{KeyPem: "", Keyfile: "../../../test/fixtures/config/mutual_tls/client_sdk_go-key.pem", CertPem: "", Certfile: "../../../test/fixtures/config/mutual_tls/client_sdk_go.pem"},
	//}

	url := "grpcs://0.0.0.0:1234"
	certPool := x509.NewCertPool()

	config.EXPECT().TLSCACertPool("cert").Return(certPool, nil)
	config.EXPECT().TLSCACertPool("").Return(certPool, nil)
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5)
	config.EXPECT().TLSClientCerts().Return([]tls.Certificate{}, nil)

	conn, err := newPeerEndorser(url, "cert", "", true, config)
	if err != nil {
		t.Fatalf("Peer conn should be constructed: %v", err)
	}

	optInsecure := reflect.ValueOf(grpc.WithInsecure())

	for _, opt := range conn.grpcDialOption {
		optr := reflect.ValueOf(opt)
		if optr.Pointer() == optInsecure.Pointer() {
			t.Fatalf("TLS enabled - insecure not allowed")
		}
	}
}

func TestNewPeerEndorserMutualTLSNoClientCerts(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)

	url := "grpcs://0.0.0.0:1234"
	certPool := x509.NewCertPool()

	config.EXPECT().TLSCACertPool("cert").Return(certPool, nil)
	config.EXPECT().TLSCACertPool("").Return(certPool, nil)
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5)
	config.EXPECT().TLSClientCerts().Return([]tls.Certificate{}, nil)

	_, err := newPeerEndorser(url, "cert", "", true, config)
	if err != nil {
		t.Fatalf("Peer conn should be constructed: %v", err)
	}
}

// TestNewPeerEndorserTLSBadPool validates that a client configured with TLS
// with a bad cert pool fails gracefully.
func TestNewPeerEndorserTLSBadPool(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mock_apiconfig.NewMockConfig(mockCtrl)

	url := "grpcs://0.0.0.0:1234"
	certPool := x509.NewCertPool()

	config.EXPECT().TLSCACertPool("").Return(certPool, nil)
	config.EXPECT().TLSCACertPool("cert").Return(certPool, errors.New("ohoh"))
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

	url := "grpc://0.0.0.0:1234"
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

	url := "grpcs://0.0.0.0:1234"
	certPool := x509.NewCertPool()

	config.EXPECT().TLSCACertPool("").Return(certPool, nil)
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
		t.Logf("Error starting test server %s", err)
		t.FailNow()
	}
	t.Logf("Starting test server (endorser server in peerendorser_test) on %s", addr)
	go grpcServer.Serve(lis)
	return endorserServer, addr
}

func TestTransactionProposalError(t *testing.T) {
	var mockError error

	mockError = TransactionProposalError{
		Endorser: "1.2.3.4:1000",
		Proposal: mockTransactionProposal(),
		Err:      errors.New("error"),
	}

	errText := mockError.Error()
	mockText := "Transaction processor (1.2.3.4:1000) returned error 'error'"

	if !strings.Contains(errText, mockText) {
		t.Fatalf("Unexpected error")
	}
}
