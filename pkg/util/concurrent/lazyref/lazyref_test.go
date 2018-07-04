/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lazyref

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func ExampleReference() {
	ref := New(func() (interface{}, error) {
		return "Value1", nil
	})
	fmt.Println(ref.MustGet())
	// Output: Value1
}

func ExampleReference_expiring() {
	sequence := 0
	ref := New(
		func() (interface{}, error) {
			sequence++
			return fmt.Sprintf("Data_%d", sequence), nil
		},
		WithIdleExpiration(200*time.Millisecond),
	)

	for i := 0; i < 5; i++ {
		fmt.Println(ref.MustGet())
		time.Sleep(100 * time.Millisecond)
	}
}

// This example demonstrates a refreshing reference.
// The reference is initialized immediately after creation
// and every 2 seconds thereafter.
func ExampleReference_refreshing() {
	sequence := 0
	ref := New(
		func() (interface{}, error) {
			sequence++
			return fmt.Sprintf("Data_%d", sequence), nil
		},
		WithRefreshInterval(InitImmediately, 200*time.Millisecond),
	)

	for i := 0; i < 5; i++ {
		fmt.Println(ref.MustGet())
		time.Sleep(300 * time.Millisecond)
	}
}

func TestGet(t *testing.T) {
	var numTimesInitialized int32
	expectedTimesInitialized := 1
	concurrency := 100
	expectedValue := "Data1"

	ref := New(func() (interface{}, error) {
		atomic.AddInt32(&numTimesInitialized, 1)
		return expectedValue, nil
	})

	var wg sync.WaitGroup
	wg.Add(concurrency)

	var errors []error
	var mutex sync.Mutex

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			value, err := ref.Get()
			if err != nil {
				panic(err.Error())
			}
			if value != expectedValue {
				mutex.Lock()
				errors = append(errors, fmt.Errorf("expecting value to be %s but got %s", expectedValue, value))
				mutex.Unlock()
			}
		}()
	}

	wg.Wait()

	if len(errors) > 0 {
		t.Fatal(errors[0].Error())
	}
	if num := atomic.LoadInt32(&numTimesInitialized); num != int32(expectedTimesInitialized) {
		t.Fatalf("expecting initializer to be called %d time(s) but was called %d time(s)", expectedTimesInitialized, num)
	}
}

func TestMustGet(t *testing.T) {
	var numTimesInitialized int32
	expectedTimesInitialized := 1
	concurrency := 100
	expectedValue := "Data1"

	ref := New(func() (interface{}, error) {
		atomic.AddInt32(&numTimesInitialized, 1)
		t.Log("Initializing Reference...\n")
		return expectedValue, nil
	})

	var wg sync.WaitGroup
	wg.Add(concurrency)

	var errors []error
	var mutex sync.Mutex

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			value := ref.MustGet().(string)
			if value != expectedValue {
				mutex.Lock()
				errors = append(errors, fmt.Errorf("expecting value to be %s but got %s", expectedValue, value))
				mutex.Unlock()
			}
		}()
	}

	wg.Wait()

	if num := atomic.LoadInt32(&numTimesInitialized); num != int32(expectedTimesInitialized) {
		t.Fatalf("expecting initializer to be called %d time(s) but was called %d time(s)", expectedTimesInitialized, num)
	}
	if len(errors) > 0 {
		t.Fatalf(errors[0].Error())
	}
}

func TestGetWithData(t *testing.T) {
	var numTimesInitialized int32
	expectedTimesInitialized := 1
	concurrency := 100

	ref := NewWithData(func(data interface{}) (interface{}, error) {
		atomic.AddInt32(&numTimesInitialized, 1)
		return data, nil
	})

	var wg sync.WaitGroup
	wg.Add(concurrency)

	var errors []error
	var mutex sync.Mutex

	for i := 0; i < concurrency; i++ {
		data := fmt.Sprintf("Iter_%d", i)
		go func() {
			defer wg.Done()
			value, err := ref.Get(data)
			if err != nil {
				panic(err.Error())
			}
			strVal := value.(string)
			if strVal[0:5] != "Iter_" {
				mutex.Lock()
				errors = append(errors, fmt.Errorf("expecting value to be %s but got %s", data, value))
				mutex.Unlock()
			}
		}()
	}

	wg.Wait()

	if len(errors) > 0 {
		t.Fatal(errors[0].Error())
	}
	if num := atomic.LoadInt32(&numTimesInitialized); num != int32(expectedTimesInitialized) {
		t.Fatalf("expecting initializer to be called %d time(s) but was called %d time(s)", expectedTimesInitialized, num)
	}
}

func TestMustGetPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expecting panic but got none")
		}
	}()

	ref := New(func() (interface{}, error) {
		return nil, fmt.Errorf("some error")
	})

	ref.MustGet()
	t.Fatal("Expecting panic but got none")
}

func TestGetWithFinalizer(t *testing.T) {
	var numTimesInitialized int32
	var numTimesFinalized int32
	expectedTimesInitialized := 1
	expectedTimesFinalized := 1
	concurrency := 100
	expectedValue := "Data1"

	ref := New(
		func() (interface{}, error) {
			t.Log("Initializing Reference...\n")
			atomic.AddInt32(&numTimesInitialized, 1)
			return expectedValue, nil
		},
		WithFinalizer(
			func(value interface{}) {
				t.Logf("Finalizer called - value: %s\n", value)
				atomic.AddInt32(&numTimesFinalized, 1)
				assert.Equal(t, expectedValue, value, "got different value than expected in finalizer")
			},
		),
	)

	var wg sync.WaitGroup
	wg.Add(concurrency)

	var errors []error
	var mutex sync.Mutex

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			value, err := ref.Get()
			if err != nil {
				panic(err.Error())
			}
			if value != expectedValue {
				mutex.Lock()
				errors = append(errors, fmt.Errorf("expecting value to be %s but got %s", expectedValue, value))
				mutex.Unlock()
			}
		}()
	}

	wg.Wait()
	ref.Close()

	if num := atomic.LoadInt32(&numTimesInitialized); num != int32(expectedTimesInitialized) {
		t.Fatalf("expecting initializer to be called %d time(s) but was called %d time(s)", expectedTimesInitialized, num)
	}
	if num := atomic.LoadInt32(&numTimesFinalized); num != int32(expectedTimesFinalized) {
		t.Fatalf("expecting finalizer to be called %d time(s) but was called %d time(s)", expectedTimesFinalized, num)
	}
	if len(errors) > 0 {
		t.Fatalf(errors[0].Error())
	}
}

func TestExpiring(t *testing.T) {
	var numTimesInitialized int32
	var numTimesFinalized int32
	concurrency := 20
	iterations := 100

	var seq int32
	ref := New(
		func() (interface{}, error) {
			atomic.AddInt32(&numTimesInitialized, 1)
			value := fmt.Sprintf("Data_%d", atomic.LoadInt32(&seq))
			t.Logf("Initializing to %s\n", value)
			return value, nil
		},
		WithFinalizer(
			func(interface{}) {
				atomic.AddInt32(&seq, 1)
				atomic.AddInt32(&numTimesFinalized, 1)
			},
		),
		WithAbsoluteExpiration(250*time.Millisecond),
	)

	var wg sync.WaitGroup
	wg.Add(concurrency)

	var errors []error
	var mutex sync.Mutex

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			previousValue := ""
			timesValueChanged := 0
			for j := 0; j < iterations; j++ {
				value, err := ref.Get()
				if err != nil {
					t.Logf("Got error: %s\n", err)
				}
				if previousValue != value {
					previousValue = value.(string)
					timesValueChanged++
				}
				time.Sleep(5 * time.Millisecond)
			}
			if timesValueChanged <= 1 {
				mutex.Lock()
				errors = append(errors, fmt.Errorf("expecting value to have changed multiple times but it changed %d time(s)", timesValueChanged))
				mutex.Unlock()
			}
		}()
	}

	wg.Wait()
	ref.Close()

	if len(errors) > 0 {
		t.Fatalf(errors[0].Error())
	}
	if num := atomic.LoadInt32(&numTimesInitialized); num <= 1 {
		t.Fatalf("expecting initializer to be called multiple times but was called %d time(s)", num)
	}
	if num := atomic.LoadInt32(&numTimesFinalized); num <= 1 {
		t.Fatalf("expecting finalizer to be called multiple times but was called %d time(s)", num)
	}
}

func TestExpiringWithErr(t *testing.T) {
	var numTimesInitialized int32
	var numTimesFinalized int32
	concurrency := 20
	iterations := 100

	seq := 0
	ref := New(
		func() (interface{}, error) {
			atomic.AddInt32(&numTimesInitialized, 1)
			if seq == 2 {
				seq++
				return nil, fmt.Errorf("returning error from initializer")
			}
			value := fmt.Sprintf("Data_%d", seq)
			t.Logf("Initializing to %s\n", value)
			return value, nil
		},
		WithFinalizer(
			func(interface{}) {
				atomic.AddInt32(&numTimesFinalized, 1)
				seq++
			},
		),
		WithExpirationProvider(
			NewGraduatingExpirationProvider(500*time.Millisecond, 1*time.Second, 5*time.Second),
			LastInitialized,
		),
	)

	var wg sync.WaitGroup
	wg.Add(concurrency)

	var errors []error
	var mutex sync.Mutex

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			previousValue := ""
			timesValueChanged := 0
			for j := 0; j < iterations; j++ {
				value, err := ref.Get()
				if err != nil {
					t.Logf("Got error: %s\n", err)
				} else if previousValue != value {
					previousValue = value.(string)
					timesValueChanged++
				}
				time.Sleep(50 * time.Millisecond)
			}
			if timesValueChanged <= 1 {
				mutex.Lock()
				errors = append(errors, fmt.Errorf("expecting value to have changed multiple times but it changed %d time(s)", timesValueChanged))
				mutex.Unlock()
			}
		}()
	}

	wg.Wait()
	ref.Close()

	if len(errors) > 0 {
		t.Fatalf(errors[0].Error())
	}
	if num := atomic.LoadInt32(&numTimesInitialized); num <= 1 {
		t.Fatalf("expecting initializer to be called multiple times but was called %d time(s)", num)
	}
	if num := atomic.LoadInt32(&numTimesFinalized); num <= 1 {
		t.Fatalf("expecting finalizer to be called multiple times but was called %d time(s)", num)
	}
}

func TestExpiringOnIdle(t *testing.T) {
	var numTimesInitialized int32
	var numTimesFinalized int32
	iterations := 20

	seq := 0
	ref := New(
		func() (interface{}, error) {
			atomic.AddInt32(&numTimesInitialized, 1)
			value := fmt.Sprintf("Data_%d", seq)
			t.Logf("Initializing to %s\n", value)
			return value, nil
		},
		WithFinalizer(
			func(interface{}) {
				seq++
				atomic.AddInt32(&numTimesFinalized, 1)
			},
		),
		WithIdleExpiration(100*time.Millisecond),
	)

	previousValue := ""
	timesValueChanged := 0
	for j := 0; j < iterations; j++ {
		value := ref.MustGet()
		if previousValue != value {
			previousValue = value.(string)
			timesValueChanged++
		}
		time.Sleep(time.Duration(20*j) * time.Millisecond)
	}

	// Wait for the ref to expire
	time.Sleep(time.Second)

	value := ref.MustGet()
	if previousValue != value {
		timesValueChanged++
	}

	if timesValueChanged <= 1 {
		t.Fatalf("expecting value to have changed multiple times but it changed %d time(s)", timesValueChanged)
	}

	ref.Close()

	if num := atomic.LoadInt32(&numTimesInitialized); num <= 1 {
		t.Fatalf("expecting initializer to be called multiple times but was called %d time(s)", num)
	}
	if num := atomic.LoadInt32(&numTimesFinalized); num <= 1 {
		t.Fatalf("expecting finalizer to be called multiple times but was called %d time(s)", num)
	}
}

func TestProactiveRefresh(t *testing.T) {
	var numTimesInitialized int32
	var numTimesFinalized int32
	expectedTimesFinalized := 1

	concurrency := 20
	iterations := 50

	seq := 0
	ref := New(
		func() (interface{}, error) {
			atomic.AddInt32(&numTimesInitialized, 1)
			seq++
			if seq == 3 {
				return nil, fmt.Errorf("returning error from initializer")
			}
			value := fmt.Sprintf("Data_%d", seq)
			t.Logf("Initializing to %s\n", value)
			return value, nil
		},
		WithFinalizer(
			func(interface{}) {
				atomic.AddInt32(&numTimesFinalized, 1)
				t.Log("Finalizer called")
			},
		),
		WithRefreshInterval(InitImmediately, 500*time.Millisecond),
	)

	var wg sync.WaitGroup
	wg.Add(concurrency)

	var errors []error
	var mutex sync.Mutex

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			previousValue := ""
			timesValueChanged := 0
			for j := 0; j < iterations; j++ {
				value, err := ref.Get()
				if err != nil {
					t.Logf("Got error: %s\n", err)
				} else if previousValue != value {
					previousValue = value.(string)
					timesValueChanged++
				}
				time.Sleep(50 * time.Millisecond)
			}
			if timesValueChanged < 2 {
				mutex.Lock()
				errors = append(errors, fmt.Errorf("expecting value to have changed multiple times but it changed %d time(s)", timesValueChanged))
				mutex.Unlock()
			}
		}()
	}

	wg.Wait()
	ref.Close()

	if len(errors) > 0 {
		t.Fatalf(errors[0].Error())
	}
	if num := atomic.LoadInt32(&numTimesInitialized); num < 2 {
		t.Fatalf("expecting initializer to be called multiple times but was called %d time(s)", num)
	}
	if num := atomic.LoadInt32(&numTimesFinalized); num != int32(expectedTimesFinalized) {
		t.Fatalf("expecting finalizer to be called %d time(s) but was called %d time(s)", expectedTimesFinalized, num)
	}
}
