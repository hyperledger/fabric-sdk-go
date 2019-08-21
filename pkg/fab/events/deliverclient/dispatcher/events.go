/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	ab "github.com/hyperledger/fabric-protos-go/orderer"
)

// SeekEvent is a SeekInfo request to the deliver server
type SeekEvent struct {
	SeekInfo *ab.SeekInfo
	ErrCh    chan<- error
}

// NewSeekEvent returns a new SeekRequestEvent
func NewSeekEvent(seekInfo *ab.SeekInfo, errch chan<- error) *SeekEvent {
	return &SeekEvent{
		SeekInfo: seekInfo,
		ErrCh:    errch,
	}
}
