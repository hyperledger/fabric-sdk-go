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

	"github.com/hyperledger/fabric-sdk-go/api/apitxn/chclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/discovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/discovery/greylist"
	txnHandlerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/txnhandler"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabric_sdk_go")

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
	greylist  *greylist.Filter
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
	greylistProvider := greylist.New(c.Config().TimeoutOrDefault(apiconfig.DiscoveryGreylistExpiry))

	channelClient := ChannelClient{
		greylist:  greylistProvider,
		context:   c,
		discovery: discovery.NewDiscoveryFilterService(c.DiscoveryService, greylistProvider),
		selection: c.SelectionService,
		channel:   c.Channel,
		eventHub:  c.EventHub,
	}

	return &channelClient, nil
}

// Query chaincode using request and optional options provided
func (cc *ChannelClient) Query(request chclient.Request, options ...chclient.Option) (chclient.Response, error) {
	return cc.InvokeHandler(txnHandlerImpl.NewQueryHandler(), request, cc.addDefaultTimeout(apiconfig.Query, options...)...)
}

// Execute prepares and executes transaction using request and optional options provided
func (cc *ChannelClient) Execute(request chclient.Request, options ...chclient.Option) (chclient.Response, error) {
	return cc.InvokeHandler(txnHandlerImpl.NewExecuteHandler(), request, cc.addDefaultTimeout(apiconfig.Execute, options...)...)
}

//InvokeHandler invokes handler using request and options provided
func (cc *ChannelClient) InvokeHandler(handler chclient.Handler, request chclient.Request, options ...chclient.Option) (chclient.Response, error) {
	//Read execute tx options
	txnOpts, err := cc.prepareOptsFromOptions(options...)
	if err != nil {
		return chclient.Response{}, err
	}

	//Prepare context objects for handler
	requestContext, clientContext, err := cc.prepareHandlerContexts(request, txnOpts)
	if err != nil {
		return chclient.Response{}, err
	}

	complete := make(chan bool)

	go func() {
	handleInvoke:
		//Perform action through handler
		handler.Handle(requestContext, clientContext)
		if cc.resolveRetry(requestContext, txnOpts) {
			goto handleInvoke
		}
		complete <- true
	}()
	select {
	case <-complete:
		return requestContext.Response, requestContext.Error
	case <-time.After(requestContext.Opts.Timeout):
		return chclient.Response{}, status.New(status.ClientStatus, status.Timeout.ToInt32(),
			"request timed out", nil)
	}
}

func (cc *ChannelClient) resolveRetry(ctx *chclient.RequestContext, opts chclient.Opts) bool {
	errs, ok := ctx.Error.(multi.Errors)
	if !ok {
		errs = append(errs, ctx.Error)
	}
	for _, e := range errs {
		if ctx.RetryHandler.Required(e) {
			logger.Infof("Retrying on error %s", e)
			cc.greylist.Greylist(e)

			// Reset context parameters
			ctx.Opts.ProposalProcessors = opts.ProposalProcessors
			ctx.Error = nil
			ctx.Response = chclient.Response{}

			return true
		}
	}
	return false
}

//prepareHandlerContexts prepares context objects for handlers
func (cc *ChannelClient) prepareHandlerContexts(request chclient.Request, options chclient.Opts) (*chclient.RequestContext, *chclient.ClientContext, error) {

	if request.ChaincodeID == "" || request.Fcn == "" {
		return nil, nil, errors.New("ChaincodeID and Fcn are required")
	}

	clientContext := &chclient.ClientContext{
		Channel:   cc.channel,
		Selection: cc.selection,
		Discovery: cc.discovery,
		EventHub:  cc.eventHub,
	}

	requestContext := &chclient.RequestContext{
		Request:      request,
		Opts:         options,
		Response:     chclient.Response{},
		RetryHandler: retry.New(options.Retry),
	}

	if requestContext.Opts.Timeout == 0 {
		requestContext.Opts.Timeout = defaultHandlerTimeout
	}

	return requestContext, clientContext, nil
}

//prepareOptsFromOptions Reads apitxn.Opts from chclient.Option array
func (cc *ChannelClient) prepareOptsFromOptions(options ...chclient.Option) (chclient.Opts, error) {
	txnOpts := chclient.Opts{}
	for _, option := range options {
		err := option(&txnOpts)
		if err != nil {
			return txnOpts, errors.WithMessage(err, "Failed to read opts")
		}
	}
	return txnOpts, nil
}

//addDefaultTimeout adds given default timeout if it is missing in options
func (cc *ChannelClient) addDefaultTimeout(timeOutType apiconfig.TimeoutType, options ...chclient.Option) []chclient.Option {
	txnOpts := chclient.Opts{}
	for _, option := range options {
		option(&txnOpts)
	}

	if txnOpts.Timeout == 0 {
		return append(options, chclient.WithTimeout(cc.context.Config().TimeoutOrDefault(timeOutType)))
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
func (cc *ChannelClient) RegisterChaincodeEvent(notify chan<- *chclient.CCEvent, chainCodeID string, eventID string) (chclient.Registration, error) {

	if cc.eventHub.IsConnected() == false {
		if err := cc.eventHub.Connect(); err != nil {
			return nil, errors.WithMessage(err, "Event hub failed to connect")
		}
	}

	// Register callback for CE
	rce := cc.eventHub.RegisterChaincodeEvent(chainCodeID, eventID, func(ce *fab.ChaincodeEvent) {
		notify <- &chclient.CCEvent{ChaincodeID: ce.ChaincodeID, EventName: ce.EventName, TxID: ce.TxID, Payload: ce.Payload}
	})

	return rce, nil
}

// UnregisterChaincodeEvent removes chain code event registration
func (cc *ChannelClient) UnregisterChaincodeEvent(registration chclient.Registration) error {

	switch regType := registration.(type) {

	case *fab.ChainCodeCBE:
		cc.eventHub.UnregisterChaincodeEvent(regType)
	default:
		return errors.Errorf("Unsupported registration type: %v", reflect.TypeOf(registration))
	}

	return nil

}
