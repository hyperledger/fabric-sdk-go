/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chclient

import (
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/retry"
)

//Handler for chaining transaction executions
type Handler interface {
	Handle(context *RequestContext, clientContext *ClientContext)
}

//ClientContext contains context parameters for handler execution
type ClientContext struct {
	CryptoSuite apicryptosuite.CryptoSuite
	Channel     apifabclient.Channel
	Discovery   apifabclient.DiscoveryService
	Selection   apifabclient.SelectionService
	EventHub    apifabclient.EventHub
}

//RequestContext contains request, opts, response parameters for handler execution
type RequestContext struct {
	Request      Request
	Opts         Opts
	Response     Response
	Error        error
	RetryHandler retry.Handler
}
