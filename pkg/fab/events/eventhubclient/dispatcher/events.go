/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// RegisterInterestsEvent registers interests with the event hub
type RegisterInterestsEvent struct {
	Interests []*pb.Interest
	ErrCh     chan<- error
}

// NewRegisterInterestsEvent returns a RegisterInterests event
func NewRegisterInterestsEvent(interests []*pb.Interest, errch chan<- error) *RegisterInterestsEvent {
	return &RegisterInterestsEvent{
		Interests: interests,
		ErrCh:     errch,
	}
}

// UnregisterInterestsEvent unregisters interests with the event hub
type UnregisterInterestsEvent struct {
	Interests []*pb.Interest
	ErrCh     chan<- error
}

// NewUnregisterInterestsEvent returns an UnregisterInterests event
func NewUnregisterInterestsEvent(interests []*pb.Interest, errch chan<- error) *UnregisterInterestsEvent {
	return &UnregisterInterestsEvent{
		Interests: interests,
		ErrCh:     errch,
	}
}
