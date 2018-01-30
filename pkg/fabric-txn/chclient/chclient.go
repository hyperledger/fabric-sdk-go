/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package chclient enables channel client
package chclient

import (
	"reflect"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn/txnhandler"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	txnHandlerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/txnhandler"
	"github.com/hyperledger/fabric-sdk-go/pkg/status"
)

const (
	defaultHandlerTimeout = time.Second * 10
)

// ChannelClient enables access to a Fabric network.
type ChannelClient struct {
	context   fab.ProviderContext
	discovery fab.DiscoveryService
	selection fab.SelectionService
	channel   fab.Channel
	eventHub  fab.EventHub
}

// Context holds the providers and services needed to create a ChannelClient.
type Context struct {
	fab.ProviderContext
	DiscoveryService fab.DiscoveryService
	SelectionService fab.SelectionService
	Channel          fab.Channel
	EventHub         fab.EventHub
}

// New returns a ChannelClient instance.
func New(c Context) (*ChannelClient, error) {

	channelClient := ChannelClient{
		context:   c,
		discovery: c.DiscoveryService,
		selection: c.SelectionService,
		channel:   c.Channel,
		eventHub:  c.EventHub,
	}

	return &channelClient, nil
}

// Query chaincode using request and optional options provided
func (cc *ChannelClient) Query(request apitxn.Request, options ...apitxn.Option) apitxn.Response {
	return cc.InvokeHandler(txnHandlerImpl.NewQueryHandler(), request, cc.addDefaultTimeout(apiconfig.Query, options...)...)
}

// Execute prepares and executes transaction using request and optional options provided
func (cc *ChannelClient) Execute(request apitxn.Request, options ...apitxn.Option) apitxn.Response {
	return cc.InvokeHandler(txnHandlerImpl.NewExecuteHandler(), request, cc.addDefaultTimeout(apiconfig.Execute, options...)...)
}

//InvokeHandler invokes handler using request and options provided
func (cc *ChannelClient) InvokeHandler(handler txnhandler.Handler, request apitxn.Request, options ...apitxn.Option) apitxn.Response {
	//TODO: this function going to be exposed through ChannelClient interface
	//Read execute tx options
	txnOpts, err := cc.prepareOptsFromOptions(options...)
	if err != nil {
		return apitxn.Response{Error: err}
	}

	//Prepare context objects for handler
	requestContext, clientContext, err := cc.prepareHandlerContexts(request, txnOpts)
	if err != nil {
		return apitxn.Response{Error: err}
	}

	complete := make(chan bool)

	go func() {
		//Perform action through handler
		handler.Handle(requestContext, clientContext)
		complete <- true
	}()
	select {
	case <-complete:
		return requestContext.Response
	case <-time.After(txnOpts.Timeout):
		return apitxn.Response{Error: status.New(status.ClientStatus, status.Timeout.ToInt32(),
			"Operation timed out", nil)}
	}
}

//prepareHandlerContexts prepares context objects for handlers
func (cc *ChannelClient) prepareHandlerContexts(request apitxn.Request, options apitxn.Opts) (*txnhandler.RequestContext, *txnhandler.ClientContext, error) {

	if request.ChaincodeID == "" || request.Fcn == "" {
		return nil, nil, errors.New("ChaincodeID and Fcn are required")
	}

	clientContext := &txnhandler.ClientContext{
		Channel:   cc.channel,
		Selection: cc.selection,
		Discovery: cc.discovery,
		EventHub:  cc.eventHub,
	}

	requestContext := &txnhandler.RequestContext{
		Request:  request,
		Opts:     options,
		Response: apitxn.Response{},
	}

	if requestContext.Opts.Timeout == 0 {
		requestContext.Opts.Timeout = defaultHandlerTimeout
	}

	return requestContext, clientContext, nil

}

//prepareOptsFromOptions Reads apitxn.Opts from apitxn.Option array
func (cc *ChannelClient) prepareOptsFromOptions(options ...apitxn.Option) (apitxn.Opts, error) {
	txnOpts := apitxn.Opts{}
	for _, option := range options {
		err := option(&txnOpts)
		if err != nil {
			return txnOpts, errors.WithMessage(err, "Failed to read opts")
		}
	}
	return txnOpts, nil
}

//addDefaultTimeout adds given default timeout if it is missing in options
func (cc *ChannelClient) addDefaultTimeout(timeOutType apiconfig.TimeoutType, options ...apitxn.Option) []apitxn.Option {
	txnOpts := apitxn.Opts{}
	for _, option := range options {
		option(&txnOpts)
	}

	if txnOpts.Timeout == 0 {
		return append(options, apitxn.WithTimeout(cc.context.Config().TimeoutOrDefault(timeOutType)))
	}
	return options
}

// Close releases channel client resources (disconnects event hub etc.)
func (cc *ChannelClient) Close() error {
	if cc.eventHub.IsConnected() == true {
		return cc.eventHub.Disconnect()
	}

	return nil
}

// RegisterChaincodeEvent registers chain code event
// @param {chan bool} channel which receives event details when the event is complete
// @returns {object} object handle that should be used to unregister
func (cc *ChannelClient) RegisterChaincodeEvent(notify chan<- *apitxn.CCEvent, chainCodeID string, eventID string) apitxn.Registration {

	// Register callback for CE
	rce := cc.eventHub.RegisterChaincodeEvent(chainCodeID, eventID, func(ce *fab.ChaincodeEvent) {
		notify <- &apitxn.CCEvent{ChaincodeID: ce.ChaincodeID, EventName: ce.EventName, TxID: ce.TxID, Payload: ce.Payload}
	})

	return rce
}

// UnregisterChaincodeEvent removes chain code event registration
func (cc *ChannelClient) UnregisterChaincodeEvent(registration apitxn.Registration) error {

	switch regType := registration.(type) {

	case *fab.ChainCodeCBE:
		cc.eventHub.UnregisterChaincodeEvent(regType)
	default:
		return errors.Errorf("Unsupported registration type: %v", reflect.TypeOf(registration))
	}

	return nil

}
