/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orderer

import (
	"net"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	ab "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	client "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	mocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

var testOrdererURL = "127.0.0.1:0"

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
	if err.Error() != "transaction is nil" {
		t.Fatalf("SendTransaction didn't return right error")
	}
}

func TestSendDeliver(t *testing.T) {
	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	mockServer, addr := startMockServer(t, grpcServer)

	orderer, _ := NewOrderer(addr, "", "", mocks.NewMockConfig())
	// Test deliver happy path
	blocks, errs := orderer.SendDeliver(&fab.SignedEnvelope{})
	select {
	case block := <-blocks:
		if string(block.Data.Data[0]) != "test" {
			t.Fatalf("Expected test block got: %#v", block)
		}
	case err := <-errs:
		t.Fatalf("Unexpected error from SendDeliver(): %s", err)
	case <-time.After(time.Second * 5):
		t.Fatalf("Did not receive block or error from SendDeliver")
	}

	// Test deliver without valid envelope
	blocks, errs = orderer.SendDeliver(nil)
	select {
	case block := <-blocks:
		t.Fatalf("Expected error got block: %#v", block)
	case err := <-errs:
		if err == nil {
			t.Fatalf("Expected error with nil envelope")
		}
	case <-time.After(time.Second * 5):
		t.Fatalf("Did not receive block or error from SendDeliver")
	}

	// Test deliver with deliver error from OS
	testError := errors.New("test error")
	mockServer.DeliverError = testError
	blocks, errs = orderer.SendDeliver(&fab.SignedEnvelope{})
	select {
	case block := <-blocks:
		t.Fatalf("Expected error got block: %#v", block)
	case err := <-errs:
		if err == nil {
			t.Fatalf("Expected test error when OS Recv() fails, got: %s", err)
		}
	case <-time.After(time.Second * 5):
		t.Fatalf("Did not receive block or error from SendDeliver")
	}

	orderer, _ = NewOrderer(testOrdererURL+"invalid-test", "", "", mocks.NewMockConfig())
	// Test deliver happy path
	blocks, errs = orderer.SendDeliver(&fab.SignedEnvelope{})
	select {
	case block := <-blocks:
		t.Fatalf("This usecase was not supposed to receive blocks : %#v", block)
	case err := <-errs:
		t.Logf("There is an error as expected : %s", err)
	case <-time.After(time.Second * 5):
		t.Fatalf("Did not receive error from SendDeliver")
	}

}

func startMockServer(t *testing.T, grpcServer *grpc.Server) (*mocks.MockBroadcastServer, string) {
	lis, err := net.Listen("tcp", testOrdererURL)
	addr := lis.Addr().String()

	broadcastServer := new(mocks.MockBroadcastServer)
	ab.RegisterAtomicBroadcastServer(grpcServer, broadcastServer)
	if err != nil {
		t.Logf("Error starting test server %s", err)
		t.FailNow()
	}
	t.Logf("Starting test server on %s", addr)
	go grpcServer.Serve(lis)

	return broadcastServer, addr
}

func startCustomizedMockServer(t *testing.T, serverURL string, grpcServer *grpc.Server, broadcastServer *mocks.MockBroadcastServer) string {
	lis, err := net.Listen("tcp", serverURL)
	addr := lis.Addr().String()

	ab.RegisterAtomicBroadcastServer(grpcServer, broadcastServer)
	if err != nil {
		t.Logf("Error starting test server %s", err)
		t.FailNow()
	}
	t.Logf("Starting test customized server on %s\n", addr)
	go grpcServer.Serve(lis)

	return addr
}

func TestNewOrdererWithTLS(t *testing.T) {
	//Positive Test case
	orderer, err := NewOrderer("grpcs://", "../../test/fixtures/fabricca/tls/ca/ca_root.pem", "", mocks.NewMockConfigCustomized(true, false, false))
	if orderer == nil || err != nil {
		t.Fatalf("Testing NewOrderer with TLS failed, cause [%s]", err)
	}

	//Negative Test case
	orderer, err = NewOrderer("grpcs://", "", "", mocks.NewMockConfigCustomized(true, false, true))
	if orderer != nil || err == nil {
		t.Fatalf("Testing NewOrderer with TLS was supposed to fail")
	}
}

func TestNewOrdererWithMutualTLS(t *testing.T) {
	//Positive Test case
	orderer, err := NewOrderer("grpcs://", "../../test/fixtures/tls/fabricca/ca/ca_root.pem", "", mocks.NewMockConfigCustomized(true, true, false))
	if orderer == nil || err != nil {
		t.Fatalf("Testing NewOrderer with Mutual TLS failed, cause [%s]", err)
	}
	//Negative Test case
	orderer, err = NewOrderer("grpcs://", "../../test/fixtures/tls/fabricca/ca/ca_root.pem", "", mocks.NewMockConfigCustomized(true, false, false))
	if orderer == nil || err != nil {
		t.Fatalf("Testing NewOrderer with Mutual TLS failed, cause [%s]", err)
	}
}

func TestSendBroadcast(t *testing.T) {
	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	_, addr := startMockServer(t, grpcServer)

	orderer, _ := NewOrderer(addr, "", "", mocks.NewMockConfig())
	_, err := orderer.SendBroadcast(&fab.SignedEnvelope{})

	if err != nil {
		t.Fatalf("Test SendBroadcast was not supposed to fail")
	}

	orderer, _ = NewOrderer(testOrdererURL+"Test", "", "", mocks.NewMockConfig())
	_, err = orderer.SendBroadcast(&fab.SignedEnvelope{})

	if err == nil || !strings.HasPrefix(err.Error(), "NewAtomicBroadcastClient") {
		t.Fatalf("Test SendBroadcast was supposed to fail with expected error, instead it fail with [%s] error", err)
	}

}

func TestSendDeliverServerBadResponse(t *testing.T) {

	broadcastServer := mocks.MockBroadcastServer{
		DeliverResponse: &ab.DeliverResponse{
			Type: &ab.DeliverResponse_Status{
				Status: common.Status_BAD_REQUEST,
			},
		},
	}

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	addr := startCustomizedMockServer(t, testOrdererURL, grpcServer, &broadcastServer)
	orderer, _ := NewOrderer(addr, "", "", mocks.NewMockConfig())

	blocks, errors := orderer.SendDeliver(&fab.SignedEnvelope{})

	select {
	case block := <-blocks:
		t.Fatalf("This usecase was not supposed to receive blocks : %#v", block)
	case err := <-errors:
		if err.Error() != "error status from ordering service BAD_REQUEST" {
			t.Fatalf("Ordering service error is not received as expected, %s", err)
		}
	case <-time.After(time.Second * 5):
		t.Fatalf("Did not receive error from SendDeliver")
	}
}

func TestSendDeliverServerSuccessResponse(t *testing.T) {

	broadcastServer := mocks.MockBroadcastServer{
		DeliverResponse: &ab.DeliverResponse{
			Type: &ab.DeliverResponse_Status{
				Status: common.Status_SUCCESS,
			},
		},
	}

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	addr := startCustomizedMockServer(t, testOrdererURL, grpcServer, &broadcastServer)

	orderer, _ := NewOrderer(addr, "", "", mocks.NewMockConfig())

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

	broadcastServer := mocks.MockBroadcastServer{
		DeliverResponse: &ab.DeliverResponse{},
	}

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	addr := startCustomizedMockServer(t, testOrdererURL, grpcServer, &broadcastServer)
	orderer, _ := NewOrderer(addr, "", "", mocks.NewMockConfig())

	blocks, errors := orderer.SendDeliver(&fab.SignedEnvelope{})

	select {
	case block := <-blocks:
		t.Fatalf("This usecase was not supposed to get valid block %v", block)
	case err := <-errors:
		if err == nil || !strings.HasPrefix(err.Error(), "unknown response from ordering service") {
			t.Fatalf("Error response is not working as expected : '%s' ", err.Error())
		}
	case <-time.After(time.Second * 5):
		t.Fatal("Did not receive any response or error from SendDeliver")
	}
}

func TestSendBroadcastServerBadResponse(t *testing.T) {

	broadcastServer := mocks.MockBroadcastServer{
		BroadcastInternalServerError: true,
	}

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	addr := startCustomizedMockServer(t, testOrdererURL, grpcServer, &broadcastServer)
	orderer, _ := NewOrderer(addr, "", "", mocks.NewMockConfig())

	status, err := orderer.SendBroadcast(&fab.SignedEnvelope{})

	if err == nil || err.Error() != "broadcast response is not success INTERNAL_SERVER_ERROR" {
		t.Fatalf("Expected internal server error, but got %s", err)
	}
	if status.String() != "INTERNAL_SERVER_ERROR" {
		t.Fatalf("Expected internal server error, but got %v", status)
	}
}

func TestSendBroadcastError(t *testing.T) {

	broadcastServer := mocks.MockBroadcastServer{
		BroadcastError: errors.New("just to test error scenario"),
	}

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	addr := startCustomizedMockServer(t, testOrdererURL, grpcServer, &broadcastServer)
	orderer, _ := NewOrderer(addr, "", "", mocks.NewMockConfig())

	status, err := orderer.SendBroadcast(&fab.SignedEnvelope{})

	if err == nil || status != nil {
		t.Fatalf("expected Send Broadcast to fail with error, but got %s", err)
	}

}

func TestInterfaces(t *testing.T) {
	var apiOrderer fab.Orderer
	var orderer Orderer

	apiOrderer = &orderer
	if apiOrderer == nil {
		t.Fatalf("this shouldn't happen.")
	}
}
