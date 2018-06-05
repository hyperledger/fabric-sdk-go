/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package futurevalue

import (
	"fmt"
	"sync"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/util/test"
)

func ExampleValue_Get() {
	fv := New(func() (interface{}, error) {
		return "Value1", nil
	})

	done := make(chan bool)
	go func() {
		value, err := fv.Get()
		if err != nil {
			fmt.Printf("Error returned from Get: %s\n", err)
		}
		fmt.Println(value)
		done <- true
	}()

	fv.Initialize()
	<-done
	// Output: Value1
}

func ExampleValue_MustGet() {
	fv := New(func() (interface{}, error) {
		return "Value1", nil
	})

	done := make(chan bool)
	go func() {
		fmt.Println(fv.MustGet())
		done <- true
	}()

	fv.Initialize()
	<-done
	// Output: Value1
}

func TestFutureValueGet(t *testing.T) {
	expectedValue := "Value1"

	fv := New(func() (interface{}, error) {
		return expectedValue, nil
	})

	concurrency := 100
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			value, err := fv.Get()
			if err != nil {
				test.Failf(t, "received error: %s", err)
			}
			if value != expectedValue {
				test.Failf(t, "expecting value [%s] but received [%s]", expectedValue, value)
			}
		}()
	}

	value, err := fv.Initialize()
	if err != nil {
		t.Fatalf("received error: %s", err)
	}

	wg.Wait()

	if value != expectedValue {
		t.Fatalf("expecting value [%s] but received [%s]", expectedValue, value)
	}
}

func TestFutureValueGetWithError(t *testing.T) {
	fv := New(func() (interface{}, error) {
		return nil, fmt.Errorf("some error")
	})

	concurrency := 100
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			if _, err := fv.Get(); err == nil {
				test.Failf(t, "expecting error but received none")
			}
		}()
	}

	if _, err := fv.Initialize(); err == nil {
		t.Fatal("expecting error but received none")
	}

	wg.Wait()
}

func TestMustGetPanic(t *testing.T) {
	done := make(chan bool)

	fv := New(func() (interface{}, error) {
		return nil, fmt.Errorf("some error")
	})

	go func() {
		defer func() {
			if r := recover(); r == nil {
				test.Failf(t, "Expecting panic but got none")
			}
			done <- true
		}()
		fv.MustGet()
		test.Failf(t, "Expecting panic but got none")
	}()

	if _, err := fv.Initialize(); err == nil {
		t.Fatal("expecting error but received none")
	}
	<-done
}
