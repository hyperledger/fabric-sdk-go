/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package status

import (
	"strconv"

	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	grpcCodes "google.golang.org/grpc/codes"
)

// Code represents a status code
type Code uint32

const (
	// OK is returned on success.
	OK Code = 0

	// Unknown represents status codes that are uncategorized or unknown to the SDK
	Unknown Code = 1

	// ConnectionFailed is returned when a network connection attempt from the SDK fails
	ConnectionFailed Code = 2

	// EndorsementMismatch is returned when there is a mismatch in endorsements received by the SDK
	EndorsementMismatch Code = 3

	// EmptyCert is return when an empty cert is returned
	EmptyCert Code = 4

	// Timeout operation timed out
	Timeout Code = 5

	// NoPeersFound No peers were discovered/configured
	NoPeersFound Code = 6

	// MultipleErrors multiple errors occurred
	MultipleErrors Code = 7

	// SignatureVerificationFailed is when signature fails verification
	SignatureVerificationFailed Code = 8

	// MissingEndorsement is if an endorsement is missing
	MissingEndorsement Code = 9

	// QueryEndorsers error indicates that no endorser group was found that would
	// satisfy the chaincode policy
	QueryEndorsers Code = 11

	// GenericTransient is generally used by tests to indicate that a retry is possible
	GenericTransient Code = 12

	// PrematureChaincodeExecution indicates that an attempt was made to invoke a chaincode that's
	// in the process of being launched.
	PrematureChaincodeExecution Code = 21

	// ChaincodeAlreadyLaunching indicates that an attempt for multiple simultaneous invokes was made to launch chaincode
	ChaincodeAlreadyLaunching Code = 22

	// ChaincodeNameNotFound indicates that an that an attempt was made to invoke a chaincode that's not yet initialized
	ChaincodeNameNotFound Code = 23
)

// CodeName maps the codes in this packages to human-readable strings
var CodeName = map[int32]string{
	0:  "OK",
	1:  "UNKNOWN",
	2:  "CONNECTION_FAILED",
	3:  "ENDORSEMENT_MISMATCH",
	4:  "EMPTY_CERT",
	5:  "TIMEOUT",
	6:  "NO_PEERS_FOUND",
	7:  "MULTIPLE_ERRORS",
	8:  "SIGNATURE_VERIFICATION_FAILED",
	9:  "MISSING_ENDORSEMENT",
	11: "QUERY_ENDORSERS",
	12: "GENERIC_TRANSIENT",
	21: "PREMATURE_CHAINCODE_EXECUTION",
	22: "CHAINCODE_ALREADY_LAUNCHING",
	23: "CHAINCODE_NAME_NOT_FOUND",
}

// ToInt32 cast to int32
func (c Code) ToInt32() int32 {
	return int32(c)
}

// String representation of the code
func (c Code) String() string {
	if s, ok := CodeName[c.ToInt32()]; ok {
		return s
	}
	return strconv.Itoa(int(c))
}

// ToSDKStatusCode cast to fabric-sdk-go status code
func ToSDKStatusCode(c int32) Code {
	return Code(c)
}

// ToGRPCStatusCode cast to gRPC status code
func ToGRPCStatusCode(c int32) grpcCodes.Code {
	return grpcCodes.Code(c)
}

// ToPeerStatusCode cast to peer status
func ToPeerStatusCode(c int32) common.Status {
	return ToFabricCommonStatusCode(c)
}

// ToOrdererStatusCode cast to peer status
func ToOrdererStatusCode(c int32) common.Status {
	return ToFabricCommonStatusCode(c)
}

// ToFabricCommonStatusCode cast to common.Status
func ToFabricCommonStatusCode(c int32) common.Status {
	return common.Status(c)
}

// ToTransactionValidationCode cast to transaction validation status code
func ToTransactionValidationCode(c int32) pb.TxValidationCode {
	return pb.TxValidationCode(c)
}
