/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package util

import (
	"math/rand"
	"reflect"
	"sync"
)

// Equals returns whether a and b are the same
type Equals func(a interface{}, b interface{}) bool

var viperLock sync.RWMutex

// IndexInSlice returns the index of given object o in array, and -1 if it is not in array.
func IndexInSlice(array interface{}, o interface{}, equals Equals) int {
	arr := reflect.ValueOf(array)
	for i := 0; i < arr.Len(); i++ {
		if equals(arr.Index(i).Interface(), o) {
			return i
		}
	}
	return -1
}

// GetRandomIndices returns indiceCount random indices
// from 0 to highestIndex.
func GetRandomIndices(indiceCount, highestIndex int) []int {
	// More choices needed than possible to choose.
	if highestIndex+1 < indiceCount {
		return nil
	}

	return rand.Perm(highestIndex + 1)[:indiceCount]
}

// Set is a generic and thread-safe
// set container
type Set struct {
	items map[interface{}]struct{}
	lock  *sync.RWMutex
}

// RandomInt returns, as an int, a non-negative pseudo-random integer in [0,n)
// It panics if n <= 0
func RandomInt(n int) int {
	return rand.Intn(n)
}

// RandomUInt64 returns a random uint64
//
// If we want a rand that's non-global and specific to gossip, we can
// establish one. Otherwise this uses the process-global locking RNG.
func RandomUInt64() uint64 {
	return rand.Uint64()
}
