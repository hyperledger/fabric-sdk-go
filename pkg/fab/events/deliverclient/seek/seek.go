/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package seek

import (
	"math"

	ab "github.com/hyperledger/fabric-protos-go/orderer"
)

// Type is the type of Seek request to perform.
type Type string

const (
	// Oldest seeks from the first block
	Oldest = "oldest"
	// Newest seeks from the last block
	Newest = "newest"
	// FromBlock seeks from a specific block
	FromBlock = "from"
)

var (
	oldestPos = &ab.SeekPosition{Type: &ab.SeekPosition_Oldest{Oldest: &ab.SeekOldest{}}}
	newestPos = &ab.SeekPosition{Type: &ab.SeekPosition_Newest{Newest: &ab.SeekNewest{}}}
	maxPos    = &ab.SeekPosition{Type: &ab.SeekPosition_Specified{Specified: &ab.SeekSpecified{Number: math.MaxUint64}}}
)

// InfoOldest returns a SeekInfo struct that indicates to the deliver server
// that we want all blocks starting from the oldest block (block 0)
func InfoOldest() *ab.SeekInfo {
	return newSeekInfo(oldestPos, maxPos)
}

// InfoNewest returns a SeekInfo struct that indicates to the deliver server
// that we just want the latest blocks
func InfoNewest() *ab.SeekInfo {
	return newSeekInfo(newestPos, maxPos)
}

// InfoFrom returns a SeekInfo struct that indicates to the deliver server
// that we want all blocks starting from the given block number
func InfoFrom(fromBlock uint64) *ab.SeekInfo {
	return newSeekInfo(seekFromPos(fromBlock), maxPos)
}

func seekFromPos(fromBlock uint64) *ab.SeekPosition {
	return &ab.SeekPosition{
		Type: &ab.SeekPosition_Specified{
			Specified: &ab.SeekSpecified{
				Number: fromBlock,
			},
		},
	}
}

func newSeekInfo(start *ab.SeekPosition, stop *ab.SeekPosition) *ab.SeekInfo {
	return &ab.SeekInfo{
		Start:    start,
		Stop:     stop,
		Behavior: ab.SeekInfo_BLOCK_UNTIL_READY,
	}
}
