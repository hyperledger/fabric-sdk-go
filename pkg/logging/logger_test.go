/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package logging

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"testing"
)

const (
	basicLevelOutputExpectedRegex  = "\\[%s\\] .* UTC - logging.* -> %s brown fox jumps over the lazy dog"
	printLevelOutputExpectedRegex  = "\\[%s\\] .* brown fox jumps over the lazy dog"
	customLevelOutputExpectedRegex = "\\[%s\\] .* CUSTOM LOG OUTPUT"
	moduleName                     = "module-xyz"
)

type fn func(...interface{})
type fnf func(string, ...interface{})

func TestDefaultLogging(t *testing.T) {

	SetCustomLogger(nil)
	logger := NewLogger(moduleName)

	//change the output to buffer
	var buf bytes.Buffer
	logger.logger.(*DefaultLogger).defaultLogger.SetOutput(&buf)

	//No level set for this module so log level should be info
	verifyTrue(t, INFO == GetLevel(moduleName), " default log level is INFO")

	//Test logger.print outputs
	verifyBasicLogging(t, -1, logger.Print, nil, &buf, false)
	verifyBasicLogging(t, -1, logger.Println, nil, &buf, false)
	verifyBasicLogging(t, -1, nil, logger.Printf, &buf, false)

	//Test logger.info outputs
	verifyBasicLogging(t, INFO, logger.Info, nil, &buf, false)
	verifyBasicLogging(t, INFO, logger.Infoln, nil, &buf, false)
	verifyBasicLogging(t, INFO, nil, logger.Infof, &buf, false)

	//Test logger.warn outputs
	verifyBasicLogging(t, WARNING, logger.Warn, nil, &buf, false)
	verifyBasicLogging(t, WARNING, logger.Warnln, nil, &buf, false)
	verifyBasicLogging(t, WARNING, nil, logger.Warnf, &buf, false)

	//Test logger.error outputs
	verifyBasicLogging(t, ERROR, logger.Error, nil, &buf, false)
	verifyBasicLogging(t, ERROR, logger.Errorln, nil, &buf, false)
	verifyBasicLogging(t, ERROR, nil, logger.Errorf, &buf, false)

	/*
		SINCE DEBUG LOG IS NOT YET ENABLED, LOG OUTPUT SHOULD BE EMPTY
	*/
	//Test logger.debug outputs when DEBUG level is not enabled
	logger.Debug("brown fox jumps over the lazy dog")
	logger.Debugln("brown fox jumps over the lazy dog")
	logger.Debugf("brown %s jumps over the lazy %s", "fox", "dog")

	verifyEmpty(t, buf.String(), "debug log isn't supposed to show up for info level")

	//Now change the log level to DEBUG
	SetLevel(DEBUG, moduleName)

	//Test logger.debug outputs
	verifyBasicLogging(t, DEBUG, logger.Debug, nil, &buf, false)
	verifyBasicLogging(t, DEBUG, logger.Debugln, nil, &buf, false)
	verifyBasicLogging(t, DEBUG, nil, logger.Debugf, &buf, false)

}

func TestLevelledLoggingForCustomLogger(t *testing.T) {

	//prepare custom logger for which output is bytes buffer
	var buf bytes.Buffer
	customLogger := log.New(&buf, fmt.Sprintf(logPrefixFormatter, moduleName), log.Ldate|log.Ltime|log.LUTC)

	//Create new logger
	logger := NewLogger(moduleName)
	//Now add custom logger
	SetCustomLogger(&SampleCustomLogger{customLogger: customLogger})

	//Test logger.print outputs
	verifyBasicLogging(t, INFO, logger.Print, nil, &buf, true)
	verifyBasicLogging(t, INFO, logger.Println, nil, &buf, true)
	verifyBasicLogging(t, INFO, nil, logger.Printf, &buf, true)

	//Test logger.info outputs
	verifyBasicLogging(t, INFO, logger.Info, nil, &buf, true)
	verifyBasicLogging(t, INFO, logger.Infoln, nil, &buf, true)
	verifyBasicLogging(t, INFO, nil, logger.Infof, &buf, true)

	//Test logger.warn outputs
	verifyBasicLogging(t, WARNING, logger.Warn, nil, &buf, true)
	verifyBasicLogging(t, WARNING, logger.Warnln, nil, &buf, true)
	verifyBasicLogging(t, WARNING, nil, logger.Warnf, &buf, true)

	//In middle of test, get new logger, it should still stick to custom logger
	logger = NewLogger(moduleName)

	//Test logger.error outputs
	verifyBasicLogging(t, ERROR, logger.Error, nil, &buf, true)
	verifyBasicLogging(t, ERROR, logger.Errorln, nil, &buf, true)
	verifyBasicLogging(t, ERROR, nil, logger.Errorf, &buf, true)

	//Test logger.debug outputs
	verifyBasicLogging(t, DEBUG, logger.Debug, nil, &buf, true)
	verifyBasicLogging(t, DEBUG, logger.Debugln, nil, &buf, true)
	verifyBasicLogging(t, DEBUG, nil, logger.Debugf, &buf, true)

	////Test logger.fatal outputs - this custom logger doesn't cause os exit code 1
	verifyBasicLogging(t, CRITICAL, logger.Fatal, nil, &buf, true)
	verifyBasicLogging(t, CRITICAL, logger.Fatalln, nil, &buf, true)
	verifyBasicLogging(t, CRITICAL, nil, logger.Fatalf, &buf, true)

	//Test logger.panic outputs - this custom logger doesn't cause panic
	verifyBasicLogging(t, CRITICAL, logger.Panic, nil, &buf, true)
	verifyBasicLogging(t, CRITICAL, logger.Panicln, nil, &buf, true)
	verifyBasicLogging(t, CRITICAL, nil, logger.Panicf, &buf, true)
}

func TestDefaultLoggingPanic(t *testing.T) {

	//Reset custom logger, need default one
	SetCustomLogger(nil)
	logger := NewLogger(moduleName)

	verifyCriticalLoggings(t, CRITICAL, logger.Panic, nil, logger)
	verifyCriticalLoggings(t, CRITICAL, logger.Panicln, nil, logger)
	verifyCriticalLoggings(t, CRITICAL, nil, logger.Panicf, logger)

}

//verifyCriticalLoggings utility func which does job calling and verifying CRITICAL log level functions - PANIC
func verifyCriticalLoggings(t *testing.T, level Level, loggerFunc fn, loggerFuncf fnf, logger *Logger) {

	//change the output to buffer
	var buf bytes.Buffer
	logger.logger.(*DefaultLogger).defaultLogger.SetOutput(&buf)

	//Handling panic as well as checking log output
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%v was supposed to panic", loggerFunc)
		}
		regex := fmt.Sprintf(basicLevelOutputExpectedRegex, moduleName, levelNames[level])
		match, err := regexp.MatchString(regex, buf.String())
		verifyEmpty(t, err, "error while matching regex with logoutput wasnt expected")
		verifyTrue(t, match, "CRITICAL logger isn't producing output as expected, \n logoutput:%s\n regex: %s", buf.String(), regex)

	}()

	//Call logger func
	if loggerFunc != nil {
		loggerFunc("brown fox jumps over the lazy dog")
	} else if loggerFuncf != nil {
		loggerFuncf("brown %s jumps over the lazy %s", "fox", "dog")
	}
}

//verifyBasicLogging utility func which does job calling and verifying basic log level functions - DEBUG, INFO, ERROR, WARNING
func verifyBasicLogging(t *testing.T, level Level, loggerFunc fn, loggerFuncf fnf, buf *bytes.Buffer, verifyCustom bool) {

	//Call logger func
	if loggerFunc != nil {
		loggerFunc("brown fox jumps over the lazy dog")
	} else if loggerFuncf != nil {
		loggerFuncf("brown %s jumps over the lazy %s", "fox", "dog")
	}

	//check output
	regex := ""
	levelName := "print"

	if verifyCustom {
		levelName = levelNames[level]
		regex = fmt.Sprintf(customLevelOutputExpectedRegex, moduleName)
	} else if level > 0 && !verifyCustom {
		levelName = levelNames[level]
		regex = fmt.Sprintf(basicLevelOutputExpectedRegex, moduleName, levelName)
	} else {
		regex = fmt.Sprintf(printLevelOutputExpectedRegex, moduleName)
	}

	match, err := regexp.MatchString(regex, buf.String())

	verifyEmpty(t, err, "error while matching regex with logoutput wasnt expected")
	verifyTrue(t, match, "%s logger isn't producing output as expected, \n logoutput:%s\n regex: %s", levelName, buf.String(), regex)

	//Reset output buffer, for next use
	buf.Reset()
}

func verifyTrue(t *testing.T, input bool, msgAndArgs ...interface{}) {
	if !input {
		failTest(t, msgAndArgs)
	}
}

func verifyFalse(t *testing.T, input bool, msgAndArgs ...interface{}) {
	if input {
		failTest(t, msgAndArgs)
	}
}

func verifyEmpty(t *testing.T, in interface{}, msgAndArgs ...interface{}) {
	if in == nil {
		return
	} else if in == "" {
		return
	}
	failTest(t, msgAndArgs...)
}

func verifyNotEmpty(t *testing.T, in interface{}, msgAndArgs ...interface{}) {
	if in != nil {
		return
	} else if in != "" {
		return
	}
	failTest(t, msgAndArgs...)
}

func failTest(t *testing.T, msgAndArgs ...interface{}) {
	if len(msgAndArgs) == 1 {
		t.Fatal(msgAndArgs[0])
	}
	if len(msgAndArgs) > 1 {
		t.Fatalf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
}

/*
	Test custom logger
*/

type SampleCustomLogger struct {
	customLogger *log.Logger
}

func (l *SampleCustomLogger) Fatal(v ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleCustomLogger) Fatalf(format string, v ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleCustomLogger) Fatalln(v ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleCustomLogger) Panic(v ...interface{})   { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleCustomLogger) Panicf(format string, v ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleCustomLogger) Panicln(v ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleCustomLogger) Print(v ...interface{})   { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleCustomLogger) Printf(format string, v ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleCustomLogger) Println(v ...interface{})  { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleCustomLogger) Debug(args ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleCustomLogger) Debugf(format string, args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleCustomLogger) Debugln(args ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleCustomLogger) Info(args ...interface{})    { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleCustomLogger) Infof(format string, args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleCustomLogger) Infoln(args ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleCustomLogger) Warn(args ...interface{})   { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleCustomLogger) Warnf(format string, args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleCustomLogger) Warnln(args ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleCustomLogger) Error(args ...interface{})  { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleCustomLogger) Errorf(format string, args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleCustomLogger) Errorln(args ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
