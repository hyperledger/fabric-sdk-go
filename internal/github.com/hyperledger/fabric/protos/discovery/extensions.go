/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discovery

// QueryType defines the types of service discovery requests
type QueryType uint8

const (
	InvalidQueryType QueryType = iota
	ConfigQueryType
	PeerMembershipQueryType
	ChaincodeQueryType
	LocalMembershipQueryType
)

// ConfigAt returns the ConfigResult at a given index in the Response,
// or an Error if present.
func (m *Response) ConfigAt(i int) (*ConfigResult, *Error) {
	r := m.Results[i]
	return r.GetConfigResult(), r.GetError()
}

// MembershipAt returns the PeerMembershipResult at a given index in the Response,
// or an Error if present.
func (m *Response) MembershipAt(i int) (*PeerMembershipResult, *Error) {
	r := m.Results[i]
	return r.GetMembers(), r.GetError()
}

// EndorsersAt returns the PeerMembershipResult at a given index in the Response,
// or an Error if present.
func (m *Response) EndorsersAt(i int) (*ChaincodeQueryResult, *Error) {
	r := m.Results[i]
	return r.GetCcQueryRes(), r.GetError()
}
