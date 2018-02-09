/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package txnhandler

import (
	"errors"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/msp"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/chclient"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/stretchr/testify/assert"
)

func TestSignatureValidationHandlerSuccess(t *testing.T) {
	request := chclient.Request{ChaincodeID: "test", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, chclient.Opts{}, t)

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}

	// Add mock msp to msp manager
	msps := make(map[string]msp.MSP)
	msps["Org1MSP"] = fcmocks.NewMockMSP(nil)

	clientContext := setupContextForSignatureValidation(fcmocks.NewMockMSPManager(msps), []apifabclient.Peer{mockPeer1}, t)

	handler := NewQueryHandler()
	handler.Handle(requestContext, clientContext)
	assert.Nil(t, requestContext.Error)
}

func TestSignatureValidationMspErrors(t *testing.T) {

	// Sample request
	request := chclient.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}
	requestContext := prepareRequestContext(request, chclient.Opts{}, t)
	handler := NewQueryHandler()

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}

	// Test #1: GetMSPs error
	msps := make(map[string]msp.MSP)
	clientContext := setupContextForSignatureValidation(fcmocks.NewMockMSPManagerWithError(msps, errors.New("GetMSPs")), []apifabclient.Peer{mockPeer1}, t)
	handler.Handle(requestContext, clientContext)
	verifyExpectedError(requestContext, "GetMSPs return error", t)

	// Test #2: MPS manager has no mps
	clientContext = setupContextForSignatureValidation(fcmocks.NewMockMSPManager(nil), []apifabclient.Peer{mockPeer1}, t)
	handler.Handle(requestContext, clientContext)
	verifyExpectedError(requestContext, "is empty", t)

	// Test #3: MSP not found
	msps = make(map[string]msp.MSP)
	msps["NotOrg1MSP"] = fcmocks.NewMockMSP(nil)
	clientContext = setupContextForSignatureValidation(fcmocks.NewMockMSPManager(msps), []apifabclient.Peer{mockPeer1}, t)
	handler.Handle(requestContext, clientContext)
	verifyExpectedError(requestContext, "not found", t)
}

func TestSignatureValidationUnmarshallEndorserError(t *testing.T) {

	// Sample request
	request := chclient.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}
	requestContext := prepareRequestContext(request, chclient.Opts{}, t)
	handler := NewQueryHandler()

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}

	// Unmarshall endorser error
	msps := make(map[string]msp.MSP)
	msps["Org1MSP"] = fcmocks.NewMockMSP(nil)
	mockPeer1.Endorser = []byte("Invalid")
	clientContext := setupContextForSignatureValidation(fcmocks.NewMockMSPManager(msps), []apifabclient.Peer{mockPeer1}, t)
	handler.Handle(requestContext, clientContext)
	verifyExpectedError(requestContext, "Unmarshal endorser error", t)

}

func TestSignatureValidationDeserializeIdentityError(t *testing.T) {

	// Sample request
	request := chclient.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}
	requestContext := prepareRequestContext(request, chclient.Opts{}, t)
	handler := NewQueryHandler()

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}

	msps := make(map[string]msp.MSP)
	msps["Org1MSP"] = fcmocks.NewMockMSP(errors.New("DeserializeIdentity"))
	clientContext := setupContextForSignatureValidation(fcmocks.NewMockMSPManager(msps), []apifabclient.Peer{mockPeer1}, t)
	handler.Handle(requestContext, clientContext)
	verifyExpectedError(requestContext, "Failed to deserialize creator identity", t)
}

func TestSignatureValidationCreatorValidateError(t *testing.T) {

	// Sample request
	request := chclient.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}
	requestContext := prepareRequestContext(request, chclient.Opts{}, t)
	handler := NewQueryHandler()

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}

	msps := make(map[string]msp.MSP)
	msps["Org1MSP"] = fcmocks.NewMockMSP(errors.New("Validate"))
	clientContext := setupContextForSignatureValidation(fcmocks.NewMockMSPManager(msps), []apifabclient.Peer{mockPeer1}, t)
	handler.Handle(requestContext, clientContext)
	verifyExpectedError(requestContext, "The creator certificate is not valid", t)
}

func TestSignatureValidationCreatorVerifyError(t *testing.T) {

	// Sample request
	request := chclient.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}
	requestContext := prepareRequestContext(request, chclient.Opts{}, t)
	handler := NewQueryHandler()

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}

	msps := make(map[string]msp.MSP)
	msps["Org1MSP"] = fcmocks.NewMockMSP(errors.New("Verify"))
	clientContext := setupContextForSignatureValidation(fcmocks.NewMockMSPManager(msps), []apifabclient.Peer{mockPeer1}, t)
	handler.Handle(requestContext, clientContext)
	verifyExpectedError(requestContext, "The creator's signature over the proposal is not valid", t)
}

func verifyExpectedError(requestContext *chclient.RequestContext, expected string, t *testing.T) {
	assert.NotNil(t, requestContext.Error)
	if requestContext.Error == nil || !strings.Contains(requestContext.Error.Error(), expected) {
		t.Fatal("Expected error: ", expected, ", Received error:", requestContext.Error)
	}
}

func setupContextForSignatureValidation(mspMgr *fcmocks.MockMSPManager, peers []apifabclient.Peer, t *testing.T) *chclient.ClientContext {

	testChannel, err := setupTestChannel()
	if err != nil {
		t.Fatalf("Failed to setup test channel: %s", err)
	}

	testChannel.SetMSPManager(mspMgr)

	orderer := fcmocks.NewMockOrderer("", nil)
	testChannel.AddOrderer(orderer)

	discoveryService, err := setupTestDiscovery(nil, nil)
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	selectionService, err := setupTestSelection(nil, peers)
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	return &chclient.ClientContext{
		Channel:   testChannel,
		Discovery: discoveryService,
		Selection: selectionService,
	}

}

var certPem = `-----BEGIN CERTIFICATE-----
MIIC5TCCAkagAwIBAgIUMYhiY5MS3jEmQ7Fz4X/e1Dx33J0wCgYIKoZIzj0EAwQw
gYwxCzAJBgNVBAYTAkNBMRAwDgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3Jv
bnRvMREwDwYDVQQKEwhsaW51eGN0bDEMMAoGA1UECxMDTGFiMTgwNgYDVQQDEy9s
aW51eGN0bCBFQ0MgUm9vdCBDZXJ0aWZpY2F0aW9uIEF1dGhvcml0eSAoTGFiKTAe
Fw0xNzEyMDEyMTEzMDBaFw0xODEyMDEyMTEzMDBaMGMxCzAJBgNVBAYTAkNBMRAw
DgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3JvbnRvMREwDwYDVQQKEwhsaW51
eGN0bDEMMAoGA1UECxMDTGFiMQ8wDQYDVQQDDAZzZGtfZ28wdjAQBgcqhkjOPQIB
BgUrgQQAIgNiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv2PotEEGuO6rMyaOu
pulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMhYZjcDi7YEOZ3Fs1h
xKmIxR+TTR2vf9KjgZAwgY0wDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsG
AQUFBwMCMAwGA1UdEwEB/wQCMAAwHQYDVR0OBBYEFDwS3xhpAWs81OVWvZt+iUNL
z26DMB8GA1UdIwQYMBaAFLRasbknomawJKuQGiyKs/RzTCujMBgGA1UdEQQRMA+C
DWZhYnJpY19zZGtfZ28wCgYIKoZIzj0EAwQDgYwAMIGIAkIAk1MxMogtMtNO0rM8
gw2rrxqbW67ulwmMQzp6EJbm/28T2pIoYWWyIwpzrquypI7BOuf8is5b7Jcgn9oz
7sdMTggCQgF7/8ZFl+wikAAPbciIL1I+LyCXKwXosdFL6KMT6/myYjsGNeeDeMbg
3YkZ9DhdH1tN4U/h+YulG/CkKOtUATtQxg==
-----END CERTIFICATE-----`
