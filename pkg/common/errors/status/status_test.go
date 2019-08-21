/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package status

import (
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/multi"
	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

func TestStatusConstructors(t *testing.T) {
	s := New(EndorserClientStatus, ConnectionFailed.ToInt32(), "test", nil)
	assert.NotNil(t, s, "Expected status to be constructed")
	assert.EqualValues(t, ConnectionFailed, ToSDKStatusCode(s.Code))
	assert.Equal(t, EndorserClientStatus, s.Group)
	assert.Equal(t, "test", s.Message, "Expected test message")

	s = NewFromGRPCStatus(nil)
	assert.Nil(t, s)
	s = NewFromGRPCStatus(grpcstatus.New(grpccodes.DeadlineExceeded, "test"))
	assert.NotNil(t, s, "Expected status to be constructed")
	assert.EqualValues(t, grpccodes.DeadlineExceeded, ToGRPCStatusCode(s.Code))
	assert.Equal(t, GRPCTransportStatus, s.Group)
	assert.Equal(t, "test", s.Message, "Expected test message")

	s = NewFromProposalResponse(nil, "")
	assert.Nil(t, s)
	s = NewFromProposalResponse(&pb.ProposalResponse{
		Response: &pb.Response{
			Status:  int32(common.Status_BAD_REQUEST),
			Message: "test",
		}}, "localhost")
	assert.NotNil(t, s, "Expected status to be constructed")
	assert.EqualValues(t, common.Status_BAD_REQUEST, ToPeerStatusCode(s.Code))
	assert.Equal(t, EndorserServerStatus, s.Group)
	assert.Equal(t, "test", s.Message, "Expected test message")
	assert.Equal(t, "localhost", s.Details[0].(string))
}

func TestFromError(t *testing.T) {
	s := New(EndorserClientStatus, ConnectionFailed.ToInt32(), "test", nil)
	derivedStatus, ok := FromError(s)
	assert.True(t, ok)
	assert.Equal(t, s, derivedStatus)

	// Test unwrap
	s1 := errors.Wrap(s, "test")
	derivedStatus, ok = FromError(s1)
	assert.True(t, ok)
	assert.Equal(t, s, derivedStatus)

	s, ok = FromError(nil)
	assert.True(t, ok)
	assert.EqualValues(t, OK.ToInt32(), s.Code)

	_, ok = FromError(fmt.Errorf("Test"))
	assert.False(t, ok)

	errs := multi.Errors{}
	errs = append(errs, fmt.Errorf("Test"))
	s, ok = FromError(errs)
	assert.True(t, ok)
	assert.Equal(t, ClientStatus, s.Group)
	assert.EqualValues(t, MultipleErrors.ToInt32(), s.Code)
	assert.Equal(t, errs.Error(), s.Message)
}

func TestStatusToError(t *testing.T) {
	s := New(EndorserClientStatus, ConnectionFailed.ToInt32(), "test", nil)
	assert.Equal(t, "Endorser Client Status Code: (2) CONNECTION_FAILED. Description: test", s.Error())
}

func TestStatuCodeConversion(t *testing.T) {
	c := ToOrdererStatusCode(int32(common.Status_FORBIDDEN))
	assert.EqualValues(t, c, common.Status_FORBIDDEN)

	c1 := ToTransactionValidationCode(int32(pb.TxValidationCode_BAD_COMMON_HEADER))
	assert.EqualValues(t, c1, pb.TxValidationCode_BAD_COMMON_HEADER)

	s := OK.String()
	assert.Equal(t, CodeName[OK.ToInt32()], s)

	invalidCode25999 := Code(25999)
	assert.Equal(t, "25999", invalidCode25999.String())
}

func TestStatusCodeString(t *testing.T) {
	s := Status{Group: GRPCTransportStatus, Code: int32(grpccodes.Aborted)}
	assert.Equal(t, grpccodes.Aborted.String(), s.codeString())

	s = Status{Group: OrdererServerStatus, Code: int32(common.Status_BAD_REQUEST)}
	assert.Equal(t, common.Status_BAD_REQUEST.String(), s.codeString())

	s = Status{Group: EndorserClientStatus, Code: int32(OK)}
	assert.Equal(t, OK.String(), s.codeString())

	s = Status{Group: EventServerStatus, Code: int32(pb.TxValidationCode_BAD_CHANNEL_HEADER)}
	assert.Equal(t, pb.TxValidationCode_BAD_CHANNEL_HEADER.String(), s.codeString())

	unknownCode45779 := 45779
	s = Status{Code: int32(unknownCode45779)}
	assert.Equal(t, Unknown.String(), s.codeString())
}

func TestStatusGroupString(t *testing.T) {
	unknownGroup77377 := Group(73777)
	assert.Equal(t, UnknownStatus.String(), unknownGroup77377.String())
}

func TestChaincodeStatus(t *testing.T) {
	s := NewFromExtractedChaincodeError(500, "key not found")
	assert.Equal(t, "key not found", s.Message)
	assert.Equal(t, int32(500), s.Code)
}
