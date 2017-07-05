/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orderer

import (
	"fmt"
	"net"
	"testing"
	"time"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	client "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	mocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"

	"strings"

	"github.com/hyperledger/fabric/protos/common"
	ab "github.com/hyperledger/fabric/protos/orderer"
	"google.golang.org/grpc"
)

var testOrdererURL = "0.0.0.0:4584"
var testOrdererURL2 = "0.0.0.0:4585"
var testOrdererURL3 = "0.0.0.0:4586"
var testOrdererURL4 = "0.0.0.0:4587"
var testOrdererURL5 = "0.0.0.0:4588"
var testOrdererURL6 = "0.0.0.0:4590"

var validRootCA = `-----BEGIN CERTIFICATE-----
MIICYjCCAgmgAwIBAgIUB3CTDOU47sUC5K4kn/Caqnh114YwCgYIKoZIzj0EAwIw
fzELMAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNh
biBGcmFuY2lzY28xHzAdBgNVBAoTFkludGVybmV0IFdpZGdldHMsIEluYy4xDDAK
BgNVBAsTA1dXVzEUMBIGA1UEAxMLZXhhbXBsZS5jb20wHhcNMTYxMDEyMTkzMTAw
WhcNMjExMDExMTkzMTAwWjB/MQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZv
cm5pYTEWMBQGA1UEBxMNU2FuIEZyYW5jaXNjbzEfMB0GA1UEChMWSW50ZXJuZXQg
V2lkZ2V0cywgSW5jLjEMMAoGA1UECxMDV1dXMRQwEgYDVQQDEwtleGFtcGxlLmNv
bTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABKIH5b2JaSmqiQXHyqC+cmknICcF
i5AddVjsQizDV6uZ4v6s+PWiJyzfA/rTtMvYAPq/yeEHpBUB1j053mxnpMujYzBh
MA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBQXZ0I9
qp6CP8TFHZ9bw5nRtZxIEDAfBgNVHSMEGDAWgBQXZ0I9qp6CP8TFHZ9bw5nRtZxI
EDAKBggqhkjOPQQDAgNHADBEAiAHp5Rbp9Em1G/UmKn8WsCbqDfWecVbZPQj3RK4
oG5kQQIgQAe4OOKYhJdh3f7URaKfGTf492/nmRmtK+ySKjpHSrU=
-----END CERTIFICATE-----`

//
// Orderer via chain setOrderer/getOrderer
//
// Set the orderer URL through the chain setOrderer method. Verify that the
// orderer URL was set correctly through the getOrderer method. Repeat the
// process by updating the orderer URL to a different address.
//
func TestOrdererViaChain(t *testing.T) {
	client := client.NewClient(mocks.NewMockConfig())
	chain, err := client.NewChannel("testChain-orderer-member")
	if err != nil {
		t.Fatalf("error from NewChain %v", err)
	}
	orderer, _ := NewOrderer("localhost:7050", "", "", mocks.NewMockConfig())
	err = chain.AddOrderer(orderer)
	if err != nil {
		t.Fatalf("Error adding orderer: %v", err)
	}

	orderers := chain.Orderers()
	if orderers == nil || len(orderers) != 1 || orderers[0].URL() != "localhost:7050" {
		t.Fatalf("Failed to retieve the new orderer URL from the chain")
	}
	chain.RemoveOrderer(orderer)
	orderer2, err := NewOrderer("localhost:7054", "", "", mocks.NewMockConfig())
	if err != nil {
		t.Fatalf("Failed to create NewOrderer error(%v)", err)
	}
	err = chain.AddOrderer(orderer2)
	if err != nil {
		t.Fatalf("Error adding orderer: %v", err)
	}
	orderers = chain.Orderers()

	if orderers == nil || len(orderers) != 1 || orderers[0].URL() != "localhost:7054" {
		t.Fatalf("Failed to retieve the new orderer URL from the chain")
	}

}

//
// Orderer via chain missing orderer
//
// Attempt to send a request to the orderer with the sendTransaction method
// before the orderer URL was set. Verify that an error is reported when tying
// to send the request.
//
func TestPeerViaChainMissingOrderer(t *testing.T) {
	client := client.NewClient(mocks.NewMockConfig())
	chain, err := client.NewChannel("testChain-orderer-member2")
	if err != nil {
		t.Fatalf("error from NewChain %v", err)
	}
	_, err = chain.SendTransaction(nil)
	if err == nil {
		t.Fatalf("SendTransaction didn't return error")
	}
	if err.Error() != "orderers is nil" {
		t.Fatalf("SendTransaction didn't return right error")
	}

}

//
// Orderer via chain nil data
//
// Attempt to send a request to the orderer with the sendTransaction method
// with the data set to null. Verify that an error is reported when tying
// to send null data.
//
func TestOrdererViaChainNilData(t *testing.T) {
	client := client.NewClient(mocks.NewMockConfig())
	chain, err := client.NewChannel("testChain-orderer-member2")
	if err != nil {
		t.Fatalf("error from NewChain %v", err)
	}
	orderer, err := NewOrderer("localhost:7050", "", "", mocks.NewMockConfig())
	if err != nil {
		t.Fatalf("Failed to create NewOrderer error(%v)", err)
	}
	err = chain.AddOrderer(orderer)
	if err != nil {
		t.Fatalf("Error adding orderer: %v", err)
	}

	_, err = chain.SendTransaction(nil)
	if err == nil {
		t.Fatalf("SendTransaction didn't return error")
	}
	if err.Error() != "Transaction is nil" {
		t.Fatalf("SendTransaction didn't return right error")
	}
}

func TestSendDeliver(t *testing.T) {
	mockServer := startMockServer(t)
	orderer, _ := NewOrderer(testOrdererURL, "", "", mocks.NewMockConfig())
	// Test deliver happy path
	blocks, errors := orderer.SendDeliver(&fab.SignedEnvelope{})
	select {
	case block := <-blocks:
		if string(block.Data.Data[0]) != "test" {
			t.Fatalf("Expected test block got: %#v", block)
		}
	case err := <-errors:
		t.Fatalf("Unexpected error from SendDeliver(): %s", err)
	case <-time.After(time.Second * 5):
		t.Fatalf("Did not receive block or error from SendDeliver")
	}

	// Test deliver without valid envelope
	blocks, errors = orderer.SendDeliver(nil)
	select {
	case block := <-blocks:
		t.Fatalf("Expected error got block: %#v", block)
	case err := <-errors:
		if err == nil {
			t.Fatalf("Expected error with nil envelope")
		}
	case <-time.After(time.Second * 5):
		t.Fatalf("Did not receive block or error from SendDeliver")
	}

	// Test deliver with deliver error from OS
	testError := fmt.Errorf("test error")
	mockServer.DeliverError = testError
	blocks, errors = orderer.SendDeliver(&fab.SignedEnvelope{})
	select {
	case block := <-blocks:
		t.Fatalf("Expected error got block: %#v", block)
	case err := <-errors:
		if err == nil {
			t.Fatalf("Expected test error when OS Recv() fails, got: %s", err)
		}
	case <-time.After(time.Second * 5):
		t.Fatalf("Did not receive block or error from SendDeliver")
	}

	orderer, _ = NewOrderer(testOrdererURL+"invalid-test", "", "", mocks.NewMockConfig())
	// Test deliver happy path
	blocks, errors = orderer.SendDeliver(&fab.SignedEnvelope{})
	select {
	case block := <-blocks:
		t.Fatalf("This usecase was not supposed to receive blocks : %#v", block)
	case err := <-errors:
		fmt.Printf("There is an error as expected : %s \n", err)
	case <-time.After(time.Second * 5):
		t.Fatalf("Did not receive error from SendDeliver")
	}

}

func startMockServer(t *testing.T) *mocks.MockBroadcastServer {
	grpcServer := grpc.NewServer()
	lis, err := net.Listen("tcp", testOrdererURL)
	broadcastServer := new(mocks.MockBroadcastServer)
	ab.RegisterAtomicBroadcastServer(grpcServer, broadcastServer)
	if err != nil {
		fmt.Printf("Error starting test server %s", err)
		t.FailNow()
	}
	fmt.Printf("Starting test server\n")
	go grpcServer.Serve(lis)

	return broadcastServer
}

func startCustomizedMockServer(t *testing.T, serverURL string, grpcServer *grpc.Server, broadcastServer *mocks.MockBroadcastServer) *mocks.MockBroadcastServer {

	lis, err := net.Listen("tcp", serverURL)
	ab.RegisterAtomicBroadcastServer(grpcServer, broadcastServer)
	if err != nil {
		fmt.Printf("Error starting test server %s", err)
		t.FailNow()
	}
	fmt.Printf("Starting test customized server\n")
	go grpcServer.Serve(lis)

	return broadcastServer
}

func TestCreateNewOrdererWithRootCAs(t *testing.T) {

	rootCA := [][]byte{
		[]byte(validRootCA),
	}

	//Without TLS
	ordr, err := CreateNewOrdererWithRootCAs("", rootCA, "", mocks.NewMockConfig())
	if ordr == nil || err != nil {
		t.Fatalf("TestCreateNewOrdererWithRootCAs Failed, cause : [ %s ]", err)
	}

	//With TLS
	ordr, err = CreateNewOrdererWithRootCAs("", rootCA, "", mocks.NewMockConfigCustomized(true, false))
	if ordr == nil || err != nil {
		t.Fatalf("TestCreateNewOrdererWithRootCAs Failed, cause : [ %s ]", err)
	}

	//With TLS, With invalid rootCA
	ordr, err = CreateNewOrdererWithRootCAs("", [][]byte{}, "", mocks.NewMockConfigCustomized(true, true))
	if ordr != nil || err == nil {
		t.Fatal("TestCreateNewOrdererWithRootCAs Failed, was expecting error")
	}

}

func TestNewOrdererWithTLS(t *testing.T) {

	//Positive Test case
	orderer, err := NewOrderer("", "../../test/fixtures/tls/fabricca/ca/ca_root.pem", "", mocks.NewMockConfigCustomized(true, false))
	if orderer == nil || err != nil {
		t.Fatalf("Testing NewOrderer with TLS failed, cause [%s]", err)
	}

	//Negative Test case
	orderer, err = NewOrderer("", "", "", mocks.NewMockConfigCustomized(true, true))
	if orderer != nil || err == nil {
		t.Fatalf("Testing NewOrderer with TLS was supposed to failed")
	}
}

func TestSendBroadcast(t *testing.T) {

	//startMockServer(t)

	orderer, _ := NewOrderer(testOrdererURL, "", "", mocks.NewMockConfig())
	_, err := orderer.SendBroadcast(&fab.SignedEnvelope{})

	if err != nil {
		t.Fatalf("Test SendBroadcast was not supposed to fail")
	}

	orderer, _ = NewOrderer(testOrdererURL+"Test", "", "", mocks.NewMockConfig())
	_, err = orderer.SendBroadcast(&fab.SignedEnvelope{})

	if err == nil || !strings.HasPrefix(err.Error(), "Error Create NewAtomicBroadcastClient rpc error") {
		t.Fatalf("Test SendBroadcast was supposed to fail with expected error, instead it fail with [%s] error", err)
	}

}

func TestSendDeliverServerBadResponse(t *testing.T) {

	grpcServer := grpc.NewServer()
	broadcastServer := mocks.MockBroadcastServer{
		DeliverResponse: &ab.DeliverResponse{
			Type: &ab.DeliverResponse_Status{
				Status: common.Status_BAD_REQUEST,
			},
		},
	}

	startCustomizedMockServer(t, testOrdererURL2, grpcServer, &broadcastServer)
	orderer, _ := NewOrderer(testOrdererURL2, "", "", mocks.NewMockConfig())

	blocks, errors := orderer.SendDeliver(&fab.SignedEnvelope{})

	select {
	case block := <-blocks:
		t.Fatalf("This usecase was not supposed to receive blocks : %#v", block)
	case err := <-errors:
		if err.Error() != "Got error status from ordering service: BAD_REQUEST" {
			t.Fatalf("Ordering service error is not received as expected, %s", err)
		}
	case <-time.After(time.Second * 5):
		t.Fatalf("Did not receive error from SendDeliver")
	}
}

func TestSendDeliverServerSuccessResponse(t *testing.T) {

	grpcServer := grpc.NewServer()
	broadcastServer := mocks.MockBroadcastServer{
		DeliverResponse: &ab.DeliverResponse{
			Type: &ab.DeliverResponse_Status{
				Status: common.Status_SUCCESS,
			},
		},
	}

	startCustomizedMockServer(t, testOrdererURL3, grpcServer, &broadcastServer)
	orderer, _ := NewOrderer(testOrdererURL3, "", "", mocks.NewMockConfig())

	blocks, errors := orderer.SendDeliver(&fab.SignedEnvelope{})

	select {
	case block := <-blocks:
		if block != nil {
			t.Fatalf("This usecase was not supposed to get valid block")
		}
	case err := <-errors:
		t.Fatalf("This usecase was not supposed to get error : %s ", err.Error())
	case <-time.After(time.Second * 5):
		t.Fatalf("Did not receive block from SendDeliver")
	}
}

func TestSendDeliverFailure(t *testing.T) {

	grpcServer := grpc.NewServer()
	broadcastServer := mocks.MockBroadcastServer{
		DeliverResponse: &ab.DeliverResponse{},
	}

	startCustomizedMockServer(t, testOrdererURL6, grpcServer, &broadcastServer)
	orderer, _ := NewOrderer(testOrdererURL6, "", "", mocks.NewMockConfig())

	blocks, errors := orderer.SendDeliver(&fab.SignedEnvelope{})

	select {
	case block := <-blocks:
		t.Fatalf("This usecase was not supposed to get valid block %v", block)
	case err := <-errors:
		if err == nil || !strings.HasPrefix(err.Error(), "Received unknown response from ordering service") {
			t.Fatalf("Error response is not working as expected : '%s' ", err.Error())
		}
	case <-time.After(time.Second * 5):
		t.Fatal("Did not receive any response or error from SendDeliver")
	}
}

func TestSendBroadcastServerBadResponse(t *testing.T) {

	grpcServer := grpc.NewServer()
	broadcastServer := mocks.MockBroadcastServer{
		BroadcastInternalServerError: true,
	}

	startCustomizedMockServer(t, testOrdererURL4, grpcServer, &broadcastServer)
	orderer, _ := NewOrderer(testOrdererURL4, "", "", mocks.NewMockConfig())

	status, err := orderer.SendBroadcast(&fab.SignedEnvelope{})

	if status.String() != "INTERNAL_SERVER_ERROR" {
		t.Fatalf("Expected internal server error, but got %v", status)
	}
	if err == nil || err.Error() != "broadcast response is not success : INTERNAL_SERVER_ERROR" {
		t.Fatalf("Expected internal server error, but got %s", err)
	}

}

func TestSendBroadcastError(t *testing.T) {

	grpcServer := grpc.NewServer()
	broadcastServer := mocks.MockBroadcastServer{
		BroadcastError: fmt.Errorf("just to test error scenario"),
	}

	startCustomizedMockServer(t, testOrdererURL5, grpcServer, &broadcastServer)
	orderer, _ := NewOrderer(testOrdererURL5, "", "", mocks.NewMockConfig())

	status, err := orderer.SendBroadcast(&fab.SignedEnvelope{})

	if err == nil || status != nil {
		t.Fatalf("expected Send Broadcast to fail with error, but got %s", err)
	}

}
