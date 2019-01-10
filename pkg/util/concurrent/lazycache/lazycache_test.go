/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lazycache

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazyref"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/fabric-sdk-go/pkg/util/test"
)

func ExampleCache_MustGet() {
	cache := New("Example_Cache", func(key Key) (interface{}, error) {
		return fmt.Sprintf("Value_for_key_%s", key), nil
	})
	defer cache.Close()

	key := NewStringKey("Key1")

	fmt.Println(cache.MustGet(key))
	// Output: Value_for_key_Key1
}

func ExampleCache_Get() {
	cache := New("Example_Cache", func(key Key) (interface{}, error) {
		if key.String() == "error" {
			return nil, fmt.Errorf("some error")
		}
		return fmt.Sprintf("Value_for_key_%s", key), nil
	})
	defer cache.Close()

	value, err := cache.Get(NewStringKey("Key1"))
	if err != nil {
		fmt.Printf("Error returned: %s\n", err)
	}
	fmt.Println(value)

	value, err = cache.Get(NewStringKey("error"))
	if err != nil {
		fmt.Printf("Error returned: %s\n", err)
	}
	fmt.Println(value)
}

func ExampleCache_Get_expiring() {
	cache := New("Example_Expiring_Cache",
		func(key Key) (interface{}, error) {
			if key.String() == "error" {
				return nil, fmt.Errorf("some error")
			}
			return fmt.Sprintf("Value_for_key_%s", key), nil
		},
		lazyref.WithAbsoluteExpiration(time.Second),
		lazyref.WithFinalizer(func(expiredValue interface{}) {
			fmt.Printf("Expired value: %s\n", expiredValue)
		}),
	)
	defer cache.Close()

	value, err := cache.Get(NewStringKey("Key1"))
	if err != nil {
		fmt.Printf("Error returned: %s\n", err)
	} else {
		fmt.Print(value)
	}

	_, err = cache.Get(NewStringKey("error"))
	if err != nil {
		fmt.Printf("Error returned: %s\n", err)
	}
}

func ExampleCache_Get_expiringWithData() {
	cache := NewWithData("Example_Expiring_Cache",
		func(key Key, data interface{}) (interface{}, error) {
			return fmt.Sprintf("Value_for_%s_%d", key, data.(int)), nil
		},
		lazyref.WithAbsoluteExpiration(20*time.Millisecond),
	)
	defer cache.Close()

	for i := 0; i < 5; i++ {
		value, err := cache.Get(NewStringKey("Key"), i)
		if err != nil {
			fmt.Printf("Error returned: %s", err)
		} else {
			fmt.Print(value)
		}
		time.Sleep(15 * time.Millisecond)
	}
}

func TestGet(t *testing.T) {
	var numTimesInitialized int32
	expectedTimesInitialized := 2

	cache := New("Example_Cache", func(key Key) (interface{}, error) {
		if key.String() == "error" {
			return nil, fmt.Errorf("some error")
		}
		atomic.AddInt32(&numTimesInitialized, 1)
		return fmt.Sprintf("Value_for_key_%s", key), nil
	})

	concurrency := 100
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()

			value, err := cache.Get(NewStringKey("Key1"))
			if err != nil {
				test.Failf(t, "Error returned: %s", err)
			}
			expectedValue := "Value_for_key_Key1"
			if value != expectedValue {
				test.Failf(t, "Expecting value [%s] but got [%s]", expectedValue, value)
			}

			value, err = cache.Get(NewStringKey("Key2"))
			if err != nil {
				test.Failf(t, "Error returned: %s", err)
			}
			expectedValue = "Value_for_key_Key2"
			if value != expectedValue {
				test.Failf(t, "Expecting value [%s] but got [%s]", expectedValue, value)
			}

			_, err = cache.Get(NewStringKey("error"))
			if err == nil {
				test.Failf(t, "Expecting error but got none")
			}
		}()
	}

	wg.Wait()

	if num := atomic.LoadInt32(&numTimesInitialized); num != int32(expectedTimesInitialized) {
		t.Fatalf("Expecting initializer to be called %d time(s) but it was called %d time(s)", expectedTimesInitialized, num)
	}
	cache.Close()
}

func TestDelete(t *testing.T) {

	cache := New("Example_Cache", func(key Key) (interface{}, error) {
		if key.String() == "error" {
			return nil, fmt.Errorf("some error")
		}
		return fmt.Sprintf("Value_for_key_%s", key), nil
	})
	defer cache.Close()

	_, err := cache.Get(NewStringKey("Key1"))
	if err != nil {
		test.Failf(t, "Error returned: %s", err)
	}
	_, ok := cache.m.Load("Key1")
	if !ok {
		test.Failf(t, "value not exist in map")
	}

	cache.Delete(NewStringKey("Key1"))

	_, ok = cache.m.Load("Key1")
	if ok {
		test.Failf(t, "value exist in map after delete")
	}

}

func TestMustGetPanic(t *testing.T) {
	cache := New("Example_Cache", func(key Key) (interface{}, error) {
		if key.String() == "error" {
			return nil, fmt.Errorf("some error")
		}
		return fmt.Sprintf("Value_for_key_%s", key), nil
	})

	value := cache.MustGet(NewStringKey("Key1"))
	expectedValue := "Value_for_key_Key1"
	if value != expectedValue {
		t.Fatalf("Expecting value [%s] but got [%s]", expectedValue, value)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Expecting panic but got none")
		}
	}()

	cache.MustGet(NewStringKey("error"))
	t.Fatal("Expecting panic but got none")
	cache.Close()
}

type closableValue struct {
	str         string
	closeCalled int32
}

func (v *closableValue) Close() {
	atomic.StoreInt32(&v.closeCalled, 1)
}

func (v *closableValue) CloseCalled() bool {
	return atomic.LoadInt32(&v.closeCalled) == 1
}

func TestClose(t *testing.T) {
	cache := New("Example_Cache", func(key Key) (interface{}, error) {
		return &closableValue{
			str: fmt.Sprintf("Value_for_key_%s", key),
		}, nil
	})

	cval, err := cache.Get(NewStringKey("Key1"))
	if err != nil {
		t.Fatalf("Error returned: %s", err)
	}

	expectedValue := "Value_for_key_Key1"

	cvalue := cval.(*closableValue)
	if cvalue.str != expectedValue {
		t.Fatalf("Expecting value [%s] but got [%s]", expectedValue, cvalue.str)
	}
	if cvalue.CloseCalled() {
		t.Fatal("Not expecting close to be called but is was")
	}

	// Get again - should succeed
	cval, err = cache.Get(NewStringKey("Key1"))
	if err != nil {
		t.Fatalf("Error returned: %s", err)
	}
	cvalue = cval.(*closableValue)
	if cvalue.str != expectedValue {
		t.Fatalf("Expecting value [%s] but got [%s]", expectedValue, cvalue.str)
	}
	if cvalue.CloseCalled() {
		t.Fatal("Not expecting close to be called but is was")
	}

	// Close the cache
	cache.Close()

	assert.True(t, cache.IsClosed())

	// Close again should be fine
	cache.Close()

	if !cvalue.CloseCalled() {
		t.Fatal("Expecting close to be called but is wasn't")
	}

	_, err = cache.Get(NewStringKey("Key1"))
	if err == nil {
		t.Fatal("Expecting error since cache is closed")
	}
}

// TestGetExpiring tests that the cache value expires and that the
// finalizer is called when it expires and also after the cache is closed.
func TestGetExpiring(t *testing.T) {
	var numTimesInitialized int32
	var numTimesFinalized int32

	cache := New("Example_Expiring_Cache",
		func(key Key) (interface{}, error) {
			if key.String() == "error" {
				return nil, fmt.Errorf("some error")
			}
			atomic.AddInt32(&numTimesInitialized, 1)
			return fmt.Sprintf("Value_for_key_%s", key), nil
		},
		lazyref.WithAbsoluteExpiration(25*time.Millisecond),
		lazyref.WithFinalizer(func(expiredValue interface{}) {
			atomic.AddInt32(&numTimesFinalized, 1)
		}),
	)

	for i := 0; i < 10; i++ {
		time.Sleep(10 * time.Millisecond)
		value, err := cache.Get(NewStringKey("Key1"))
		require.NoErrorf(t, err, "error returned for Key1")
		expectedValue := "Value_for_key_Key1"
		assert.Equal(t, expectedValue, value)

		value, err = cache.Get(NewStringKey("Key2"))
		require.NoErrorf(t, err, "error returned for Key2")
		expectedValue = "Value_for_key_Key2"
		assert.Equal(t, expectedValue, value)

		_, err = cache.Get(NewStringKey("error"))
		require.Errorf(t, err, "expecting error 'error' key")
	}

	initializedTimes := atomic.LoadInt32(&numTimesInitialized)
	assert.Truef(t, initializedTimes > 2, "Expecting initializer to be called more than %d times but it was called %d time(s)", 2, initializedTimes)
	finalizedTimes := atomic.LoadInt32(&numTimesFinalized)
	assert.Truef(t, finalizedTimes > 2, "Expecting finalizer to be called more than %d times but it was called %d time(s)", 2, finalizedTimes)

	// Closing the cache should also close all of the lazy refs and call our finalizer
	cache.Close()

	finalizedTimesAfterClose := atomic.LoadInt32(&numTimesFinalized)
	assert.Truef(t, finalizedTimesAfterClose == initializedTimes, "Expecting finalizer to be called %d times but it was called %d time(s)", initializedTimes, finalizedTimesAfterClose)
}

// TestGetExpiringWithData tests that the data passed to Get() is
// used in the initializer each time the value expires.
func TestGetExpiringWithData(t *testing.T) {
	var numTimesInitialized int32
	var numTimesFinalized int32

	cache := NewWithData("Example_Expiring_Cache",
		func(key Key, data interface{}) (interface{}, error) {
			atomic.AddInt32(&numTimesInitialized, 1)
			return fmt.Sprintf("Value_for_key_%s_[%d]", key, data.(int)), nil
		},
		lazyref.WithAbsoluteExpiration(25*time.Millisecond),
		lazyref.WithFinalizer(func(expiredValue interface{}) {
			atomic.AddInt32(&numTimesFinalized, 1)
		}),
	)
	defer cache.Close()

	numTimesIndexChanged := 0
	prevIndex := 0
	for i := 0; i < 10; i++ {
		time.Sleep(10 * time.Millisecond)
		value, err := cache.Get(NewStringKey("Key"), i)
		require.NoError(t, err)

		strValue := value.(string)
		i := strings.Index(strValue, "[")
		assert.Truef(t, i > 0, "expecting to find [ in value")
		j := strings.Index(strValue, "]")
		assert.Truef(t, j > 0, "expecting to find ] in value")

		index, err := strconv.Atoi(strValue[i+1 : j])
		require.NoError(t, err)

		assert.Truef(t, index <= i, "expecting index to be less than or equal to i")
		if index != prevIndex {
			numTimesIndexChanged++
			prevIndex = index
		}
	}
	assert.Truef(t, numTimesIndexChanged > 2, "expecting that the index would change at least 2 times but it changed %d tim(s)", numTimesIndexChanged)
}

// TestGetExpiringError tests that the lazy reference value is NOT cached if
// an error is returned from the initializer.
func TestGetExpiringError(t *testing.T) {
	var numTimesInitialized int32
	var numTimesFinalized int32

	cache := New("Example_Expiring_Cache",
		func(key Key) (interface{}, error) {
			atomic.AddInt32(&numTimesInitialized, 1)
			return nil, fmt.Errorf("some error")
		},
		lazyref.WithFinalizer(func(expiredValue interface{}) {
			atomic.AddInt32(&numTimesFinalized, 1)
		}),
	)

	iterations := 10
	for i := 0; i < iterations; i++ {
		time.Sleep(10 * time.Millisecond)
		_, err := cache.Get(NewStringKey("error"))
		require.Errorf(t, err, "expecting error 'error' key")
	}

	initializedTimes := atomic.LoadInt32(&numTimesInitialized)
	assert.Equalf(t, int32(iterations), initializedTimes, "Expecting initializer to be called every time since no value should have been cached when returning an error")
	finalizedTimes := atomic.LoadInt32(&numTimesFinalized)
	assert.Equalf(t, int32(0), finalizedTimes, "Expecting finalizer not to be called due to error but it was called %d time(s)", finalizedTimes)

	cache.Close()

	finalizedTimesAfterClose := atomic.LoadInt32(&numTimesFinalized)
	assert.Equalf(t, int32(0), finalizedTimesAfterClose, "Expecting finalizer not to be called due to error but it was called %d time(s)", finalizedTimesAfterClose)
}
