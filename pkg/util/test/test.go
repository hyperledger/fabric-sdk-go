/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package test

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

// Logf writes to stdout and flushes. Applicable for when t.Logf can't be used.
func Logf(template string, args ...interface{}) {
	f := bufio.NewWriter(os.Stdout)

	_, err := f.WriteString(fmt.Sprintf(template, args...))
	if err != nil {
		panic(fmt.Sprintf("writing to output failed: %s", err))
	}

	err = f.WriteByte('\n')
	if err != nil {
		panic(fmt.Sprintf("writing to output failed: %s", err))
	}

	err = f.Flush()
	if err != nil {
		panic(fmt.Sprintf("writing to output failed: %s", err))
	}
}

// Failf - as t.Fatalf is not goroutine safe, this function behaves like t.Fatalf.
func Failf(t *testing.T, template string, args ...interface{}) {
	f := bufio.NewWriter(os.Stdout)

	_, err := f.WriteString(fmt.Sprintf(template, args...))
	if err != nil {
		panic(fmt.Sprintf("writing to output failed: %s", err))
	}

	err = f.WriteByte('\n')
	if err != nil {
		panic(fmt.Sprintf("writing to output failed: %s", err))
	}

	err = f.Flush()
	if err != nil {
		panic(fmt.Sprintf("writing to output failed: %s", err))
	}

	t.Fail()
}
