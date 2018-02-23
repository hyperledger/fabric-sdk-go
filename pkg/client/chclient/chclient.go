/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package chclient enables channel client
package chclient

import (
	"reflect"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/discovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/discovery/greylist"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
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

// ChannelClient enables access to a Fabric network.
/*
 * A channel client instance provides a handler to interact with peers on specified channel.
 * An application that requires interaction with multiple channels should create a separate
 * instance of the channel client for each channel. Channel client supports non-admin functions only.
 *
 * Each Client instance maintains {@link Channel} instance representing channel and the associated
 * private ledgers.
 *
 */
type ChannelClient struct {
	context    context.ProviderContext
	discovery  fab.DiscoveryService
	selection  fab.SelectionService
	channel    fab.Channel
	transactor fab.Transactor
	eventHub   fab.EventHub
	greylist   *greylist.Filter
}

// Context holds the providers and services needed to create a ChannelClient.
type Context struct {
	context.ProviderContext
	DiscoveryService fab.DiscoveryService
	SelectionService fab.SelectionService
	ChannelService   fab.ChannelService
}

// New returns a ChannelClient instance.
func New(c Context) (*ChannelClient, error) {
	greylistProvider := greylist.New(c.Config().TimeoutOrDefault(core.DiscoveryGreylistExpiry))

	eventHub, err := c.ChannelService.EventHub()
	if err != nil {
		return nil, errors.WithMessage(err, "event hub creation failed")
	}

	transactor, err := c.ChannelService.Transactor()
	if err != nil {
		return nil, errors.WithMessage(err, "transactor creation failed")
	}

	// TODO - this should be removed once MSP is split out.
	channel, err := c.ChannelService.Channel()
	if err != nil {
		return nil, errors.WithMessage(err, "channel client creation failed")
	}

	channelClient := ChannelClient{
		greylist:   greylistProvider,
		context:    c,
		discovery:  discovery.NewDiscoveryFilterService(c.DiscoveryService, greylistProvider),
		selection:  c.SelectionService,
		channel:    channel,
		transactor: transactor,
		eventHub:   eventHub,
	}

	return &channelClient, nil
}

// Query chaincode using request and optional options provided
func (cc *ChannelClient) Query(request Request, options ...Option) (Response, error) {
	return cc.InvokeHandler(NewQueryHandler(), request, cc.addDefaultTimeout(core.Query, options...)...)
}

// Execute prepares and executes transaction using request and optional options provided
func (cc *ChannelClient) Execute(request Request, options ...Option) (Response, error) {
	return cc.InvokeHandler(NewExecuteHandler(), request, cc.addDefaultTimeout(core.Execute, options...)...)
}

//InvokeHandler invokes handler using request and options provided
func (cc *ChannelClient) InvokeHandler(handler Handler, request Request, options ...Option) (Response, error) {
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
		return requestContext.Response, requestContext.Error
	case <-time.After(requestContext.Opts.Timeout):
		return Response{}, status.New(status.ClientStatus, status.Timeout.ToInt32(),
			"request timed out", nil)
	}
}

func (cc *ChannelClient) resolveRetry(ctx *RequestContext, opts Opts) bool {
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
			ctx.Response = Response{}

			return true
		}
	}
	return false
}

//prepareHandlerContexts prepares context objects for handlers
func (cc *ChannelClient) prepareHandlerContexts(request Request, options Opts) (*RequestContext, *ClientContext, error) {

	if request.ChaincodeID == "" || request.Fcn == "" {
		return nil, nil, errors.New("ChaincodeID and Fcn are required")
	}

	clientContext := &ClientContext{
		Selection:  cc.selection,
		Discovery:  cc.discovery,
		Channel:    cc.channel,
		Transactor: cc.transactor,
		EventHub:   cc.eventHub,
	}

	requestContext := &RequestContext{
		Request:      request,
		Opts:         options,
		Response:     Response{},
		RetryHandler: retry.New(options.Retry),
	}

	if requestContext.Opts.Timeout == 0 {
		requestContext.Opts.Timeout = defaultHandlerTimeout
	}

	return requestContext, clientContext, nil
}

//prepareOptsFromOptions Reads apitxn.Opts from Option array
func (cc *ChannelClient) prepareOptsFromOptions(options ...Option) (Opts, error) {
	txnOpts := Opts{}
	for _, option := range options {
		err := option(&txnOpts)
		if err != nil {
			return txnOpts, errors.WithMessage(err, "Failed to read opts")
		}
	}
	return txnOpts, nil
}

//addDefaultTimeout adds given default timeout if it is missing in options
func (cc *ChannelClient) addDefaultTimeout(timeOutType core.TimeoutType, options ...Option) []Option {
	txnOpts := Opts{}
	for _, option := range options {
		option(&txnOpts)
	}

	if txnOpts.Timeout == 0 {
		return append(options, WithTimeout(cc.context.Config().TimeoutOrDefault(timeOutType)))
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
func (cc *ChannelClient) RegisterChaincodeEvent(notify chan<- *CCEvent, chainCodeID string, eventID string) (Registration, error) {

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
func (cc *ChannelClient) UnregisterChaincodeEvent(registration Registration) error {

	switch regType := registration.(type) {

	case *fab.ChainCodeCBE:
		cc.eventHub.UnregisterChaincodeEvent(regType)
	default:
		return errors.Errorf("Unsupported registration type: %v", reflect.TypeOf(registration))
	}

	return nil

}
