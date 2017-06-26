/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"encoding/pem"
)

// Peer ...
/**
 * The Peer class represents a peer in the target blockchain network to which
 * HFC sends endorsement proposals, transaction ordering or query requests.
 *
 * The Peer class represents the remote Peer node and its network membership materials,
 * aka the ECert used to verify signatures. Peer membership represents organizations,
 * unlike User membership which represents individuals.
 *
 * When constructed, a Peer instance can be designated as an event source, in which case
 * a “eventSourceUrl” attribute should be configured. This allows the SDK to automatically
 * attach transaction event listeners to the event stream.
 *
 * It should be noted that Peer event streams function at the Peer level and not at the
 * channel and chaincode levels.
 */
type Peer interface {
	ConnectEventSource()
	IsEventListened(event string, channel Channel) (bool, error)
	AddListener(eventType string, eventTypeData interface{}, eventCallback interface{}) (string, error)
	RemoveListener(eventListenerRef string) (bool, error)
	Name() string
	SetName(name string)
	Roles() []string
	SetRoles(roles []string)
	EnrollmentCertificate() *pem.Block
	SetEnrollmentCertificate(pem *pem.Block)
	URL() string
	SendProposal(proposal *TransactionProposal) (*TransactionProposalResponse, error)
}
