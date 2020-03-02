/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

type list struct {
	None       CommitHandlerFactory
	OrgAll     CommitHandlerFactory
	OrgAny     CommitHandlerFactory
	NetworkAll CommitHandlerFactory
	NetworkAny CommitHandlerFactory
}

// DefaultCommitHandlers provides the built-in commit handler implementations.
var DefaultCommitHandlers = &list{
	None:       nil,
	OrgAll:     orgAll,
	OrgAny:     orgAny,
	NetworkAll: networkAll,
	NetworkAny: networkAny,
}

type commithandler struct {
	transactionID string
	network       Network
}

func (ch *commithandler) StartListening() {
}

func (ch *commithandler) WaitForEvents(timeout int64) {
}

func (ch *commithandler) CancelListening() {
}

type commithandlerfactory struct {
}

func (chf *commithandlerfactory) Create(txid string, network Network) CommitHandler {
	return &commithandler{
		transactionID: txid,
		network:       network,
	}
}

var orgAll = &commithandlerfactory{}
var orgAny = &commithandlerfactory{}
var networkAll = &commithandlerfactory{}
var networkAny = &commithandlerfactory{}
