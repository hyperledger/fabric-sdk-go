/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package txnhandler

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	txnhandlerApi "github.com/hyperledger/fabric-sdk-go/api/apitxn/txnhandler"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	"github.com/stretchr/testify/assert"

	"strings"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	txnmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/mocks"
)

const (
	testTimeOut           = 20 * time.Second
	discoveryServiceError = "Discovery service error"
	selectionServiceError = "Selection service error"
	filterTxError         = "Filter Tx error"
)

func TestQueryHandlerSuccess(t *testing.T) {

	//Sample request
	request := apitxn.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, apitxn.Opts{}, t)
	clientContext := setupChannelClientContext(nil, nil, nil, t)

	//Get query handler
	queryHandler := NewQueryHandler()

	//Perform action through handler
	queryHandler.Handle(requestContext, clientContext)
	if requestContext.Response.Error != nil {
		t.Fatal("Query handler failed", requestContext.Response.Error)
	}
}

func TestExecuteTxHandlerSuccess(t *testing.T) {

	//Sample request
	request := apitxn.Request{ChaincodeID: "test", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, apitxn.Opts{}, t)
	clientContext := setupChannelClientContext(nil, nil, nil, t)

	//Prepare mock eventhub
	mockEventHub := fcmocks.NewMockEventHub()
	clientContext.EventHub = mockEventHub

	go func() {
		select {
		case callback := <-mockEventHub.RegisteredTxCallbacks:
			callback("txid", 0, nil)
		case <-time.After(requestContext.Opts.Timeout):
			t.Fatal("Execute handler : time out not expected")
		}
	}()

	//Get query handler
	executeHandler := NewExecuteHandler()
	//Perform action through handler
	executeHandler.Handle(requestContext, clientContext)
	assert.Nil(t, requestContext.Response.Error)
}

func TestQueryHandlerErrors(t *testing.T) {

	//Error Scenario 1
	request := apitxn.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, apitxn.Opts{}, t)
	clientContext := setupChannelClientContext(errors.New(discoveryServiceError), nil, nil, t)

	//Get query handler
	queryHandler := NewQueryHandler()

	//Perform action through handler
	queryHandler.Handle(requestContext, clientContext)
	if requestContext.Response.Error == nil || !strings.Contains(requestContext.Response.Error.Error(), discoveryServiceError) {
		t.Fatal("Expected error: ", discoveryServiceError, ", Received error:", requestContext.Response.Error.Error())
	}

	//Error Scenario 2
	clientContext = setupChannelClientContext(nil, errors.New(selectionServiceError), nil, t)

	//Perform action through handler
	queryHandler.Handle(requestContext, clientContext)
	if requestContext.Response.Error == nil || !strings.Contains(requestContext.Response.Error.Error(), selectionServiceError) {
		t.Fatal("Expected error: ", selectionServiceError, ", Received error:", requestContext.Response.Error.Error())
	}

}

//prepareHandlerContexts prepares context objects for handlers
func prepareRequestContext(request apitxn.Request, opts apitxn.Opts, t *testing.T) *txnhandlerApi.RequestContext {

	var requestContext *txnhandlerApi.RequestContext

	requestContext = &txnhandlerApi.RequestContext{Request: request,
		Opts:     opts,
		Response: apitxn.Response{},
	}

	requestContext.Opts.Timeout = testTimeOut

	return requestContext

}

func setupTestChannel() (*channel.Channel, error) {
	ctx := setupTestContext()
	return channel.New(ctx, "testChannel")
}

func setupTestContext() apifabclient.Context {
	user := fcmocks.NewMockUser("test")
	ctx := fcmocks.NewMockContext(user)
	return ctx
}

func setupChannelClientContext(discErr error, selectionErr error, peers []apifabclient.Peer, t *testing.T) *txnhandlerApi.ClientContext {

	testChannel, err := setupTestChannel()
	if err != nil {
		t.Fatalf("Failed to setup test channel: %s", err)
	}

	orderer := fcmocks.NewMockOrderer("", nil)
	testChannel.AddOrderer(orderer)

	discoveryService, err := setupTestDiscovery(discErr, nil)
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	selectionService, err := setupTestSelection(selectionErr, peers)
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	return &txnhandlerApi.ClientContext{
		Channel:   testChannel,
		Discovery: discoveryService,
		Selection: selectionService,
	}

}

func setupTestDiscovery(discErr error, peers []apifabclient.Peer) (apifabclient.DiscoveryService, error) {

	mockDiscovery, err := txnmocks.NewMockDiscoveryProvider(discErr, peers)
	if err != nil {
		return nil, errors.WithMessage(err, "NewMockDiscoveryProvider failed")
	}

	return mockDiscovery.NewDiscoveryService("mychannel")
}

func setupTestSelection(discErr error, peers []apifabclient.Peer) (apifabclient.SelectionService, error) {

	mockSelection, err := txnmocks.NewMockSelectionProvider(discErr, peers)
	if err != nil {
		return nil, errors.WithMessage(err, "NewMockSelectinProvider failed")
	}

	return mockSelection.NewSelectionService("mychannel")
}
