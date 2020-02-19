/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package multi

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorString(t *testing.T) {
	testErr := fmt.Errorf("test")
	var errs Errors

	assert.Equal(t, "", errs.Error())

	errs = append(errs, testErr)
	assert.Equal(t, testErr.Error(), errs.Error())

	errs = append(errs, testErr)
	assert.Equal(t, "Multiple errors occurred: - test - test", errs.Error())
}

func TestAppend(t *testing.T) {
	testErr := fmt.Errorf("test")
	testErr2 := fmt.Errorf("test2")

	m := Append(nil, nil)
	assert.Nil(t, m)

	m = Append(nil, testErr)
	assert.Equal(t, testErr, m)

	m = Append(testErr, testErr2)
	m1, ok := m.(Errors)
	assert.True(t, ok)
	assert.Equal(t, testErr, m1[0])
	assert.Equal(t, testErr2, m1[1])

	m = Append(Errors{testErr}, nil)
	assert.Equal(t, Errors{testErr}, m)

	m = Append(Errors{testErr}, testErr2)
	m1, ok = m.(Errors)
	assert.True(t, ok)
	assert.Equal(t, testErr, m1[0])
	assert.Equal(t, testErr2, m1[1])
}

func TestToError(t *testing.T) {
	testErr := fmt.Errorf("test")
	var errs Errors

	assert.Equal(t, nil, errs.ToError())

	errs = append(errs, testErr)
	assert.Equal(t, testErr, errs.ToError())

	errs = append(errs, testErr)
	assert.Equal(t, errs, errs.ToError())
}
