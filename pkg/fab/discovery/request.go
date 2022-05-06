/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discovery

import (
	"github.com/hyperledger/fabric-protos-go/peer"
	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
)

// Request aggregates several queries inside it
type Request struct {
	r *discclient.Request
}

// NewRequest creates a new request
func NewRequest() *Request {
	return &Request{discclient.NewRequest()}
}

// AddConfigQuery adds to the request a config query
func (req *Request) AddConfigQuery() *Request {
	req.r.AddConfigQuery()
	return req
}

// AddLocalPeersQuery adds to the request a local peer query
func (req *Request) AddLocalPeersQuery() *Request {
	req.r.AddLocalPeersQuery()
	return req
}

// OfChannel sets the next queries added to be in the given channel's context
func (req *Request) OfChannel(ch string) *Request {
	req.r.OfChannel(ch)
	return req
}

// AddEndorsersQuery adds to the request a query for given chaincodes
// interests are the chaincode interests that the client wants to query for.
// All interests for a given channel should be supplied in an aggregated slice
func (req *Request) AddEndorsersQuery(interests ...*peer.ChaincodeInterest) (*Request, error) {
	_, err := req.r.AddEndorsersQuery(interests...)
	return req, err
}

// AddPeersQuery adds to the request a peer query
func (req *Request) AddPeersQuery(invocationChain ...*peer.ChaincodeCall) *Request {
	req.r.AddPeersQuery(invocationChain...)
	return req
}

// CcCalls creates an array of ChaincodeCalls based of cc names, can be used in AddPeersQuery(CcCalls(...))
func CcCalls(ccNames ...string) []*peer.ChaincodeCall {
	var call []*peer.ChaincodeCall

	for _, ccName := range ccNames {
		call = append(call, &peer.ChaincodeCall{
			Name: ccName,
		})
	}

	return call
}

// CcInterests creates an array of ChaincodeInterests based of ChaincodeCalls, can be used in AddEndorsersQuery(CcInterests(CcCalls(...)))
func CcInterests(invocationsChains ...[]*peer.ChaincodeCall) []*peer.ChaincodeInterest {
	var interests []*peer.ChaincodeInterest

	for _, invocationChain := range invocationsChains {
		interests = append(interests, &peer.ChaincodeInterest{
			Chaincodes: invocationChain,
		})
	}

	return interests
}
