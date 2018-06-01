/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package eventhubclient

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/endpoint"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/eventhubclient/connection"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/eventhubclient/dispatcher"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/fab")

var ehConnProvider = func(context context.Client, chConfig fab.ChannelCfg, peer fab.Peer) (api.Connection, error) {
	eventEndpoint, ok := peer.(api.EventEndpoint)
	if !ok {
		panic("peer is not an EventEndpoint")
	}
	return connection.New(context, chConfig, eventEndpoint.EventURL(), eventEndpoint.Opts()...)
}

// Client connects to a peer and receives channel events, such as bock, filtered block, chaincode, and transaction status events.
type Client struct {
	client.Client
	params
}

// New returns a new event hub client
func New(context context.Client, chConfig fab.ChannelCfg, discovery fab.DiscoveryService, opts ...options.Opt) (*Client, error) {
	params := defaultParams()

	// FIXME: Temporarily set the default to block events since Fabric 1.0 does
	// not support filtered block events
	opts = append(opts, client.WithBlockEvents())

	options.Apply(params, opts)

	// Use a custom Discovery Service which wraps the given discovery service
	// and produces event endpoints containing the event URL and
	// additional GRPC options.
	discoveryWrapper, err := endpoint.NewEndpointDiscoveryWrapper(
		context, chConfig.ID(), discovery,
		endpoint.WithTargetFilter(newMSPFilter(context.Identifier().MSPID)))
	if err != nil {
		return nil, err
	}

	client := &Client{
		Client: *client.New(
			dispatcher.New(context, chConfig, discoveryWrapper, params.connProvider, opts...),
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
	logger.Debug("sending register interests request....")

	errch := make(chan error)
	if err := c.Submit(dispatcher.NewRegisterInterestsEvent(c.interests, errch)); err != nil {
		logger.Errorf("unable to submit new register interests events: %s", err)
		return err
	}

	var err error
	select {
	case err = <-errch:
	case <-time.After(c.respTimeout):
		err = errors.New("timeout waiting for register interests response")
	}

	if err != nil {
		logger.Errorf("unable to send register interests request: %s", err)
		return err
	}

	logger.Debug("successfully sent register interests")
	return nil
}

type mspFilter struct {
	mspID string
}

func newMSPFilter(mspID string) *mspFilter {
	return &mspFilter{mspID: mspID}
}

func (f *mspFilter) Accept(peer fab.Peer) bool {
	return peer.MSPID() == f.mspID
}
