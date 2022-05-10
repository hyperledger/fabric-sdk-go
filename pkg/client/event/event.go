/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package event enables access to a channel events on a Fabric network. Event client receives events such as block, filtered block,
// chaincode, and transaction status events.
//  Basic Flow:
//  1) Prepare channel client context
//  2) Create event client
//  3) Register for events
//  4) Process events (or timeout)
//  5) Unregister
package event

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/dispatcher"
	"github.com/pkg/errors"
)

// Client enables access to a channel events on a Fabric network.
type Client struct {
	eventService         fab.EventService
	permitBlockEvents    bool
	fromBlock            uint64
	seekType             seek.Type
	chaincodeID          string
	eventConsumerTimeout *time.Duration
}

// New returns a Client instance. Client receives events such as block, filtered block,
// chaincode, and transaction status events.
// nolint: gocyclo
func New(channelProvider context.ChannelProvider, opts ...ClientOption) (*Client, error) {

	channelContext, err := channelProvider()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create channel context")
	}

	eventClient := Client{}

	for _, param := range opts {
		err1 := param(&eventClient)
		if err1 != nil {
			return nil, errors.WithMessage(err, "option failed")
		}
	}

	if channelContext.ChannelService() == nil {
		return nil, errors.New("channel service not initialized")
	}

	var es fab.EventService
	if eventClient.permitBlockEvents {
		var opts []options.Opt
		opts = append(opts, client.WithBlockEvents())
		if eventClient.seekType != "" {
			opts = append(opts, deliverclient.WithSeekType(eventClient.seekType))
			if eventClient.seekType == seek.FromBlock {
				opts = append(opts, deliverclient.WithBlockNum(eventClient.fromBlock))
			}
		}
		if eventClient.chaincodeID != "" {
			opts = append(opts, deliverclient.WithChaincodeID(eventClient.chaincodeID))
		}
		if eventClient.eventConsumerTimeout != nil {
			opts = append(opts, dispatcher.WithEventConsumerTimeout(*eventClient.eventConsumerTimeout))
		}
		es, err = channelContext.ChannelService().EventService(opts...)
	} else {
		es, err = channelContext.ChannelService().EventService()
	}

	if err != nil {
		return nil, errors.WithMessage(err, "event service creation failed")
	}

	eventClient.eventService = es

	return &eventClient, nil
}

// RegisterBlockEvent registers for block events. If the caller does not have permission
// to register for block events then an error is returned. Unregister must be called when the registration is no longer needed.
//  Parameters:
//  filter is an optional filter that filters out unwanted events. (Note: Only one filter may be specified.)
//
//  Returns:
//  the registration and a channel that is used to receive events. The channel is closed when Unregister is called.
func (c *Client) RegisterBlockEvent(filter ...fab.BlockFilter) (fab.Registration, <-chan *fab.BlockEvent, error) {
	return c.eventService.RegisterBlockEvent(filter...)
}

// RegisterFilteredBlockEvent registers for filtered block events. Unregister must be called when the registration is no longer needed.
//  Returns:
//  the registration and a channel that is used to receive events. The channel is closed when Unregister is called.
func (c *Client) RegisterFilteredBlockEvent() (fab.Registration, <-chan *fab.FilteredBlockEvent, error) {
	return c.eventService.RegisterFilteredBlockEvent()
}

// RegisterChaincodeEvent registers for chaincode events. Unregister must be called when the registration is no longer needed.
//  Parameters:
//  ccID is the chaincode ID for which events are to be received
//  eventFilter is the chaincode event filter (regular expression) for which events are to be received
//
//  Returns:
//  the registration and a channel that is used to receive events. The channel is closed when Unregister is called.
func (c *Client) RegisterChaincodeEvent(ccID, eventFilter string) (fab.Registration, <-chan *fab.CCEvent, error) {
	return c.eventService.RegisterChaincodeEvent(ccID, eventFilter)
}

// RegisterTxStatusEvent registers for transaction status events. Unregister must be called when the registration is no longer needed.
//  Parameters:
//  txID is the transaction ID for which events are to be received
//
//  Returns:
//  the registration and a channel that is used to receive events. The channel is closed when Unregister is called.
func (c *Client) RegisterTxStatusEvent(txID string) (fab.Registration, <-chan *fab.TxStatusEvent, error) {
	return c.eventService.RegisterTxStatusEvent(txID)
}

// Unregister removes the given registration and closes the event channel.
//  Parameters:
//  reg is the registration handle that was returned from one of the Register functions
func (c *Client) Unregister(reg fab.Registration) {
	c.eventService.Unregister(reg)
}
