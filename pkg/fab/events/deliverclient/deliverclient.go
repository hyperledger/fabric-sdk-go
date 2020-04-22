/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package deliverclient

import (
	"math"
	"time"

	ab "github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	fabcontext "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client"
	deliverconn "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/connection"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/dispatcher"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/endpoint"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/fab")

// deliverProvider is the connection provider used for connecting to the Deliver service
var deliverProvider = func(context fabcontext.Client, chConfig fab.ChannelCfg, peer fab.Peer) (api.Connection, error) {
	if peer == nil {
		return nil, errors.New("Peer is nil")
	}

	eventEndpoint, ok := peer.(api.EventEndpoint)
	if !ok {
		panic("peer is not an EventEndpoint")
	}
	return deliverconn.New(context, chConfig, deliverconn.Deliver, peer.URL(), eventEndpoint.Opts()...)
}

// deliverFilteredProvider is the connection provider used for connecting to the DeliverFiltered service
var deliverFilteredProvider = func(context fabcontext.Client, chConfig fab.ChannelCfg, peer fab.Peer) (api.Connection, error) {
	if peer == nil {
		return nil, errors.New("Peer is nil")
	}

	eventEndpoint, ok := peer.(api.EventEndpoint)
	if !ok {
		panic("peer is not an EventEndpoint")
	}
	return deliverconn.New(context, chConfig, deliverconn.DeliverFiltered, peer.URL(), eventEndpoint.Opts()...)
}

// Client connects to a peer and receives channel events, such as bock, filtered block, chaincode, and transaction status events.
type Client struct {
	*client.Client
	params
}

// New returns a new deliver event client
func New(context fabcontext.Client, chConfig fab.ChannelCfg, discoveryService fab.DiscoveryService, opts ...options.Opt) (*Client, error) {
	params := defaultParams()
	options.Apply(params, opts)

	// Use a custom Discovery Service which wraps the given discovery service
	// and produces event endpoints containing additional GRPC options.
	discoveryWrapper, err := endpoint.NewEndpointDiscoveryWrapper(context, chConfig.ID(), discoveryService)
	if err != nil {
		return nil, err
	}

	dispatcher := dispatcher.New(context, chConfig, discoveryWrapper, params.connProvider, opts...)

	//default seek type is `Newest`
	if params.seekType == "" {
		params.seekType = seek.Newest
		//discard (do not publish) next BlockEvent/FilteredBlockEvent in dispatcher, since default seek type 'newest' is
		// only needed for block height calculations
		dispatcher.UpdateLastBlockInfoOnly()
	}

	client := &Client{
		Client: client.New(dispatcher, opts...),
		params: *params,
	}

	client.SetAfterConnectHandler(client.seek)
	client.SetBeforeReconnectHandler(client.setSeekFromLastBlockReceived)

	if err := client.Start(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) seek() error {
	logger.Debug("Sending seek request....")

	seekInfo, err := c.seekInfo()
	if err != nil {
		return err
	}

	errch := make(chan error, 1)
	err1 := c.Submit(dispatcher.NewSeekEvent(seekInfo, errch))
	if err1 != nil {
		return err1
	}
	select {
	case err = <-errch:
	case <-time.After(c.respTimeout):
		err = errors.New("timeout waiting for deliver status response")
	}

	if err != nil {
		logger.Errorf("Unable to send seek request: %s", err)
		return err
	}

	logger.Debug("Successfully sent seek")
	return nil
}

func (c *Client) setSeekFromLastBlockReceived() error {
	c.Lock()
	defer c.Unlock()

	// Make sure that, when we reconnect, we receive all of the events that we've missed
	lastBlockNum := c.Dispatcher().LastBlockNum()
	if lastBlockNum < math.MaxUint64 {
		c.seekType = seek.FromBlock
		c.fromBlock = c.Dispatcher().LastBlockNum() + 1
		logger.Debugf("Setting seek info from last block received + 1: %d", c.fromBlock)
	} else {
		// We haven't received any blocks yet. Just ask for the newest
		logger.Debugf("Setting seek info from newest")
		c.seekType = seek.Newest
	}
	return nil
}

func (c *Client) seekInfo() (*ab.SeekInfo, error) {
	c.RLock()
	defer c.RUnlock()

	switch c.seekType {
	case seek.Newest:
		logger.Debugf("Returning seek info: Newest")
		return seek.InfoNewest(), nil
	case seek.Oldest:
		logger.Debugf("Returning seek info: Oldest")
		return seek.InfoOldest(), nil
	case seek.FromBlock:
		logger.Debugf("Returning seek info: FromBlock(%d)", c.fromBlock)
		return seek.InfoFrom(c.fromBlock), nil
	default:
		return nil, errors.Errorf("unsupported seek type:[%s]", c.seekType)
	}
}
