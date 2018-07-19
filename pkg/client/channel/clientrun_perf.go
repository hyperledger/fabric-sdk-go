// +build pprof

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	// TODO update metrics package with Fabric's copy, once officially released
	// TODO and pinned into Go SDK with the below commented out import statement
	//"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/metrics"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel/invoke"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/greylist"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"

	// TODO remove below metrics declaration once Fabric's copy is ready to be used
	"github.com/hyperledger/fabric-sdk-go/test/performance/metrics"
	"github.com/uber-go/tally"
)

type clientTally struct {
	queryCount     tally.Counter
	queryFailCount tally.Counter
	queryTimer     tally.Timer

	executeCount     tally.Counter
	executeFailCount tally.Counter
	executeTimer     tally.Timer
}

func newClientTally(channelContext context.Channel) clientTally {
	return clientTally{
		queryCount:     metrics.RootScope.SubScope(channelContext.ChannelID()).Counter("ch_client_query_calls"),
		queryFailCount: metrics.RootScope.SubScope(channelContext.ChannelID()).Counter("ch_client_query_errors"),
		queryTimer:     metrics.RootScope.SubScope(channelContext.ChannelID()).Timer("ch_client_query_processing_time_seconds"),

		executeCount:     metrics.RootScope.SubScope(channelContext.ChannelID()).Counter("ch_client_execute_calls"),
		executeFailCount: metrics.RootScope.SubScope(channelContext.ChannelID()).Counter("ch_client_execute_errors"),
		executeTimer:     metrics.RootScope.SubScope(channelContext.ChannelID()).Timer("ch_client_execute_processing_time_seconds"),
	}
}

func newClient(channelContext context.Channel, membership fab.ChannelMembership, eventService fab.EventService, greylistProvider *greylist.Filter) Client {
	ct := newClientTally(channelContext)

	channelClient := Client{
		membership:   membership,
		eventService: eventService,
		greylist:     greylistProvider,
		context:      channelContext,
		clientTally:  ct,
	}
	return channelClient
}

func callQuery(cc *Client, request Request, options ...RequestOption) (Response, error) {
	cc.executeCount.Inc(1)
	stopWatch := cc.queryTimer.Start()
	defer stopWatch.Stop()

	r, err := cc.InvokeHandler(invoke.NewQueryHandler(), request, options...)
	if err != nil {
		cc.queryFailCount.Inc(1)
	}
	return r, err
}

func callExecute(cc *Client, request Request, options ...RequestOption) (Response, error) {
	cc.executeCount.Inc(1)
	stopWatch := cc.executeTimer.Start()
	defer stopWatch.Stop()

	r, err := cc.InvokeHandler(invoke.NewExecuteHandler(), request, options...)
	if err != nil {
		cc.executeFailCount.Inc(1)
	}
	return r, err
}
