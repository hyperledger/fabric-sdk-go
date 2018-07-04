/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package futurevalue

import (
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Initializer initializes the value
type Initializer func() (interface{}, error)

// valueHolder holds the actual value
type valueHolder struct {
	value interface{}
	err   error
}

// Value implements a Future Value in which a reference is initialized once
// (and only once) using the Initialize function. Only one Go routine can call
// Initialize whereas multiple Go routines may invoke Get, and will wait
// until the reference has been initialized.
// Regardless of whether Initialize returns success or error,
// the value cannot be initialized again.
type Value struct {
	sync.RWMutex
	ref         unsafe.Pointer
	initializer Initializer
}

// New returns a new future value
func New(initializer Initializer) *Value {
	f := &Value{
		initializer: initializer,
	}
	f.Lock()
	return f
}

// Initialize initializes the future value.
// This function must be called only once. Subsequent
// calls may result in deadlock.
func (f *Value) Initialize() (interface{}, error) {
	value, err := f.initializer()
	f.set(value, err)
	f.Unlock()

	return value, err
}

// Get returns the value and/or error that occurred during initialization.
func (f *Value) Get() (interface{}, error) {
	// Try outside of a lock
	if ok, value, err := f.get(); ok {
		return value, err
	}

	f.RLock()
	defer f.RUnlock()

	_, value, err := f.get()
	return value, err
}

// MustGet returns the value. If an error resulted
// during initialization then this function will panic.
func (f *Value) MustGet() interface{} {
	value, err := f.Get()
	if err != nil {
		panic(fmt.Sprintf("get returned error: %s", err))
	}
	return value
}

// IsSet returns true if the value has been set, otherwise false is returned
func (f *Value) IsSet() bool {
	p := atomic.LoadPointer(&f.ref)
	return p != nil
}

func (f *Value) get() (bool, interface{}, error) {
	p := atomic.LoadPointer(&f.ref)
	if p == nil {
		return false, nil, nil
	}
	holder := (*valueHolder)(p)
	return true, holder.value, holder.err
}

func (f *Value) set(value interface{}, err error) {
	holder := &valueHolder{
		value: value,
		err:   err,
	}
	atomic.StorePointer(&f.ref, unsafe.Pointer(holder)) //nolint
}
