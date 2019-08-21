/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package status defines metadata for errors returned by fabric-sdk-go. This
// information may be used by SDK users to make decisions about how to handle
// certain error conditions.
// Status codes are divided by group, where each group represents a particular
// component and the codes correspond to those returned by the component.
// These are defined in detail below.
package status

import (
	"fmt"

	"github.com/pkg/errors"

	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/multi"
	grpcstatus "google.golang.org/grpc/status"
)

// Status provides additional information about an unsuccessful operation
// performed by fabric-sdk-go. Essentially, this object contains metadata about
// an error returned by the SDK.
type Status struct {
	// Group status group
	Group Group
	// Code status code
	Code int32
	// Message status message
	Message string
	// Details any additional status details
	Details []interface{}
}

// Group of status to help users infer status codes from various components
type Group int32

const (
	// UnknownStatus unknown status group
	UnknownStatus Group = iota

	// TransportStatus defines the status returned by the transport layer of
	// the connections made by fabric-sdk-go

	// GRPCTransportStatus is the status associated with requests made over
	// gRPC connections
	GRPCTransportStatus
	// HTTPTransportStatus is the status associated with requests made over HTTP
	// connections
	HTTPTransportStatus

	// ServerStatus defines the status returned by various servers that fabric-sdk-go
	// is a client to

	// EndorserServerStatus status returned by the endorser server
	EndorserServerStatus
	// EventServerStatus status returned by the event service
	EventServerStatus
	// OrdererServerStatus status returned by the ordering service
	OrdererServerStatus
	// FabricCAServerStatus status returned by the Fabric CA server
	FabricCAServerStatus

	// ClientStatus defines the status from responses inferred by fabric-sdk-go.
	// This could be a result of response validation performed by the SDK - for example,
	// a client status could be produced by validating endorsements

	// EndorserClientStatus status returned from the endorser client
	EndorserClientStatus
	// OrdererClientStatus status returned from the orderer client
	OrdererClientStatus
	// ClientStatus is a generic client status
	ClientStatus

	// ChaincodeStatus defines the status codes returned by chaincode
	ChaincodeStatus

	// DiscoveryServerStatus status returned by the Discovery Server
	DiscoveryServerStatus

	// TestStatus is used by tests to create retry codes.
	TestStatus
)

// GroupName maps the groups in this packages to human-readable strings
var GroupName = map[int32]string{
	0:  "Unknown",
	1:  "gRPC Transport Status",
	2:  "HTTP Transport Status",
	3:  "Endorser Server Status",
	4:  "Event Server Status",
	5:  "Orderer Server Status",
	6:  "Fabric CA Server Status",
	7:  "Endorser Client Status",
	8:  "Orderer Client Status",
	9:  "Client Status",
	10: "Chaincode status",
	11: "Discovery status",
	12: "Test status",
}

func (g Group) String() string {
	if s, ok := GroupName[int32(g)]; ok {
		return s
	}
	return UnknownStatus.String()
}

// FromError returns a Status representing err if available,
// otherwise it returns nil, false.
func FromError(err error) (s *Status, ok bool) {
	if err == nil {
		return &Status{Code: int32(OK)}, true
	}
	if s, ok := err.(*Status); ok {
		return s, true
	}
	unwrappedErr := errors.Cause(err)
	if s, ok := unwrappedErr.(*Status); ok {
		return s, true
	}
	if m, ok := unwrappedErr.(multi.Errors); ok {
		// Return all of the errors in the details
		var errors []interface{}
		for _, err := range m {
			errors = append(errors, err)
		}
		return New(ClientStatus, MultipleErrors.ToInt32(), m.Error(), errors), true
	}

	return nil, false
}

func (s *Status) Error() string {
	return fmt.Sprintf("%s Code: (%d) %s. Description: %s", s.Group.String(), s.Code, s.codeString(), s.Message)
}

func (s *Status) codeString() string {
	switch s.Group {
	case GRPCTransportStatus:
		return ToGRPCStatusCode(s.Code).String()
	case EndorserServerStatus, OrdererServerStatus:
		return ToFabricCommonStatusCode(s.Code).String()
	case EventServerStatus:
		return ToTransactionValidationCode(s.Code).String()
	case EndorserClientStatus, OrdererClientStatus, ClientStatus:
		return ToSDKStatusCode(s.Code).String()
	default:
		return Unknown.String()
	}
}

// New returns a Status with the given parameters
func New(group Group, code int32, msg string, details []interface{}) *Status {
	return &Status{Group: group, Code: code, Message: msg, Details: details}
}

// NewFromProposalResponse creates a status created from the given ProposalResponse
func NewFromProposalResponse(res *pb.ProposalResponse, endorser string) *Status {
	if res == nil {
		return nil
	}
	details := []interface{}{endorser, res.Response.Payload}

	return New(EndorserServerStatus, res.Response.Status, res.Response.Message, details)
}

// NewFromGRPCStatus new Status from gRPC status response
func NewFromGRPCStatus(s *grpcstatus.Status) *Status {
	if s == nil {
		return nil
	}
	details := make([]interface{}, len(s.Proto().Details))
	for i, detail := range s.Proto().Details {
		details[i] = detail
	}

	return &Status{Group: GRPCTransportStatus, Code: s.Proto().Code,
		Message: s.Message(), Details: details}
}

// NewFromExtractedChaincodeError returns Status when a chaincode error occurs
func NewFromExtractedChaincodeError(code int, message string) *Status {
	return &Status{Group: ChaincodeStatus, Code: int32(code),
		Message: message, Details: nil}
}
