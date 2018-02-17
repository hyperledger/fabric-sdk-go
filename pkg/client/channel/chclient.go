/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package channel enables access to a channel on a Fabric network.
package channel

import (
	"reflect"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel/invoke"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/greylist"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabric_sdk_go")

const (
	defaultHandlerTimeout = time.Second * 10
)

// Client enables access to a channel on a Fabric network.
//
// A channel client instance provides a handler to interact with peers on specified channel.
// An application that requires interaction with multiple channels should create a separate
// instance of the channel client for each channel. Channel client supports non-admin functions only.
type Client struct {
	context    context.ProviderContext
	discovery  fab.DiscoveryService
	selection  fab.SelectionService
	membership fab.ChannelMembership
	transactor fab.Transactor
	eventHub   fab.EventHub
	greylist   *greylist.Filter
}

// Context holds the providers and services needed to create a Client.
type Context struct {
	context.ProviderContext
	DiscoveryService fab.DiscoveryService
	SelectionService fab.SelectionService
	ChannelService   fab.ChannelService
}

// New returns a Client instance.
func New(c Context) (*Client, error) {
	greylistProvider := greylist.New(c.Config().TimeoutOrDefault(core.DiscoveryGreylistExpiry))

	eventHub, err := c.ChannelService.EventHub()
	if err != nil {
		return nil, errors.WithMessage(err, "event hub creation failed")
	}

	transactor, err := c.ChannelService.Transactor()
	if err != nil {
		return nil, errors.WithMessage(err, "transactor creation failed")
	}

	membership, err := c.ChannelService.Membership()
	if err != nil {
		return nil, errors.WithMessage(err, "membership creation failed")
	}

	channelClient := Client{
		greylist:   greylistProvider,
		context:    c,
		discovery:  discovery.NewDiscoveryFilterService(c.DiscoveryService, greylistProvider),
		selection:  c.SelectionService,
		membership: membership,
		transactor: transactor,
		eventHub:   eventHub,
	}

	return &channelClient, nil
}

// Query chaincode using request and optional options provided
func (cc *Client) Query(request Request, options ...Option) (Response, error) {
	return cc.InvokeHandler(invoke.NewQueryHandler(), request, cc.addDefaultTimeout(core.Query, options...)...)
}

// Execute prepares and executes transaction using request and optional options provided
func (cc *Client) Execute(request Request, options ...Option) (Response, error) {
	return cc.InvokeHandler(invoke.NewExecuteHandler(), request, cc.addDefaultTimeout(core.Execute, options...)...)
}

//InvokeHandler invokes handler using request and options provided
func (cc *Client) InvokeHandler(handler invoke.Handler, request Request, options ...Option) (Response, error) {
	//Read execute tx options
	txnOpts, err := cc.prepareOptsFromOptions(options...)
	if err != nil {
		return Response{}, err
	}

	//Prepare context objects for handler
	requestContext, clientContext, err := cc.prepareHandlerContexts(request, txnOpts)
	if err != nil {
		return Response{}, err
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
		return Response(requestContext.Response), requestContext.Error
	case <-time.After(requestContext.Opts.Timeout):
		return Response{}, status.New(status.ClientStatus, status.Timeout.ToInt32(),
			"request timed out", nil)
	}
}

func (cc *Client) resolveRetry(ctx *invoke.RequestContext, o opts) bool {
	errs, ok := ctx.Error.(multi.Errors)
	if !ok {
		errs = append(errs, ctx.Error)
	}
	for _, e := range errs {
		if ctx.RetryHandler.Required(e) {
			logger.Infof("Retrying on error %s", e)
			cc.greylist.Greylist(e)

			// Reset context parameters
			ctx.Opts.ProposalProcessors = o.ProposalProcessors
			ctx.Error = nil
			ctx.Response = invoke.Response{}

			return true
		}
	}
	return false
}

//prepareHandlerContexts prepares context objects for handlers
func (cc *Client) prepareHandlerContexts(request Request, o opts) (*invoke.RequestContext, *invoke.ClientContext, error) {

	if request.ChaincodeID == "" || request.Fcn == "" {
		return nil, nil, errors.New("ChaincodeID and Fcn are required")
	}

	clientContext := &invoke.ClientContext{
		Selection:  cc.selection,
		Discovery:  cc.discovery,
		Membership: cc.membership,
		Transactor: cc.transactor,
		EventHub:   cc.eventHub,
	}

	requestContext := &invoke.RequestContext{
		Request:      invoke.Request(request),
		Opts:         invoke.Opts(o),
		Response:     invoke.Response{},
		RetryHandler: retry.New(o.Retry),
	}

	if requestContext.Opts.Timeout == 0 {
		requestContext.Opts.Timeout = defaultHandlerTimeout
	}

	return requestContext, clientContext, nil
}

//prepareOptsFromOptions Reads apitxn.Opts from Option array
func (cc *Client) prepareOptsFromOptions(options ...Option) (opts, error) {
	txnOpts := opts{}
	for _, option := range options {
		err := option(&txnOpts)
		if err != nil {
			return txnOpts, errors.WithMessage(err, "Failed to read opts")
		}
	}
	return txnOpts, nil
}

//addDefaultTimeout adds given default timeout if it is missing in options
func (cc *Client) addDefaultTimeout(timeOutType core.TimeoutType, options ...Option) []Option {
	txnOpts := opts{}
	for _, option := range options {
		option(&txnOpts)
	}

	if txnOpts.Timeout == 0 {
		return append(options, WithTimeout(cc.context.Config().TimeoutOrDefault(timeOutType)))
	}
	return options
}

// Close releases channel client resources (disconnects event hub etc.)
func (cc *Client) Close() error {
	if cc.eventHub.IsConnected() == true {
		return cc.eventHub.Disconnect()
	}

	return nil
}

// RegisterChaincodeEvent registers chain code event
// @param {chan bool} channel which receives event details when the event is complete
// @returns {object} object handle that should be used to unregister
func (cc *Client) RegisterChaincodeEvent(notify chan<- *CCEvent, chainCodeID string, eventID string) (Registration, error) {

	if cc.eventHub.IsConnected() == false {
		if err := cc.eventHub.Connect(); err != nil {
			return nil, errors.WithMessage(err, "Event hub failed to connect")
		}
	}

	// Register callback for CE
	rce := cc.eventHub.RegisterChaincodeEvent(chainCodeID, eventID, func(ce *fab.ChaincodeEvent) {
		notify <- &CCEvent{ChaincodeID: ce.ChaincodeID, EventName: ce.EventName, TxID: ce.TxID, Payload: ce.Payload}
	})

	return rce, nil
}

// UnregisterChaincodeEvent removes chain code event registration
func (cc *Client) UnregisterChaincodeEvent(registration Registration) error {

	switch regType := registration.(type) {

	case *fab.ChainCodeCBE:
		cc.eventHub.UnregisterChaincodeEvent(regType)
	default:
		return errors.Errorf("Unsupported registration type: %v", reflect.TypeOf(registration))
	}

	return nil

}
