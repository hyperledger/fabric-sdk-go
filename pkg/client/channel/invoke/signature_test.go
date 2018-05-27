/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package invoke

import (
	"strings"
	"testing"

	txnmocks "github.com/hyperledger/fabric-sdk-go/pkg/client/common/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/stretchr/testify/assert"
)

func TestSignatureValidationHandlerSuccess(t *testing.T) {
	request := Request{ChaincodeID: "test", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, Opts{}, t)

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}

	clientContext := setupContextForSignatureValidation(nil, nil, []fab.Peer{mockPeer1}, t)

	handler := NewQueryHandler()
	handler.Handle(requestContext, clientContext)
	assert.Nil(t, requestContext.Error)
}

func TestSignatureValidationCreatorValidateError(t *testing.T) {
	validateErr := status.New(status.EndorserClientStatus, status.SignatureVerificationFailed.ToInt32(), "", nil)
	// Sample request
	request := Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}
	requestContext := prepareRequestContext(request, Opts{}, t)
	handler := NewQueryHandler()

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}

	clientContext := setupContextForSignatureValidation(nil, validateErr, []fab.Peer{mockPeer1}, t)
	handler.Handle(requestContext, clientContext)
	verifyExpectedError(requestContext, validateErr.Error(), t)
}

func TestSignatureValidationCreatorVerifyError(t *testing.T) {
	verifyErr := status.New(status.EndorserClientStatus, status.SignatureVerificationFailed.ToInt32(), "", nil)

	// Sample request
	request := Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}
	requestContext := prepareRequestContext(request, Opts{}, t)
	handler := NewQueryHandler()

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}

	clientContext := setupContextForSignatureValidation(verifyErr, nil, []fab.Peer{mockPeer1}, t)
	handler.Handle(requestContext, clientContext)
	verifyExpectedError(requestContext, verifyErr.Error(), t)
}

func verifyExpectedError(requestContext *RequestContext, expected string, t *testing.T) {
	assert.NotNil(t, requestContext.Error)
	if requestContext.Error == nil || !strings.Contains(requestContext.Error.Error(), expected) {
		t.Fatal("Expected error: ", expected, ", Received error:", requestContext.Error)
	}
}

func setupContextForSignatureValidation(verifyErr, validateErr error, peers []fab.Peer, t *testing.T) *ClientContext {
	ctx := setupTestContext()
	membership := fcmocks.NewMockMembership()
	membership.ValidateErr = validateErr
	membership.VerifyErr = verifyErr

	transactor := txnmocks.MockTransactor{
		Ctx:       ctx,
		ChannelID: "",
	}

	return &ClientContext{
		Membership: membership,
		Discovery:  fcmocks.NewMockDiscoveryService(nil),
		Selection:  fcmocks.NewMockSelectionService(nil, peers...),
		Transactor: &transactor,
	}

}
