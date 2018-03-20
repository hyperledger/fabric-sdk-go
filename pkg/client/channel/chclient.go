/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package channel enables access to a channel on a Fabric network.
package channel

import (
	reqContext "context"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel/invoke"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/greylist"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/errors/multi"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/client")

// Client enables access to a channel on a Fabric network.
//
// A channel client instance provides a handler to interact with peers on specified channel.
// An application that requires interaction with multiple channels should create a separate
// instance of the channel client for each channel. Channel client supports non-admin functions only.
type Client struct {
	context      context.Channel
	membership   fab.ChannelMembership
	eventService fab.EventService
	greylist     *greylist.Filter
}

// ClientOption describes a functional parameter for the New constructor
type ClientOption func(*Client) error

// New returns a Client instance.
func New(channelProvider context.ChannelProvider, opts ...ClientOption) (*Client, error) {

	channelContext, err := channelProvider()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create channel context")
	}

	greylistProvider := greylist.New(channelContext.Config().TimeoutOrDefault(core.DiscoveryGreylistExpiry))

	if channelContext.ChannelService() == nil {
		return nil, errors.New("channel service not initialized")
	}

	eventService, err := channelContext.ChannelService().EventService()
	if err != nil {
		return nil, errors.WithMessage(err, "event service creation failed")
	}

	membership, err := channelContext.ChannelService().Membership()
	if err != nil {
		return nil, errors.WithMessage(err, "membership creation failed")
	}

	channelClient := Client{
		membership:   membership,
		eventService: eventService,
		greylist:     greylistProvider,
		context:      channelContext,
	}

	for _, param := range opts {
		param(&channelClient)
	}

	return &channelClient, nil
}

// Query chaincode using request and optional options provided
func (cc *Client) Query(request Request, options ...RequestOption) (Response, error) {
	return cc.InvokeHandler(invoke.NewQueryHandler(), request, cc.addDefaultTimeout(cc.context, core.Query, options...)...)
}

// Execute prepares and executes transaction using request and optional options provided
func (cc *Client) Execute(request Request, options ...RequestOption) (Response, error) {
	return cc.InvokeHandler(invoke.NewExecuteHandler(), request, cc.addDefaultTimeout(cc.context, core.Execute, options...)...)
}

//InvokeHandler invokes handler using request and options provided
func (cc *Client) InvokeHandler(handler invoke.Handler, request Request, options ...RequestOption) (Response, error) {
	//Read execute tx options
	txnOpts, err := cc.prepareOptsFromOptions(cc.context, options...)
	if err != nil {
		return Response{}, err
	}

	reqCtx, cancel := cc.createReqContext(&txnOpts)
	defer cancel()

	//Prepare context objects for handler
	requestContext, clientContext, err := cc.prepareHandlerContexts(reqCtx, request, txnOpts)
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
	case <-reqCtx.Done():
		return Response{}, status.New(status.ClientStatus, status.Timeout.ToInt32(),
			"request timed out or been cancelled", nil)
	}
}

func (cc *Client) resolveRetry(ctx *invoke.RequestContext, o requestOptions) bool {
	errs, ok := ctx.Error.(multi.Errors)
	if !ok {
		errs = append(errs, ctx.Error)
	}
	for _, e := range errs {
		if ctx.RetryHandler.Required(e) {
			logger.Infof("Retrying on error %s", e)
			cc.greylist.Greylist(e)

			// Reset context parameters
			ctx.Opts.Targets = o.Targets
			ctx.Error = nil
			ctx.Response = invoke.Response{}

			return true
		}
	}
	return false
}

//createReqContext creates req context for invoke handler
func (cc *Client) createReqContext(txnOpts *requestOptions) (reqContext.Context, reqContext.CancelFunc) {

	if txnOpts.Timeouts == nil {
		txnOpts.Timeouts = make(map[core.TimeoutType]time.Duration)
	}

	//setting default timeouts when not provided
	if txnOpts.Timeouts[core.Execute] == 0 {
		txnOpts.Timeouts[core.Execute] = cc.context.Config().TimeoutOrDefault(core.Execute)
	}

	reqCtx, cancel := contextImpl.NewRequest(cc.context, contextImpl.WithTimeout(txnOpts.Timeouts[core.Execute]),
		contextImpl.WithParent(txnOpts.ParentContext))
	//Add timeout overrides here as a value so that it can be used by immediate child contexts (in handlers/transactors)
	reqCtx = reqContext.WithValue(reqCtx, contextImpl.ReqContextTimeoutOverrides, txnOpts.Timeouts)

	return reqCtx, cancel
}

//prepareHandlerContexts prepares context objects for handlers
func (cc *Client) prepareHandlerContexts(reqCtx reqContext.Context, request Request, o requestOptions) (*invoke.RequestContext, *invoke.ClientContext, error) {

	if request.ChaincodeID == "" || request.Fcn == "" {
		return nil, nil, errors.New("ChaincodeID and Fcn are required")
	}

	chConfig, err := cc.context.ChannelService().ChannelConfig()
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to retrieve channel config")
	}
	transactor, err := cc.context.InfraProvider().CreateChannelTransactor(reqCtx, chConfig)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to create transactor")
	}

	peerFilter := func(peer fab.Peer) bool {
		if !cc.greylist.Accept(peer) {
			return false
		}
		if o.TargetFilter != nil && !o.TargetFilter.Accept(peer) {
			return false
		}
		return true
	}

	clientContext := &invoke.ClientContext{
		Selection:    cc.context.SelectionService(),
		Discovery:    cc.context.DiscoveryService(),
		Membership:   cc.membership,
		Transactor:   transactor,
		EventService: cc.eventService,
	}

	requestContext := &invoke.RequestContext{
		Request:         invoke.Request(request),
		Opts:            invoke.Opts(o),
		Response:        invoke.Response{},
		RetryHandler:    retry.New(o.Retry),
		Ctx:             reqCtx,
		SelectionFilter: peerFilter,
	}

	return requestContext, clientContext, nil
}

//prepareOptsFromOptions Reads apitxn.Opts from Option array
func (cc *Client) prepareOptsFromOptions(ctx context.Client, options ...RequestOption) (requestOptions, error) {
	txnOpts := requestOptions{}
	for _, option := range options {
		err := option(ctx, &txnOpts)
		if err != nil {
			return txnOpts, errors.WithMessage(err, "Failed to read opts")
		}
	}
	return txnOpts, nil
}

//addDefaultTimeout adds given default timeout if it is missing in options
func (cc *Client) addDefaultTimeout(ctx context.Client, timeOutType core.TimeoutType, options ...RequestOption) []RequestOption {
	txnOpts := requestOptions{}
	for _, option := range options {
		option(ctx, &txnOpts)
	}

	if txnOpts.Timeouts[timeOutType] == 0 {
		//InvokeHandler relies on Execute timeout
		return append(options, WithTimeout(core.Execute, cc.context.Config().TimeoutOrDefault(timeOutType)))
	}
	return options
}

// RegisterChaincodeEvent registers chain code event
// @param {chan bool} channel which receives event details when the event is complete
// @returns {object} object handle that should be used to unregister
func (cc *Client) RegisterChaincodeEvent(chainCodeID string, eventFilter string) (fab.Registration, <-chan *fab.CCEvent, error) {
	// Register callback for CE
	return cc.eventService.RegisterChaincodeEvent(chainCodeID, eventFilter)
}

// UnregisterChaincodeEvent removes chain code event registration
func (cc *Client) UnregisterChaincodeEvent(registration fab.Registration) {
	cc.eventService.Unregister(registration)
}
