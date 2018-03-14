/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package channel enables access to a channel on a Fabric network.
package channel

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel/invoke"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/greylist"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/fab")

const (
	defaultHandlerTimeout = time.Second * 180
)

// Client enables access to a channel on a Fabric network.
//
// A channel client instance provides a handler to interact with peers on specified channel.
// An application that requires interaction with multiple channels should create a separate
// instance of the channel client for each channel. Channel client supports non-admin functions only.
type Client struct {
	context         context.Channel
	membership      fab.ChannelMembership
	transactor      fab.Transactor
	eventService    fab.EventService
	greylist        *greylist.Filter
	discoveryFilter fab.TargetFilter
}

type customChannelContext struct {
	context.Channel
	discoveryService fab.DiscoveryService
}

func (ccc *customChannelContext) DiscoveryService() fab.DiscoveryService {
	return ccc.discoveryService
}

// ClientOption describes a functional parameter for the New constructor
type ClientOption func(*Client) error

// WithTargetFilter option to configure new
func WithTargetFilter(filter fab.TargetFilter) ClientOption {
	return func(client *Client) error {
		client.discoveryFilter = filter
		return nil
	}
}

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

	transactor, err := channelContext.ChannelService().Transactor()
	if err != nil {
		return nil, errors.WithMessage(err, "transactor creation failed")
	}

	membership, err := channelContext.ChannelService().Membership()
	if err != nil {
		return nil, errors.WithMessage(err, "membership creation failed")
	}

	channelClient := Client{
		membership:   membership,
		transactor:   transactor,
		eventService: eventService,
		greylist:     greylistProvider,
	}

	for _, param := range opts {
		param(&channelClient)
	}

	//target filter
	discoveryService := discovery.NewDiscoveryFilterService(channelContext.DiscoveryService(), channelClient.discoveryFilter)

	//greylist filter
	customDiscoveryService := discovery.NewDiscoveryFilterService(discoveryService, greylistProvider)

	//update context
	channelClient.context = &customChannelContext{Channel: channelContext, discoveryService: customDiscoveryService}

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

//prepareHandlerContexts prepares context objects for handlers
func (cc *Client) prepareHandlerContexts(request Request, o requestOptions) (*invoke.RequestContext, *invoke.ClientContext, error) {

	if request.ChaincodeID == "" || request.Fcn == "" {
		return nil, nil, errors.New("ChaincodeID and Fcn are required")
	}

	clientContext := &invoke.ClientContext{
		Selection:    cc.context.SelectionService(),
		Discovery:    cc.context.DiscoveryService(),
		Membership:   cc.membership,
		Transactor:   cc.transactor,
		EventService: cc.eventService,
	}

	requestContext := &invoke.RequestContext{
		Request:      invoke.Request(request),
		Opts:         invoke.Opts(o),
		Response:     invoke.Response{},
		RetryHandler: retry.New(o.Retry),
	}

	if requestContext.Opts.Timeout == 0 {
		to := cc.context.Config().Timeout(core.Execute)
		if to == 0 {
			requestContext.Opts.Timeout = defaultHandlerTimeout
		} else {
			requestContext.Opts.Timeout = to
		}
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

	if txnOpts.Timeout == 0 {
		return append(options, WithTimeout(cc.context.Config().TimeoutOrDefault(timeOutType)))
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
