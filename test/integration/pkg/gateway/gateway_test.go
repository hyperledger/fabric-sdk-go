/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"testing"
)

func TestGatewayFromConfig(t *testing.T) {
	t.Run("Base", func(t *testing.T) {
		RunWithConfig(t)
	})
}

func TestGatewayFromSDK(t *testing.T) {
	t.Run("Base", func(t *testing.T) {
		RunWithSDK(t)
	})
}
