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

func TestGatewayWithSubmit(t *testing.T) {
	t.Run("Base", func(t *testing.T) {
		RunWithSubmit(t)
	})
}

func TestGatewayWithWallet(t *testing.T) {
	t.Run("Base", func(t *testing.T) {
		RunWithWallet(t)
	})
}

func TestTransientData(t *testing.T) {
	t.Run("Base", func(t *testing.T) {
		RunWithTransient(t)
	})
}

func TestContractEvent(t *testing.T) {
	t.Run("Base", func(t *testing.T) {
		RunWithContractEvent(t)
	})
}

func TestBlockEvent(t *testing.T) {
	t.Run("Base", func(t *testing.T) {
		RunWithBlockEvent(t)
	})
}
