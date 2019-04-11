// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"path/filepath"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/modlog"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

func TestDefLoggerFactory(t *testing.T) {
	// Cleanup logging singleton
	logging.UnsafeReset()

	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, sdkConfigFile)
	_, err := New(configImpl.FromFile(configPath))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	const moduleName = "mymodule"
	l1 := logging.NewLogger(moduleName)

	// output a log message to force initializatin
	l1.Info("message")

	// ensure that the logger cannot be overridden
	// (initializing a new logger should have no effect)
	lf := NewMockLoggerFactory()
	logging.Initialize(lf)

	l2 := logging.NewLogger(moduleName)

	// output a log message to force initializatin
	l2.Info("message")

	if lf.ActiveModules[moduleName] {
		t.Fatal("Unexpected logger factory is set")
	}
}

func TestOptLoggerFactory(t *testing.T) {
	// Cleanup logging singleton
	logging.UnsafeReset()

	lf := NewMockLoggerFactory()

	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, sdkConfigFile)
	_, err := New(configImpl.FromFile(configPath),
		WithLoggerPkg(lf))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	const moduleName = "mymodule"
	l := logging.NewLogger(moduleName)

	// output a log message to force initializatin
	l.Info("message")

	if !lf.ActiveModules[moduleName] {
		t.Fatal("Unexpected logger factory is set")
	}
}

// MockLoggerFactory records the modules that have loggers
type MockLoggerFactory struct {
	ActiveModules map[string]bool
	logger        api.LoggerProvider
}

func NewMockLoggerFactory() *MockLoggerFactory {
	lf := MockLoggerFactory{}
	lf.ActiveModules = make(map[string]bool)
	lf.logger = modlog.LoggerProvider()

	return &lf
}

func (lf *MockLoggerFactory) GetLogger(module string) api.Logger {
	lf.ActiveModules[module] = true
	return lf.logger.GetLogger(module)
}
