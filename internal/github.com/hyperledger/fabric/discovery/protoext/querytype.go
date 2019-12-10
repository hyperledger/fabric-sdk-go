/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package protoext

import "github.com/hyperledger/fabric/protos/discovery"

// QueryType defines the types of service discovery requests
type QueryType uint8

const (
	InvalidQueryType QueryType = iota
	ConfigQueryType
	PeerMembershipQueryType
	ChaincodeQueryType
	LocalMembershipQueryType
)

// GetType returns the type of the request
func GetQueryType(q *discovery.Query) QueryType {
	switch {
	case q.GetCcQuery() != nil:
		return ChaincodeQueryType
	case q.GetConfigQuery() != nil:
		return ConfigQueryType
	case q.GetPeerQuery() != nil:
		return PeerMembershipQueryType
	case q.GetLocalPeers() != nil:
		return LocalMembershipQueryType
	default:
		return InvalidQueryType
	}
}
