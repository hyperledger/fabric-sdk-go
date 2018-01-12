// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defclient"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defsvc"
	apisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/deflogger"
)

func dePkgSuiteWithLogger(logger apilogging.LoggerProvider) SDKOption {
	pkgSuite := apisdk.PkgSuite{
		Core:    defcore.NewProviderFactory(),
		Service: defsvc.NewProviderFactory(),
		Context: defclient.NewOrgClientFactory(),
		Session: defclient.NewSessionClientFactory(),
		Logger:  logger,
	}
	return PkgSuiteAsOpt(pkgSuite)
}
func TestDefLoggerFactory(t *testing.T) {
	// Cleanup logging singleton
	logging.UnsafeReset()

	_, err := New(ConfigFile("../../test/fixtures/config/config_test.yaml"), defPkgSuite())
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	const moduleName = "mymodule"
	l1, err := logging.GetLogger(moduleName)
	if err != nil {
		t.Fatal("Unexpected error getting logger")
	}

	// output a log message to force initializatin
	l1.Info("message")

	// ensure that the logger cannot be overridden
	// (initializing a new logger should have no effect)
	lf := NewMockLoggerFactory()
	logging.InitLogger(lf)

	l2, err := logging.GetLogger(moduleName)
	if err != nil {
		t.Fatal("Unexpected error getting logger")
	}

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

	_, err := New(ConfigFile("../../test/fixtures/config/config_test.yaml"), dePkgSuiteWithLogger(lf))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	const moduleName = "mymodule"
	l, err := logging.GetLogger(moduleName)
	if err != nil {
		t.Fatal("Unexpected error getting logger")
	}

	// output a log message to force initializatin
	l.Info("message")

	if !lf.ActiveModules[moduleName] {
		t.Fatal("Unexpected logger factory is set")
	}
}

// MockLoggerFactory records the modules that have loggers
type MockLoggerFactory struct {
	ActiveModules map[string]bool
	logger        apilogging.LoggerProvider
}

func NewMockLoggerFactory() *MockLoggerFactory {
	lf := MockLoggerFactory{}
	lf.ActiveModules = make(map[string]bool)
	lf.logger = deflogger.LoggerProvider()

	return &lf
}

func (lf *MockLoggerFactory) GetLogger(module string) apilogging.Logger {
	lf.ActiveModules[module] = true
	return lf.logger.GetLogger(module)
}
