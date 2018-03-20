/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package logging

import (
	"bytes"
	"sync"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/modlog"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/testdata"
	"github.com/stretchr/testify/assert"
)

var moduleName = "module-xyz"
var moduleName2 = "module-xyz-deftest"
var buf bytes.Buffer

func TestLoggingForCustomLogger(t *testing.T) {

	//Now add sample logger
	resetLoggerInstance()
	Initialize(testdata.GetSampleLoggingProvider(&buf))
	//Create new logger
	logger := NewLogger(moduleName)

	//Test logger.print outputs
	modlog.VerifyBasicLogging(t, api.INFO, logger.Print, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, api.INFO, logger.Println, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, api.INFO, nil, logger.Printf, &buf, true, moduleName)

	//Test logger.info outputs
	modlog.VerifyBasicLogging(t, api.INFO, logger.Info, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, api.INFO, logger.Infoln, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, api.INFO, nil, logger.Infof, &buf, true, moduleName)

	//Test logger.warn outputs
	modlog.VerifyBasicLogging(t, api.WARNING, logger.Warn, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, api.WARNING, logger.Warnln, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, api.WARNING, nil, logger.Warnf, &buf, true, moduleName)

	//In middle of test, get new logger, it should still stick to custom logger
	logger = NewLogger(moduleName)

	//Test logger.error outputs
	modlog.VerifyBasicLogging(t, api.ERROR, logger.Error, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, api.ERROR, logger.Errorln, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, api.ERROR, nil, logger.Errorf, &buf, true, moduleName)

	//Test logger.debug outputs
	modlog.VerifyBasicLogging(t, api.DEBUG, logger.Debug, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, api.DEBUG, logger.Debugln, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, api.DEBUG, nil, logger.Debugf, &buf, true, moduleName)

	////Test logger.fatal outputs - this custom logger doesn't cause os exit code 1
	modlog.VerifyBasicLogging(t, api.CRITICAL, logger.Fatal, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, api.CRITICAL, logger.Fatalln, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, api.CRITICAL, nil, logger.Fatalf, &buf, true, moduleName)

	//Test logger.panic outputs - this custom logger doesn't cause panic
	modlog.VerifyBasicLogging(t, api.CRITICAL, logger.Panic, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, api.CRITICAL, logger.Panicln, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, api.CRITICAL, nil, logger.Panicf, &buf, true, moduleName)

}

func TestDefaultModulledLoggingBehavior(t *testing.T) {

	//Init logger with default logger
	resetLoggerInstance()
	Initialize(modlog.LoggerProvider())
	//Get new logger
	dlogger := NewLogger(moduleName)
	// force initialization
	dlogger.logger()
	//Change output
	dlogger.instance.(*modlog.Log).ChangeOutput(&buf)

	//No level set for this module so log level should be info
	assert.True(t, api.INFO == modlog.GetLevel(moduleName), " default log level is INFO")

	//Test logger.print outputs
	modlog.VerifyBasicLogging(t, -1, dlogger.Print, nil, &buf, false, moduleName)
	modlog.VerifyBasicLogging(t, -1, dlogger.Println, nil, &buf, false, moduleName)
	modlog.VerifyBasicLogging(t, -1, nil, dlogger.Printf, &buf, false, moduleName)

	//Test logger.info outputs
	modlog.VerifyBasicLogging(t, api.INFO, dlogger.Info, nil, &buf, false, moduleName)
	modlog.VerifyBasicLogging(t, api.INFO, dlogger.Infoln, nil, &buf, false, moduleName)
	modlog.VerifyBasicLogging(t, api.INFO, nil, dlogger.Infof, &buf, false, moduleName)

	//Test logger.warn outputs
	modlog.VerifyBasicLogging(t, api.WARNING, dlogger.Warn, nil, &buf, false, moduleName)
	modlog.VerifyBasicLogging(t, api.WARNING, dlogger.Warnln, nil, &buf, false, moduleName)
	modlog.VerifyBasicLogging(t, api.WARNING, nil, dlogger.Warnf, &buf, false, moduleName)

	//Test logger.error outputs
	modlog.VerifyBasicLogging(t, api.ERROR, dlogger.Error, nil, &buf, false, moduleName)
	modlog.VerifyBasicLogging(t, api.ERROR, dlogger.Errorln, nil, &buf, false, moduleName)
	modlog.VerifyBasicLogging(t, api.ERROR, nil, dlogger.Errorf, &buf, false, moduleName)

	/*
		SINCE DEBUG LOG IS NOT YET ENABLED, LOG OUTPUT SHOULD BE EMPTY
	*/
	//Test logger.debug outputs when DEBUG level is not enabled
	dlogger.Debug("brown fox jumps over the lazy dog")
	dlogger.Debugln("brown fox jumps over the lazy dog")
	dlogger.Debugf("brown %s jumps over the lazy %s", "fox", "dog")

	assert.Empty(t, buf.String(), "debug log isn't supposed to show up for info level")

	//Should be false
	assert.False(t, modlog.IsEnabledFor(moduleName, api.DEBUG), "logging.IsEnabled for is not working as expected, expected false but got true")

	//Now change the log level to DEBUG
	modlog.SetLevel(moduleName, api.DEBUG)

	//Should be false
	assert.True(t, modlog.IsEnabledFor(moduleName, api.DEBUG), "logging.IsEnabled for is not working as expected, expected true but got false")

	//Test logger.debug outputs
	modlog.VerifyBasicLogging(t, api.DEBUG, dlogger.Debug, nil, &buf, false, moduleName)
	modlog.VerifyBasicLogging(t, api.DEBUG, dlogger.Debugln, nil, &buf, false, moduleName)
	modlog.VerifyBasicLogging(t, api.DEBUG, nil, dlogger.Debugf, &buf, false, moduleName)

}

func TestLoggerSetting(t *testing.T) {
	resetLoggerInstance()
	logger := NewLogger(moduleName)
	assert.True(t, loggerProviderInstance == nil, "Logger is not supposed to be initialized now")
	logger.Info("brown fox jumps over the lazy dog")
	assert.True(t, loggerProviderInstance != nil, "Logger is supposed to be initialized now")
	resetLoggerInstance()
	Initialize(modlog.LoggerProvider())
	assert.True(t, loggerProviderInstance != nil, "Logger is supposed to be initialized now")
}

func resetLoggerInstance() {
	loggerProviderInstance = nil
	loggerProviderOnce = sync.Once{}
}

func TestDefaultCustomModuledLoggingBehavior(t *testing.T) {

	//Init logger with default logger
	resetLoggerInstance()
	//Set custom logger in place of default logger
	modlog.InitLogger(testdata.GetSampleLoggingProvider(&buf))
	//Get new logger
	dlogger := NewLogger(moduleName2)
	// force initialization
	dlogger.logger()

	//No level set for this module so log level should be info
	assert.True(t, api.INFO == modlog.GetLevel(moduleName2), " default log level is INFO")

	//Test logger.print outputs
	modlog.VerifyBasicLogging(t, api.INFO, dlogger.Print, nil, &buf, true, moduleName2)
	modlog.VerifyBasicLogging(t, api.INFO, dlogger.Println, nil, &buf, true, moduleName2)
	modlog.VerifyBasicLogging(t, api.INFO, nil, dlogger.Printf, &buf, true, moduleName2)

	//Test logger.info outputs
	modlog.VerifyBasicLogging(t, api.INFO, dlogger.Info, nil, &buf, true, moduleName2)
	modlog.VerifyBasicLogging(t, api.INFO, dlogger.Infoln, nil, &buf, true, moduleName2)
	modlog.VerifyBasicLogging(t, api.INFO, nil, dlogger.Infof, &buf, true, moduleName2)

	//Test logger.warn outputs
	modlog.VerifyBasicLogging(t, api.WARNING, dlogger.Warn, nil, &buf, true, moduleName2)
	modlog.VerifyBasicLogging(t, api.WARNING, dlogger.Warnln, nil, &buf, true, moduleName2)
	modlog.VerifyBasicLogging(t, api.WARNING, nil, dlogger.Warnf, &buf, true, moduleName2)

	//Test logger.error outputs
	modlog.VerifyBasicLogging(t, api.ERROR, dlogger.Error, nil, &buf, true, moduleName2)
	modlog.VerifyBasicLogging(t, api.ERROR, dlogger.Errorln, nil, &buf, true, moduleName2)
	modlog.VerifyBasicLogging(t, api.ERROR, nil, dlogger.Errorf, &buf, true, moduleName2)

	//Should be false
	assert.False(t, modlog.IsEnabledFor(moduleName2, api.DEBUG), "logging.IsEnabled for is not working as expected, expected false but got true")

	//Now change the log level to DEBUG
	modlog.SetLevel(moduleName2, api.DEBUG)

	//Should be false
	assert.True(t, modlog.IsEnabledFor(moduleName2, api.DEBUG), "logging.IsEnabled for is not working as expected, expected true but got false")

	//Test logger.debug outputs
	modlog.VerifyBasicLogging(t, api.DEBUG, dlogger.Debug, nil, &buf, true, moduleName2)
	modlog.VerifyBasicLogging(t, api.DEBUG, dlogger.Debugln, nil, &buf, true, moduleName2)
	modlog.VerifyBasicLogging(t, api.DEBUG, nil, dlogger.Debugf, &buf, true, moduleName2)

}
