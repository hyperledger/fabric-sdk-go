/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package retry

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
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

// DefaultOpts default retry options
var DefaultOpts = Opts{
	Attempts:       DefaultAttempts,
	InitialBackoff: DefaultInitialBackoff,
	MaxBackoff:     DefaultMaxBackoff,
	BackoffFactor:  DefaultBackoffFactor,
	RetryableCodes: DefaultRetryableCodes,
}

// DefaultRetryableCodes these are the error codes, grouped by source of error,
// that are considered to be transient error conditions by default
var DefaultRetryableCodes = map[status.Group][]status.Code{
	status.EndorserClientStatus: []status.Code{
		status.EndorsementMismatch,
	},
	status.EndorserServerStatus: []status.Code{
		status.Code(common.Status_SERVICE_UNAVAILABLE),
		status.Code(common.Status_INTERNAL_SERVER_ERROR),
	},
	status.OrdererServerStatus: []status.Code{
		status.Code(common.Status_SERVICE_UNAVAILABLE),
		status.Code(common.Status_INTERNAL_SERVER_ERROR),
	},
	status.EventServerStatus: []status.Code{
		status.Code(pb.TxValidationCode_DUPLICATE_TXID),
		status.Code(pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE),
		status.Code(pb.TxValidationCode_MVCC_READ_CONFLICT),
		status.Code(pb.TxValidationCode_PHANTOM_READ_CONFLICT),
	},
	// TODO: gRPC introduced retries in v1.8.0. This can be replaced with the
	// gRPC fail fast option, once available
	status.GRPCTransportStatus: []status.Code{
		status.Code(grpcCodes.Unavailable),
	},
}

// ChannelClientRetryableCodes are the suggested codes that should be treated as
// transient by fabric-sdk-go/api/apitxn.ChannelClient
var ChannelClientRetryableCodes = map[status.Group][]status.Code{
	status.EndorserClientStatus: []status.Code{
		status.ConnectionFailed, status.EndorsementMismatch,
	},
	status.EndorserServerStatus: []status.Code{
		status.Code(common.Status_SERVICE_UNAVAILABLE),
		status.Code(common.Status_INTERNAL_SERVER_ERROR),
	},
	status.OrdererClientStatus: []status.Code{
		status.ConnectionFailed,
	},
	status.OrdererServerStatus: []status.Code{
		status.Code(common.Status_SERVICE_UNAVAILABLE),
		status.Code(common.Status_INTERNAL_SERVER_ERROR),
	},
	status.EventServerStatus: []status.Code{
		status.Code(pb.TxValidationCode_DUPLICATE_TXID),
		status.Code(pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE),
		status.Code(pb.TxValidationCode_MVCC_READ_CONFLICT),
		status.Code(pb.TxValidationCode_PHANTOM_READ_CONFLICT),
	},
	// TODO: gRPC introduced retries in v1.8.0. This can be replaced with the
	// gRPC fail fast option, once available
	status.GRPCTransportStatus: []status.Code{
		status.Code(grpcCodes.Unavailable),
	},
}
