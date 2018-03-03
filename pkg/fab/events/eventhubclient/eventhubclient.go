/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package eventhubclient

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/context"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/eventhubclient/connection"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/eventhubclient/dispatcher"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/options"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabric_sdk_go")

var ehConnProvider = func(channelID string, context context.Client, peer fab.Peer) (api.Connection, error) {
	eventEndpoint, ok := peer.(api.EventEndpoint)
	if !ok {
		panic("peer is not an EventEndpoint")
	}

	return connection.New(
		context, channelID, eventEndpoint.EventURL(),
	)
}

// Client connects to a peer and receives channel events, such as bock, filtered block, chaincode, and transaction status events.
type Client struct {
	client.Client
	params
}

// New returns a new event hub client
func New(context context.Client, channelID string, discoveryService fab.DiscoveryService, opts ...options.Opt) (*Client, error) {
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
	client.SetAfterConnectHandler(client.registerInterests)

	if err := client.Start(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) registerInterests() error {
	logger.Debugf("sending register interests request....\n")

	errch := make(chan error)
	c.Submit(dispatcher.NewRegisterInterestsEvent(c.interests, errch))

	var err error
	select {
	case err = <-errch:
	case <-time.After(c.respTimeout):
		err = errors.New("timeout waiting for register interests response")
	}

	if err != nil {
		logger.Errorf("unable to send register interests request: %s\n", err)
		return err
	}

	logger.Debugf("successfully sent register interests\n")
	return nil
}
