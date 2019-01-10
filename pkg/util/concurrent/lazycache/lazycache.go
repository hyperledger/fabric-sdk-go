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
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/futurevalue"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazyref"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/util")

// Key holds the string key for the cache entry
type Key interface {
	String() string
}

// EntryInitializer creates a cache value for the given key
type EntryInitializer func(key Key) (interface{}, error)

// EntryInitializerWithData creates a cache value for the given key and the
// additional data passed in from Get(). With expiring cache entries, the
// initializer is called with the same key, but the latest data is passed from
// the Get() call that triggered the data to be cached/re-cached.
type EntryInitializerWithData func(key Key, data interface{}) (interface{}, error)

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
	initializer EntryInitializerWithData
	closed      int32
	useRef      bool
}

// New creates a new lazy cache.
// - name is the name of the cache and is only used for debugging purpose
// - initializer is invoked the first time an entry is being cached
// - opts are options for the cache. If any lazyref option is passed then a lazy reference
//   is created for each of the cache entries to hold the actual value. This makes it possible
//   to have expiring values and values that proactively refresh.
func New(name string, initializer EntryInitializer, opts ...options.Opt) *Cache {
	return NewWithData(name,
		func(key Key, data interface{}) (interface{}, error) {
			return initializer(key)
		},
		opts...,
	)
}

// NewWithData creates a new lazy cache. The provided initializer accepts optional data that
// is passed in from Get().
// - name is the name of the cache and is only used for debugging purpose
// - initializer is invoked the first time an entry is being cached
// - opts are options for the cache. If any lazyref option is passed then a lazy reference
//   is created for each of the cache entries to hold the actual value. This makes it possible
//   to have expiring values and values that proactively refresh.
func NewWithData(name string, initializer EntryInitializerWithData, opts ...options.Opt) *Cache {
	useRef := useLazyRef(opts...)
	if useRef {
		initializer = newLazyRefInitializer(name, initializer, opts...)
	}
	return &Cache{
		name:        name,
		initializer: initializer,
		useRef:      useRef,
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
func (c *Cache) Get(key Key, data ...interface{}) (interface{}, error) {
	keyStr := key.String()

	f, ok := c.m.Load(keyStr)
	if ok {
		v, err := f.(future).Get()
		if err != nil {
			return nil, err
		}
		return c.value(v, first(data))
	}

	// The key wasn't found. Attempt to add one.
	newFuture := futurevalue.New(
		func() (interface{}, error) {
			if closed := atomic.LoadInt32(&c.closed); closed == 1 {
				return nil, errors.Errorf("%s - cache is closed", c.name)
			}
			return c.initializer(key, first(data))
		},
	)

	f, loaded := c.m.LoadOrStore(keyStr, newFuture)
	if loaded {
		// Another thread has added the key before us. Return the value.
		v, err := f.(future).Get()
		if err != nil {
			return nil, err
		}
		return c.value(v, first(data))
	}

	// We added the key. It must be initialized.
	value, err := newFuture.Initialize()
	if err != nil {
		// Failed. Delete the key.
		logger.Debugf("%s - Failed to initialize key [%s]: %s. Deleting key.", c.name, keyStr, err)
		c.m.Delete(keyStr)
		return nil, err
	}
	return c.value(value, first(data))
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
	c.DeleteAll()
}

// IsClosed reeturns true if the cache has been closed
func (c *Cache) IsClosed() bool {
	return atomic.LoadInt32(&c.closed) == 1
}

// DeleteAll does the following:
// - calls Close on all values that implement a Close() function
// - deletes all entries from the cache
func (c *Cache) DeleteAll() {
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

// Delete does the following:
// - calls Close on all values that implement a Close() function
// - deletes key from the cache
func (c *Cache) Delete(key Key) {
	logger.Debugf("%s - Deleting cache key", key.String())
	value, ok := c.m.Load(key.String())
	if ok {
		c.close(key.String(), value.(future))
		c.m.Delete(key.String())
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

func newLazyRefInitializer(name string, initializer EntryInitializerWithData, opts ...options.Opt) EntryInitializerWithData {
	return func(key Key, data interface{}) (interface{}, error) {
		logger.Debugf("%s - Calling initializer for [%s], data [%#v]", name, key, data)
		ref := lazyref.NewWithData(
			func(data interface{}) (interface{}, error) {
				logger.Debugf("%s - Calling lazyref initializer for [%s], data [%#v]", name, key, data)
				return initializer(key, data)
			},
			opts...,
		)

		// Make sure no error is returned from lazyref.Get(). If there is
		// then return the error. We don't want to cache a reference that always
		// returns an error, especially if it's a refreshing reference.
		_, err := ref.Get(data)
		if err != nil {
			logger.Debugf("%s - Error returned from lazyref initializer [%s], data [%#v]: %s", name, key, data, err)
			ref.Close()
			return nil, err
		}
		logger.Debugf("%s - Returning lazyref for [%s], data [%#v]", name, key, data)
		return ref, nil
	}
}

func (c *Cache) value(value interface{}, data interface{}) (interface{}, error) {
	if value != nil && c.useRef {
		return value.(*lazyref.Reference).Get(data)
	}
	return value, nil
}

func first(data []interface{}) interface{} {
	if len(data) == 0 {
		return nil
	}
	return data[0]
}

// useLazyRef returns true if the cache should used lazy references to hold the actual value
func useLazyRef(opts ...options.Opt) bool {
	chk := &refOptCheck{}
	options.Apply(chk, opts)
	return chk.useRef
}
