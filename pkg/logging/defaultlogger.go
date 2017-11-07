/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
)

func getDefaultLogger(module string) apilogging.Logger {
	newLogger := log.New(os.Stdout, fmt.Sprintf(logPrefixFormatter, module), log.Ldate|log.Ltime|log.LUTC)
	return &DefaultLogger{defaultLogger: newLogger, module: module}
}

//DefaultLogger default underlying logger used by logging.Logger
type DefaultLogger struct {
	defaultLogger *log.Logger
	module        string
}

const (
	logLevelFormatter   = "UTC %s-> %s "
	logPrefixFormatter  = " [%s] "
	callerInfoFormatter = "- %s "
)

// Fatal is CRITICAL log followed by a call to os.Exit(1).
func (l *DefaultLogger) Fatal(args ...interface{}) {

	l.log(CRITICAL, args...)
	l.defaultLogger.Fatal(args...)
}

// Fatalf is CRITICAL log formatted followed by a call to os.Exit(1).
func (l *DefaultLogger) Fatalf(format string, args ...interface{}) {
	l.logf(CRITICAL, format, args...)
	l.defaultLogger.Fatalf(format, args...)
}

// Fatalln is CRITICAL log ln followed by a call to os.Exit(1).
func (l *DefaultLogger) Fatalln(args ...interface{}) {
	l.logln(CRITICAL, args...)
	l.defaultLogger.Fatalln(args...)
}

// Panic is CRITICAL log followed by a call to panic()
func (l *DefaultLogger) Panic(args ...interface{}) {
	l.log(CRITICAL, args...)
	l.defaultLogger.Panic(args...)
}

// Panicf is CRITICAL log formatted followed by a call to panic()
func (l *DefaultLogger) Panicf(format string, args ...interface{}) {
	l.logf(CRITICAL, format, args...)
	l.defaultLogger.Panicf(format, args...)
}

// Panicln is CRITICAL log ln followed by a call to panic()
func (l *DefaultLogger) Panicln(args ...interface{}) {
	l.logln(CRITICAL, args...)
	l.defaultLogger.Panicln(args...)
}

// Print calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *DefaultLogger) Print(args ...interface{}) {
	l.defaultLogger.Print(args...)
}

// Printf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *DefaultLogger) Printf(format string, args ...interface{}) {
	l.defaultLogger.Printf(format, args...)
}

// Println calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *DefaultLogger) Println(args ...interface{}) {
	l.defaultLogger.Println(args...)
}

// Debug calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *DefaultLogger) Debug(args ...interface{}) {
	l.log(DEBUG, args...)
}

// Debugf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *DefaultLogger) Debugf(format string, args ...interface{}) {
	l.logf(DEBUG, format, args...)
}

// Debugln calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *DefaultLogger) Debugln(args ...interface{}) {
	l.logln(DEBUG, args...)
}

// Info calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *DefaultLogger) Info(args ...interface{}) {
	l.log(INFO, args...)
}

// Infof calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *DefaultLogger) Infof(format string, args ...interface{}) {
	l.logf(INFO, format, args...)
}

// Infoln calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *DefaultLogger) Infoln(args ...interface{}) {
	l.logln(INFO, args...)
}

// Warn calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *DefaultLogger) Warn(args ...interface{}) {
	l.log(WARNING, args...)
}

// Warnf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *DefaultLogger) Warnf(format string, args ...interface{}) {
	l.logf(WARNING, format, args...)
}

// Warnln calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *DefaultLogger) Warnln(args ...interface{}) {
	l.logln(WARNING, args...)
}

// Error calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *DefaultLogger) Error(args ...interface{}) {
	l.log(ERROR, args...)
}

// Errorf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *DefaultLogger) Errorf(format string, args ...interface{}) {
	l.logf(ERROR, format, args...)
}

// Errorln calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *DefaultLogger) Errorln(args ...interface{}) {
	l.logln(ERROR, args...)
}

func (l *DefaultLogger) logf(level apilogging.Level, format string, args ...interface{}) {
	//Format prefix to show function name and log level and to indicate that timezone used is UTC
	customPrefix := fmt.Sprintf(logLevelFormatter, l.getCallerInfo(), levelNames[level])
	l.defaultLogger.Output(2, customPrefix+fmt.Sprintf(format, args...))
}

func (l *DefaultLogger) log(level apilogging.Level, args ...interface{}) {

	//Format prefix to show function name and log level and to indicate that timezone used is UTC
	customPrefix := fmt.Sprintf(logLevelFormatter, l.getCallerInfo(), levelNames[level])
	l.defaultLogger.Output(2, customPrefix+fmt.Sprint(args...))
}

func (l *DefaultLogger) logln(level apilogging.Level, args ...interface{}) {
	//Format prefix to show function name and log level and to indicate that timezone used is UTC
	customPrefix := fmt.Sprintf(logLevelFormatter, l.getCallerInfo(), levelNames[level])
	l.defaultLogger.Output(2, customPrefix+fmt.Sprintln(args...))
}

func (l *DefaultLogger) getCallerInfo() string {

	if !IsCallerInfoEnabled(l.module) {
		return ""
	}

	const MAXCALLERS = 5                  // search MAXCALLERS frames for the real caller
	const SKIPCALLERS = 4                 // skip SKIPCALLERS frames when determining the real caller
	const LOGPREFIX = "logging.(*Logger)" // LOGPREFIX indicates the upcoming frame contains the real caller and skip the frame
	const LOGBRIDGEPREFIX = "logbridge."  // LOGBRIDGEPREFIX indicates to skip the frame due to being a logbridge
	const NOTFOUND = "n/a"

	fpcs := make([]uintptr, MAXCALLERS)

	n := runtime.Callers(SKIPCALLERS, fpcs)
	if n == 0 {
		return fmt.Sprintf(callerInfoFormatter, NOTFOUND)
	}

	frames := runtime.CallersFrames(fpcs[:n])
	funcIsNext := false
	for f, more := frames.Next(); more; f, more = frames.Next() {
		_, funName := filepath.Split(f.Function)
		if f.Func == nil || f.Function == "" {
			funName = NOTFOUND // not a function or unknown
		}

		if strings.HasPrefix(funName, LOGPREFIX) || strings.HasPrefix(funName, LOGBRIDGEPREFIX) {
			funcIsNext = true
		} else if funcIsNext {
			return fmt.Sprintf(callerInfoFormatter, funName)
		}
	}

	return fmt.Sprintf(callerInfoFormatter, NOTFOUND)
}
