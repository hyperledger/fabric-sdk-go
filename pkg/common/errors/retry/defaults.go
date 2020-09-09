/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package retry

import (
	"time"

	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	grpcCodes "google.golang.org/grpc/codes"
)

const (
	// DefaultAttempts number of retry attempts made by default
	DefaultAttempts = 3
	// DefaultInitialBackoff default initial backoff
	DefaultInitialBackoff = 500 * time.Millisecond
	// DefaultMaxBackoff default maximum backoff
	DefaultMaxBackoff = 60 * time.Second
	// DefaultBackoffFactor default backoff factor
	DefaultBackoffFactor = 2.0
)

// Resource Management Suggested Defaults
const (
	// ResMgmtDefaultAttempts number of retry attempts made by default
	ResMgmtDefaultAttempts = 5
	// ResMgmtDefaultInitialBackoff default initial backoff
	ResMgmtDefaultInitialBackoff = time.Second
	// ResMgmtDefaultMaxBackoff default maximum backoff
	ResMgmtDefaultMaxBackoff = 60 * time.Second
	// ResMgmtDefaultBackoffFactor default backoff factor
	ResMgmtDefaultBackoffFactor = 2.5
)

// DefaultOpts default retry options
var DefaultOpts = Opts{
	Attempts:       DefaultAttempts,
	InitialBackoff: DefaultInitialBackoff,
	MaxBackoff:     DefaultMaxBackoff,
	BackoffFactor:  DefaultBackoffFactor,
	RetryableCodes: DefaultRetryableCodes,
}

// DefaultChannelOpts default retry options for the channel client
var DefaultChannelOpts = Opts{
	Attempts:       DefaultAttempts,
	InitialBackoff: DefaultInitialBackoff,
	MaxBackoff:     DefaultMaxBackoff,
	BackoffFactor:  DefaultBackoffFactor,
	RetryableCodes: ChannelClientRetryableCodes,
}

// DefaultResMgmtOpts default retry options for the resource management client
var DefaultResMgmtOpts = Opts{
	Attempts:       ResMgmtDefaultAttempts,
	InitialBackoff: ResMgmtDefaultInitialBackoff,
	MaxBackoff:     ResMgmtDefaultMaxBackoff,
	BackoffFactor:  ResMgmtDefaultBackoffFactor,
	RetryableCodes: ResMgmtDefaultRetryableCodes,
}

// DefaultRetryableCodes these are the error codes, grouped by source of error,
// that are considered to be transient error conditions by default
var DefaultRetryableCodes = map[status.Group][]status.Code{
	status.EndorserClientStatus: {
		status.EndorsementMismatch,
		status.ChaincodeNameNotFound,
	},
	status.EndorserServerStatus: {
		status.Code(common.Status_SERVICE_UNAVAILABLE),
		status.Code(common.Status_INTERNAL_SERVER_ERROR),
	},
	status.OrdererServerStatus: {
		status.Code(common.Status_SERVICE_UNAVAILABLE),
		status.Code(common.Status_INTERNAL_SERVER_ERROR),
	},
	status.EventServerStatus: {
		status.Code(pb.TxValidationCode_DUPLICATE_TXID),
		status.Code(pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE),
		status.Code(pb.TxValidationCode_MVCC_READ_CONFLICT),
		status.Code(pb.TxValidationCode_PHANTOM_READ_CONFLICT),
	},
	// TODO: gRPC introduced retries in v1.8.0. This can be replaced with the
	// gRPC fail fast option, once available
	status.GRPCTransportStatus: {
		status.Code(grpcCodes.Unavailable),
	},
}

// ResMgmtDefaultRetryableCodes are the suggested codes that should be treated as
// transient by fabric-sdk-go/pkg/client/resmgmt.Client
var ResMgmtDefaultRetryableCodes = map[status.Group][]status.Code{
	status.EndorserClientStatus: {
		status.ConnectionFailed,
		status.EndorsementMismatch,
		status.ChaincodeNameNotFound,
	},
	status.EndorserServerStatus: {
		status.Code(common.Status_SERVICE_UNAVAILABLE),
		status.Code(common.Status_INTERNAL_SERVER_ERROR),
	},
	status.OrdererServerStatus: {
		status.Code(common.Status_SERVICE_UNAVAILABLE),
		status.Code(common.Status_INTERNAL_SERVER_ERROR),
		status.Code(common.Status_BAD_REQUEST),
		status.Code(common.Status_NOT_FOUND),
	},
	status.EventServerStatus: {
		status.Code(pb.TxValidationCode_DUPLICATE_TXID),
		status.Code(pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE),
		status.Code(pb.TxValidationCode_MVCC_READ_CONFLICT),
		status.Code(pb.TxValidationCode_PHANTOM_READ_CONFLICT),
	},
	// TODO: gRPC introduced retries in v1.8.0. This can be replaced with the
	// gRPC fail fast option, once available
	status.GRPCTransportStatus: {
		status.Code(grpcCodes.Unavailable),
	},
}

// ChannelClientRetryableCodes are the suggested codes that should be treated as
// transient by fabric-sdk-go/pkg/client/channel.Client
var ChannelClientRetryableCodes = map[status.Group][]status.Code{
	status.EndorserClientStatus: {
		status.ConnectionFailed, status.EndorsementMismatch,
		status.Code(pb.TxValidationCode_MVCC_READ_CONFLICT),
		status.ChaincodeNameNotFound,
	},
	status.EndorserServerStatus: {
		status.Code(common.Status_SERVICE_UNAVAILABLE),
		status.Code(common.Status_INTERNAL_SERVER_ERROR),
		status.PvtDataDisseminationFailed,
	},
	status.OrdererClientStatus: {
		status.ConnectionFailed,
	},
	status.OrdererServerStatus: {
		status.Code(common.Status_SERVICE_UNAVAILABLE),
		status.Code(common.Status_INTERNAL_SERVER_ERROR),
	},
	status.EventServerStatus: {
		status.Code(pb.TxValidationCode_DUPLICATE_TXID),
		status.Code(pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE),
		status.Code(pb.TxValidationCode_MVCC_READ_CONFLICT),
		status.Code(pb.TxValidationCode_PHANTOM_READ_CONFLICT),
	},
	// TODO: gRPC introduced retries in v1.8.0. This can be replaced with the
	// gRPC fail fast option, once available
	status.GRPCTransportStatus: {
		status.Code(grpcCodes.Unavailable),
	},
}

// ChannelConfigRetryableCodes error codes to be taken into account for query channel config retry
var ChannelConfigRetryableCodes = map[status.Group][]status.Code{
	status.EndorserClientStatus: {status.EndorsementMismatch},
}

// TestRetryableCodes are used by tests to determine error situations that can be retried.
var TestRetryableCodes = map[status.Group][]status.Code{
	status.TestStatus: {
		status.GenericTransient,
	},
	status.DiscoveryServerStatus: {
		status.QueryEndorsers,
		status.Code(pb.TxValidationCode_MVCC_READ_CONFLICT),
	},
	status.EndorserClientStatus: {
		status.ConnectionFailed, status.EndorsementMismatch,
		status.ChaincodeNameNotFound,
		status.Code(pb.TxValidationCode_MVCC_READ_CONFLICT),
	},
	status.EndorserServerStatus: {
		status.Code(common.Status_SERVICE_UNAVAILABLE),
		status.Code(common.Status_INTERNAL_SERVER_ERROR),
		status.Code(pb.TxValidationCode_MVCC_READ_CONFLICT),
	},
	status.OrdererClientStatus: {
		status.ConnectionFailed,
	},
	status.OrdererServerStatus: {
		status.Code(common.Status_SERVICE_UNAVAILABLE),
		status.Code(common.Status_INTERNAL_SERVER_ERROR),
	},
	status.EventServerStatus: {
		status.Code(pb.TxValidationCode_DUPLICATE_TXID),
		status.Code(pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE),
		status.Code(pb.TxValidationCode_MVCC_READ_CONFLICT),
		status.Code(pb.TxValidationCode_PHANTOM_READ_CONFLICT),
	},
	// TODO: gRPC introduced retries in v1.8.0. This can be replaced with the
	// gRPC fail fast option, once available
	status.GRPCTransportStatus: {
		status.Code(grpcCodes.Unavailable),
	},
}

const (
	// TestAttempts number of retry attempts made by default
	TestAttempts = 10
	// TestInitialBackoff default initial backoff
	TestInitialBackoff = 200 * time.Millisecond
	// TestMaxBackoff default maximum backoff
	TestMaxBackoff = 50 * time.Second
	// TestBackoffFactor default backoff factor
	TestBackoffFactor = 1.75
)

// TestRetryOpts are used by tests to determine retry parameters.
var TestRetryOpts = Opts{
	Attempts:       TestAttempts,
	InitialBackoff: TestInitialBackoff,
	MaxBackoff:     TestMaxBackoff,
	BackoffFactor:  TestBackoffFactor,
	RetryableCodes: TestRetryableCodes,
}
