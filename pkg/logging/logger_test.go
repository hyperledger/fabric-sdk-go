/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package logging

import (
	"bytes"
	"fmt"
	"log"
	"sync"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/deflogger"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/utils"
)

var moduleName = "module-xyz"
var logPrefixFormatter = " [%s] "
var buf bytes.Buffer

func TestLevelledLoggingForCustomLogger(t *testing.T) {

	//Now add sample logger
	resetLoggerInstance()
	InitLogger(&sampleLoggingProvider{})
	//Create new logger
	logger := NewLogger(moduleName)

	//Test logger.print outputs
	deflogger.VerifyBasicLogging(t, apilogging.INFO, logger.Print, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.INFO, logger.Println, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.INFO, nil, logger.Printf, &buf, true)

	//Test logger.info outputs
	deflogger.VerifyBasicLogging(t, apilogging.INFO, logger.Info, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.INFO, logger.Infoln, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.INFO, nil, logger.Infof, &buf, true)

	//Test logger.warn outputs
	deflogger.VerifyBasicLogging(t, apilogging.WARNING, logger.Warn, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.WARNING, logger.Warnln, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.WARNING, nil, logger.Warnf, &buf, true)

	//In middle of test, get new logger, it should still stick to custom logger
	logger = NewLogger(moduleName)

	//Test logger.error outputs
	deflogger.VerifyBasicLogging(t, apilogging.ERROR, logger.Error, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.ERROR, logger.Errorln, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.ERROR, nil, logger.Errorf, &buf, true)

	//Test logger.debug outputs
	deflogger.VerifyBasicLogging(t, apilogging.DEBUG, logger.Debug, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.DEBUG, logger.Debugln, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.DEBUG, nil, logger.Debugf, &buf, true)

	////Test logger.fatal outputs - this custom logger doesn't cause os exit code 1
	deflogger.VerifyBasicLogging(t, apilogging.CRITICAL, logger.Fatal, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.CRITICAL, logger.Fatalln, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.CRITICAL, nil, logger.Fatalf, &buf, true)

	//Test logger.panic outputs - this custom logger doesn't cause panic
	deflogger.VerifyBasicLogging(t, apilogging.CRITICAL, logger.Panic, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.CRITICAL, logger.Panicln, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.CRITICAL, nil, logger.Panicf, &buf, true)

}

/*
	Sample logging provider
*/
type sampleLoggingProvider struct {
}

//GetLogger returns default logger implementation
func (p *sampleLoggingProvider) GetLogger(module string) apilogging.Logger {
	sampleLogger := log.New(&buf, fmt.Sprintf(logPrefixFormatter, moduleName), log.Ldate|log.Ltime|log.LUTC)
	return &SampleLogger{customLogger: sampleLogger, module: moduleName}
}

/*
	Sample logger
*/

type SampleLogger struct {
	customLogger *log.Logger
	module       string
}

func (l *SampleLogger) Fatal(v ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleLogger) Fatalf(format string, v ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Fatalln(v ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleLogger) Panic(v ...interface{})   { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleLogger) Panicf(format string, v ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Panicln(v ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleLogger) Print(v ...interface{})   { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleLogger) Printf(format string, v ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Println(v ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleLogger) Debug(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Debugf(format string, args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Debugln(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Info(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Infof(format string, args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Infoln(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Warn(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Warnf(format string, args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Warnln(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Error(args ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleLogger) Errorf(format string, args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Errorln(args ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }

func TestDefaultBehavior(t *testing.T) {

	//Init logger with default logger
	resetLoggerInstance()
	InitLogger(deflogger.LoggerProvider())
	//Get new logger
	dlogger := NewLogger(moduleName)
	// force initialization
	dlogger.logger()
	//Change output
	dlogger.instance.(*deflogger.Logger).ChangeOutput(&buf)

	//No level set for this module so log level should be info
	utils.VerifyTrue(t, apilogging.INFO == deflogger.GetLevel(moduleName), " default log level is INFO")

	//Test logger.print outputs
	deflogger.VerifyBasicLogging(t, -1, dlogger.Print, nil, &buf, false)
	deflogger.VerifyBasicLogging(t, -1, dlogger.Println, nil, &buf, false)
	deflogger.VerifyBasicLogging(t, -1, nil, dlogger.Printf, &buf, false)

	//Test logger.info outputs
	deflogger.VerifyBasicLogging(t, apilogging.INFO, dlogger.Info, nil, &buf, false)
	deflogger.VerifyBasicLogging(t, apilogging.INFO, dlogger.Infoln, nil, &buf, false)
	deflogger.VerifyBasicLogging(t, apilogging.INFO, nil, dlogger.Infof, &buf, false)

	//Test logger.warn outputs
	deflogger.VerifyBasicLogging(t, apilogging.WARNING, dlogger.Warn, nil, &buf, false)
	deflogger.VerifyBasicLogging(t, apilogging.WARNING, dlogger.Warnln, nil, &buf, false)
	deflogger.VerifyBasicLogging(t, apilogging.WARNING, nil, dlogger.Warnf, &buf, false)

	//Test logger.error outputs
	deflogger.VerifyBasicLogging(t, apilogging.ERROR, dlogger.Error, nil, &buf, false)
	deflogger.VerifyBasicLogging(t, apilogging.ERROR, dlogger.Errorln, nil, &buf, false)
	deflogger.VerifyBasicLogging(t, apilogging.ERROR, nil, dlogger.Errorf, &buf, false)

	/*
		SINCE DEBUG LOG IS NOT YET ENABLED, LOG OUTPUT SHOULD BE EMPTY
	*/
	//Test logger.debug outputs when DEBUG level is not enabled
	dlogger.Debug("brown fox jumps over the lazy dog")
	dlogger.Debugln("brown fox jumps over the lazy dog")
	dlogger.Debugf("brown %s jumps over the lazy %s", "fox", "dog")

	utils.VerifyEmpty(t, buf.String(), "debug log isn't supposed to show up for info level")

	//Should be false
	utils.VerifyFalse(t, deflogger.IsEnabledFor(moduleName, apilogging.DEBUG), "logging.IsEnabled for is not working as expected, expected false but got true")

	//Now change the log level to DEBUG
	deflogger.SetLevel(moduleName, apilogging.DEBUG)

	//Should be false
	utils.VerifyTrue(t, deflogger.IsEnabledFor(moduleName, apilogging.DEBUG), "logging.IsEnabled for is not working as expected, expected true but got false")

	//Test logger.debug outputs
	deflogger.VerifyBasicLogging(t, apilogging.DEBUG, dlogger.Debug, nil, &buf, false)
	deflogger.VerifyBasicLogging(t, apilogging.DEBUG, dlogger.Debugln, nil, &buf, false)
	deflogger.VerifyBasicLogging(t, apilogging.DEBUG, nil, dlogger.Debugf, &buf, false)

}

func TestLoggerSetting(t *testing.T) {
	resetLoggerInstance()
	logger := NewLogger(moduleName)
	utils.VerifyTrue(t, loggerProviderInstance == nil, "Logger is not supposed to be initialized now")
	logger.Info("brown fox jumps over the lazy dog")
	utils.VerifyTrue(t, loggerProviderInstance != nil, "Logger is supposed to be initialized now")
	resetLoggerInstance()
	InitLogger(deflogger.LoggerProvider())
	utils.VerifyTrue(t, loggerProviderInstance != nil, "Logger is supposed to be initialized now")
}

func resetLoggerInstance() {
	loggerProviderInstance = nil
	loggerProviderOnce = sync.Once{}
}
