/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package deliverclient

import (
	"math"
	"sync"
	"time"

	ab "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
	fabcontext "github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client"
	deliverconn "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/connection"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/dispatcher"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/options"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/fab")

// deliverProvider is the connection provider used for connecting to the Deliver service
var deliverProvider = func(channelID string, context fabcontext.Client, peer fab.Peer) (api.Connection, error) {
	return deliverconn.New(context, channelID, deliverconn.Deliver, peer.URL())
}

// deliverFilteredProvider is the connection provider used for connecting to the DeliverFiltered service
var deliverFilteredProvider = func(channelID string, context fabcontext.Client, peer fab.Peer) (api.Connection, error) {
	return deliverconn.New(context, channelID, deliverconn.DeliverFiltered, peer.URL())
}

// Client connects to a peer and receives channel events, such as bock, filtered block, chaincode, and transaction status events.
type Client struct {
	sync.RWMutex
	client.Client
	params
	connEvent            chan *fab.ConnectionEvent
	connectionState      int32
	stopped              int32
	registerOnce         sync.Once
	blockEventsPermitted bool
}

// New returns a new deliver event client
func New(context fabcontext.Client, channelID string, discoveryService fab.DiscoveryService, opts ...options.Opt) (*Client, error) {
	if channelID == "" {
		return nil, errors.New("expecting channel ID")
	}

	params := defaultParams()
	options.Apply(params, opts)

	client := &Client{
		Client: *client.New(
			params.permitBlockEvents,
			dispatcher.New(context, channelID, params.connProvider, discoveryService, opts...),
			opts...,
		),
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
	logger.Debugf("sending seek request....\n")

	seekInfo, err := c.seekInfo()
	if err != nil {
		return err
	}

	errch := make(chan error)
	c.Submit(dispatcher.NewSeekEvent(seekInfo, errch))

	select {
	case err = <-errch:
	case <-time.After(c.respTimeout):
		err = errors.New("timeout waiting for deliver status response")
	}

	if err != nil {
		logger.Errorf("unable to send seek request: %s\n", err)
		return err
	}

	logger.Debugf("successfully sent seek\n")
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
	} else {
		// We haven't received any blocks yet. Just ask for the newest
		c.seekType = seek.Newest
	}
	return nil
}

func (c *Client) seekInfo() (*ab.SeekInfo, error) {
	c.RLock()
	defer c.RUnlock()

	switch c.seekType {
	case seek.Newest:
		return seek.InfoNewest(), nil
	case seek.Oldest:
		return seek.InfoOldest(), nil
	case seek.FromBlock:
		return seek.InfoFrom(c.fromBlock), nil
	default:
		return nil, errors.Errorf("unsupported seek type:[%s]", c.seekType)
	}
}
