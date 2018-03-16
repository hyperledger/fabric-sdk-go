/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lazyref

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
)

var logger = logging.NewLogger("fabsdk/util")

// Initializer is a function that initializes the value
type Initializer func() (interface{}, error)

// Finalizer is a function that is called when the reference
// is closed
type Finalizer func()

// ExpirationProvider is a function that returns the
// expiration time of a reference
type ExpirationProvider func() time.Duration

// valueHolder holds the actual value
type valueHolder struct {
	value interface{}
}

// expirationHandler is invoked when the
// reference expires
type expirationHandler func()

// ExpirationType indicates how to handle expiration of the reference
type ExpirationType uint

const (
	// LastAccessed specifies that the expiration time is calculated
	// from the last access time
	LastAccessed ExpirationType = iota

	// LastInitialized specifies that the expiration time is calculated
	// from the time the reference was initialized
	LastInitialized
)

// Reference holds a value that is initialized on first access using the provided
// Initializer function. The Reference has an optional expiring feature
// wherin the value is reset after the provided period of time. A subsequent call
// to Get or MustGet causes the Initializer function to be invoked again.
// The Reference also has a proactive refresh capability, in which the Initializer
// function is periodically called out of band (out of band means that the caller
// of Get or MustGet does not need to wait for the initializer function to complete:
// the old value will be used until the new value has finished initializing).
// An optional Finalizer function may be provided to be invoked whenever the Reference
// is closed (via a call to Close) or if it expires. (Note: The Finalizer function
// is not called every time the value is refreshed with the periodic refresh feature.)
type Reference struct {
	sync.RWMutex
	ref                unsafe.Pointer
	lastTimeAccessed   unsafe.Pointer
	initializer        Initializer
	finalizer          Finalizer
	expirationHandler  expirationHandler
	expirationProvider ExpirationProvider
	initialInit        time.Duration
	expiryType         ExpirationType
	closed             chan bool
}

// New creates a new reference
func New(initializer Initializer, opts ...Opt) *Reference {
	lazyRef := &Reference{
		initializer: initializer,
		initialInit: InitOnFirstAccess,
		closed:      make(chan bool, 1),
	}

	for _, opt := range opts {
		opt(lazyRef)
	}

	if lazyRef.expirationProvider != nil {
		// This is an expiring reference. After the initializer is
		// called, set a timer that will call the expiration handler.
		initializer := lazyRef.initializer
		lazyRef.initializer = func() (interface{}, error) {
			value, err := initializer()
			if err == nil {
				lazyRef.startTimer(lazyRef.expirationProvider())
			}
			return value, err
		}
		if lazyRef.expirationHandler == nil {
			// Set a default expiration handler
			lazyRef.expirationHandler = lazyRef.resetValue
		}
		if lazyRef.initialInit >= 0 {
			lazyRef.startTimer(lazyRef.initialInit)
		}
	}

	return lazyRef
}

// Get returns the value, or an error if the initialiser returned an error.
func (r *Reference) Get() (interface{}, error) {
	// Try outside of a lock
	if value, ok := r.get(); ok {
		return value, nil
	}

	r.Lock()
	defer r.Unlock()

	// Try again inside the lock
	if value, ok := r.get(); ok {
		return value, nil
	}

	// Value hasn't been set yet

	value, err := r.initializer()
	if err != nil {
		return nil, err
	}
	r.set(value)

	return value, nil
}

// MustGet returns the value. If an error is returned
// during initialization of the value then this function
// will panic.
func (r *Reference) MustGet() interface{} {
	value, err := r.Get()
	if err != nil {
		panic(fmt.Sprintf("error returned from Get: %s", err))
	}
	return value
}

// Close ensures that the finalizer (if provided) is called.
// Close should be called for expiring references and
// rerences that specify finalizers.
func (r *Reference) Close() {
	r.Lock()
	defer r.Unlock()

	logger.Debug("Closing reference")

	if r.expirationHandler != nil {
		r.closed <- true
	}
	if r.finalizer != nil {
		r.finalizer()
	}
}

func (r *Reference) get() (interface{}, bool) {
	r.setLastAccessed()
	p := atomic.LoadPointer(&r.ref)
	if p == nil {
		return nil, false
	}
	return (*valueHolder)(p).value, true
}

func (r *Reference) set(value interface{}) {
	atomic.StorePointer(&r.ref, unsafe.Pointer(&valueHolder{value: value}))
}

func (r *Reference) setLastAccessed() {
	now := time.Now()
	atomic.StorePointer(&r.lastTimeAccessed, unsafe.Pointer(&now))
}

func (r *Reference) lastAccessed() time.Time {
	p := atomic.LoadPointer(&r.lastTimeAccessed)
	return *(*time.Time)(p)
}

func (r *Reference) startTimer(expiration time.Duration) {
	r.setLastAccessed()

	go func() {
		expiry := expiration
		for {
			select {
			case <-r.closed:
			case <-time.After(expiry):
				if r.expiryType == LastInitialized {
					r.handleExpiration()
					return
				}

				// Check how long it's been since last access
				durSinceLastAccess := time.Now().Sub(r.lastAccessed())
				if durSinceLastAccess > expiration {
					r.handleExpiration()
					return
				}
				// Set another expiry for the remainder of the time
				expiry = expiration - durSinceLastAccess
			}
		}
	}()
}

func (r *Reference) handleExpiration() {
	r.Lock()
	defer r.Unlock()

	logger.Debug("Invoking expiration handler")
	r.expirationHandler()
}

// resetValue is an expiration handler that calls the
// finalizer and resets the reference to nil.
// Note: This function is invoked from inside a write
// lock so there's no need to lock
func (r *Reference) resetValue() {
	if r.finalizer != nil {
		r.finalizer()
	}
	atomic.StorePointer(&r.ref, nil)
}

// refreshValue is an expiration handler that calls the
// initializer and, if the initializer was successful, resets
// the reference with the new value.
// Note: This function is invoked from inside a write
// lock so there's no need to lock
func (r *Reference) refreshValue() {
	if value, err := r.initializer(); err != nil {
		expiration := r.expirationProvider()
		logger.Warnf("Error - initializer returned error: %s. Will retry in %s", err, expiration)
		// Start the timer so that we can retry
		r.startTimer(expiration)
	} else {
		r.set(value)
	}
}
