/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lazycache

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

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
