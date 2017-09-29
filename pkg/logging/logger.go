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

	"sync"

	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
)

var mutex = &sync.Mutex{}

//Logger basic implementation of api.Logger interface
type Logger struct {
	logger apilogging.Logger
	module string
}

var moduleLevels = moduleLeveled{}

var customLogger apilogging.Logger

const (
	logLevelFormatter  = "UTC - %s -> %s "
	logPrefixFormatter = " [%s] "
)

// GetLogger creates and returns a Logger object based on the module name.
func GetLogger(module string) (*Logger, error) {
	return &Logger{logger: getDefaultLogger(module), module: module}, nil
}

// NewLogger is like GetLogger but panics if the logger can't be created.
func NewLogger(module string) *Logger {
	logger, err := GetLogger(module)
	if err != nil {
		panic("logger: " + module + ": " + err.Error())
	}
	return logger
}

//SetCustomLogger sets new custom logger which takes over logging operations already created and
//new logger which are going to be created. Care should be taken while using this method.
//It is recommended to add Custom loggers before making any loggings.
func SetCustomLogger(newCustomLogger apilogging.Logger) {
	mutex.Lock()
	customLogger = newCustomLogger
	mutex.Unlock()
}

//SetLevel - setting log level for given module
func SetLevel(level Level, module string) {
	moduleLevels.SetLevel(level, module)
}

//GetLevel - getting log level for given module
func GetLevel(module string) Level {
	return moduleLevels.GetLevel(module)
}

//IsEnabledFor - Check if given log level is enabled for given module
func IsEnabledFor(level Level, module string) bool {
	return moduleLevels.IsEnabledFor(level, module)
}

//Fatal calls Fatal function of underlying logger
func (l *Logger) Fatal(args ...interface{}) {
	l.getCurrentLogger().Fatal(args...)
}

//Fatalf calls Fatalf function of underlying logger
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.getCurrentLogger().Fatalf(format, args...)
}

//Fatalln calls Fatalln function of underlying logger
func (l *Logger) Fatalln(args ...interface{}) {
	l.getCurrentLogger().Fatalln(args...)
}

//Panic calls Panic function of underlying logger
func (l *Logger) Panic(args ...interface{}) {
	l.getCurrentLogger().Panic(args...)
}

//Panicf calls Panicf function of underlying logger
func (l *Logger) Panicf(format string, args ...interface{}) {
	l.getCurrentLogger().Panicf(format, args...)
}

//Panicln calls Panicln function of underlying logger
func (l *Logger) Panicln(args ...interface{}) {
	l.getCurrentLogger().Panicln(args...)
}

//Print calls Print function of underlying logger
func (l *Logger) Print(args ...interface{}) {
	l.getCurrentLogger().Print(args...)
}

//Printf calls Printf function of underlying logger
func (l *Logger) Printf(format string, args ...interface{}) {
	l.getCurrentLogger().Printf(format, args...)
}

//Println calls Println function of underlying logger
func (l *Logger) Println(args ...interface{}) {
	l.getCurrentLogger().Println(args...)
}

//Debug calls Debug function of underlying logger
func (l *Logger) Debug(args ...interface{}) {
	if IsEnabledFor(DEBUG, l.module) {
		l.getCurrentLogger().Debug(args...)
	}
}

//Debugf calls Debugf function of underlying logger
func (l *Logger) Debugf(format string, args ...interface{}) {
	if IsEnabledFor(DEBUG, l.module) {
		l.getCurrentLogger().Debugf(format, args...)
	}
}

//Debugln calls Debugln function of underlying logger
func (l *Logger) Debugln(args ...interface{}) {
	if IsEnabledFor(DEBUG, l.module) {
		l.getCurrentLogger().Debugln(args...)
	}
}

//Info calls Info function of underlying logger
func (l *Logger) Info(args ...interface{}) {
	if IsEnabledFor(INFO, l.module) {
		l.getCurrentLogger().Info(args...)
	}
}

//Infof calls Infof function of underlying logger
func (l *Logger) Infof(format string, args ...interface{}) {
	if IsEnabledFor(INFO, l.module) {
		l.getCurrentLogger().Infof(format, args...)
	}
}

//Infoln calls Infoln function of underlying logger
func (l *Logger) Infoln(args ...interface{}) {
	if IsEnabledFor(INFO, l.module) {
		l.getCurrentLogger().Infoln(args...)
	}
}

//Warn calls Warn function of underlying logger
func (l *Logger) Warn(args ...interface{}) {
	if IsEnabledFor(WARNING, l.module) {
		l.getCurrentLogger().Warn(args...)
	}
}

//Warnf calls Warnf function of underlying logger
func (l *Logger) Warnf(format string, args ...interface{}) {
	if IsEnabledFor(WARNING, l.module) {
		l.getCurrentLogger().Warnf(format, args...)
	}
}

//Warnln calls Warnln function of underlying logger
func (l *Logger) Warnln(args ...interface{}) {
	if IsEnabledFor(WARNING, l.module) {
		l.getCurrentLogger().Warnln(args...)
	}
}

//Error calls Error function of underlying logger
func (l *Logger) Error(args ...interface{}) {
	if IsEnabledFor(ERROR, l.module) {
		l.getCurrentLogger().Error(args...)
	}
}

//Errorf calls Errorf function of underlying logger
func (l *Logger) Errorf(format string, args ...interface{}) {
	if IsEnabledFor(ERROR, l.module) {
		l.getCurrentLogger().Errorf(format, args...)
	}
}

//Errorln calls Errorln function of underlying logger
func (l *Logger) Errorln(args ...interface{}) {
	if IsEnabledFor(ERROR, l.module) {
		l.getCurrentLogger().Errorln(args...)
	}
}

//getCurrentLogger - returns customlogger is set, or default logger
func (l *Logger) getCurrentLogger() apilogging.Logger {
	if customLogger != nil {
		return customLogger
	}
	return l.logger
}

/*
	Default logger Implementation
*/

func getDefaultLogger(module string) apilogging.Logger {
	newLogger := log.New(os.Stdout, fmt.Sprintf(logPrefixFormatter, module), log.Ldate|log.Ltime|log.LUTC)
	return &DefaultLogger{defaultLogger: newLogger}
}

//DefaultLogger default underlying logger used by logging.Logger
type DefaultLogger struct {
	defaultLogger *log.Logger
}

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

func (l *DefaultLogger) logf(level Level, format string, args ...interface{}) {
	//Format prefix to show function name and log level and to indicate that timezone used is UTC
	customPrefix := fmt.Sprintf(logLevelFormatter, l.getCaller(), level)
	l.defaultLogger.Output(2, customPrefix+fmt.Sprintf(format, args...))
}

func (l *DefaultLogger) log(level Level, args ...interface{}) {

	//Format prefix to show function name and log level and to indicate that timezone used is UTC
	customPrefix := fmt.Sprintf(logLevelFormatter, l.getCaller(), level)
	l.defaultLogger.Output(2, customPrefix+fmt.Sprint(args...))
}

func (l *DefaultLogger) logln(level Level, args ...interface{}) {
	//Format prefix to show function name and log level and to indicate that timezone used is UTC
	customPrefix := fmt.Sprintf(logLevelFormatter, l.getCaller(), level)
	l.defaultLogger.Output(2, customPrefix+fmt.Sprintln(args...))
}

func (l *DefaultLogger) getCaller() string {
	const MAXCALLERS = 5                  // search MAXCALLERS frames for the real caller
	const SKIPCALLERS = 4                 // skip SKIPCALLERS frames when determining the real caller
	const LOGPREFIX = "logging.(*Logger)" // LOGPREFIX indicates the upcoming frame contains the real caller and skip the frame
	const LOGBRIDGEPREFIX = "logbridge."  // LOGBRIDGEPREFIX indicates to skip the frame due to being a logbridge
	const NOTFOUND = "n/a"

	fpcs := make([]uintptr, MAXCALLERS)

	n := runtime.Callers(SKIPCALLERS, fpcs)
	if n == 0 {
		return NOTFOUND
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
			return funName
		}
	}

	return NOTFOUND
}
