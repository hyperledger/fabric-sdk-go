/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lazycache

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/futurevalue"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/util")

// Key holds the string key for the cache entry
type Key interface {
	String() string
}

// EntryInitializer creates a cache value for the given key
type EntryInitializer func(key Key) (interface{}, error)

type future interface {
	Get() (interface{}, error)
	MustGet() interface{}
	IsSet() bool
}

type closable interface {
	Close()
}

// Cache implements a lazy initializing cache. A cache entry is created
// the first time a value is accessed (via Get or MustGet) by invoking
// the provided Initializer. If the Initializer returns an error then the
// entry will not be added.
type Cache struct {
	// name is useful for debugging
	name        string
	m           sync.Map
	initializer EntryInitializer
	closed      int32
}

// New creates a new lazy cache with the given name
// (Note that the name is only used for debugging purpose)
func New(name string, initializer EntryInitializer) *Cache {
	return &Cache{
		name:        name,
		initializer: initializer,
	}
}

// Name returns the name of the cache (useful for debugging)
func (c *Cache) Name() string {
	return c.name
}

// Get returns the value for the given key. If the
// key doesn't exist then the initializer is invoked
// to create the value, and the key is inserted. If the
// initializer returns an error then the key is removed
// from the cache.
func (c *Cache) Get(key Key) (interface{}, error) {
	keyStr := key.String()

	f, ok := c.m.Load(keyStr)
	if ok {
		return f.(future).Get()
	}

	// The key wasn't found. Attempt to add one.
	newFuture := futurevalue.New(
		func() (interface{}, error) {
			if closed := atomic.LoadInt32(&c.closed); closed == 1 {
				return nil, errors.Errorf("%s - cache is closed", c.name)
			}
			return c.initializer(key)
		},
	)

	f, loaded := c.m.LoadOrStore(keyStr, newFuture)
	if loaded {
		// Another thread has added the key before us. Return the value.
		return f.(future).Get()
	}

	// We added the key. It must be initailized.
	value, err := newFuture.Initialize()
	if err != nil {
		// Failed. Delete the key.
		logger.Debugf("%s - Failed to initialize key [%s]: %s. Deleting key.", c.name, keyStr, err)
		c.m.Delete(keyStr)
	}
	return value, err
}

// MustGet returns the value for the given key. If the key doesn't
// exist then the initializer is invoked to create the value and the
// key is inserted. If an error is returned during initialization of the
// value then this function will panic.
func (c *Cache) MustGet(key Key) interface{} {
	value, err := c.Get(key)
	if err != nil {
		panic(fmt.Sprintf("error returned from Get: %s", err))
	}
	return value
}

// Close does the following:
// - calls Close on all values that implement a Close() function
// - deletes all entries from the cache
// - prevents further calls to the cache
func (c *Cache) Close() {
	if !atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		// Already closed
		return
	}

	logger.Debugf("%s - Closing cache", c.name)

	var keys []interface{}
	c.m.Range(func(key interface{}, value interface{}) bool {
		c.close(key.(string), value.(future))
		keys = append(keys, key)
		return true
	})

	for _, key := range keys {
		c.m.Delete(key)
	}
}

func (c *Cache) close(key string, f future) {
	if !f.IsSet() {
		logger.Debugf("%s - Reference for [%q] is not set", c.name, key)
		return
	}
	value, err := f.Get()
	if err == nil && value != nil {
		if clos, ok := value.(closable); ok && c != nil {
			logger.Debugf("%s - Invoking Close on value for key [%q].", c.name, key)
			clos.Close()
		}
	}
}
