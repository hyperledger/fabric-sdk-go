/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package modlog

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/metadata"
)

var rwmutex = &sync.RWMutex{}
var moduleLevels = &metadata.ModuleLevels{}
var callerInfos = &metadata.CallerInfo{}
var useCustomLogger int32

// default logger factory singleton
var loggerProviderInstance api.LoggerProvider
var loggerProviderOnce sync.Once

// Provider is the default logger implementation
type Provider struct {
}

//GetLogger returns SDK logger implementation
func (p *Provider) GetLogger(module string) api.Logger {
	newDefLogger := log.New(os.Stdout, fmt.Sprintf(logPrefixFormatter, module), log.Ldate|log.Ltime|log.LUTC)
	return &Log{deflogger: newDefLogger, module: module}
}

//LoggerProvider returns logging provider for SDK logger
func LoggerProvider() api.LoggerProvider {
	return &Provider{}
}

//InitLogger sets custom logger which will be used over deflogger.
//It is required to call this function before making any loggings.
func InitLogger(l api.LoggerProvider) {
	loggerProviderOnce.Do(func() {
		loggerProviderInstance = l
		atomic.StoreInt32(&useCustomLogger, 1)
	})
}

//Log is a standard SDK logger implementation
type Log struct {
	deflogger    *log.Logger
	customLogger api.Logger
	module       string
	custom       bool
	once         sync.Once
}

//LoggerOpts  for all logger customization options
type loggerOpts struct {
	levelEnabled      bool
	callerInfoEnabled bool
}

const (
	logLevelFormatter   = "UTC %s-> %4.4s "
	logPrefixFormatter  = " [%s] "
	callerInfoFormatter = "- %s "
)

//SetLevel - setting log level for given module
func SetLevel(module string, level api.Level) {
	rwmutex.Lock()
	defer rwmutex.Unlock()
	moduleLevels.SetLevel(module, level)
}

//GetLevel - getting log level for given module
func GetLevel(module string) api.Level {
	rwmutex.RLock()
	defer rwmutex.RUnlock()
	return moduleLevels.GetLevel(module)
}

//IsEnabledFor - Check if given log level is enabled for given module
func IsEnabledFor(module string, level api.Level) bool {
	rwmutex.RLock()
	defer rwmutex.RUnlock()
	return moduleLevels.IsEnabledFor(module, level)
}

//ShowCallerInfo - Show caller info in log lines for given log level
func ShowCallerInfo(module string, level api.Level) {
	rwmutex.Lock()
	defer rwmutex.Unlock()
	callerInfos.ShowCallerInfo(module, level)
}

//HideCallerInfo - Do not show caller info in log lines for given log level
func HideCallerInfo(module string, level api.Level) {
	rwmutex.Lock()
	defer rwmutex.Unlock()
	callerInfos.HideCallerInfo(module, level)
}

//getLoggerOpts - returns LoggerOpts which can be used for customization
func getLoggerOpts(module string, level api.Level) *loggerOpts {
	rwmutex.RLock()
	defer rwmutex.RUnlock()
	return &loggerOpts{
		levelEnabled:      moduleLevels.IsEnabledFor(module, level),
		callerInfoEnabled: callerInfos.IsCallerInfoEnabled(module, level),
	}
}

// Fatal is CRITICAL log followed by a call to os.Exit(1).
func (l *Log) Fatal(args ...interface{}) {
	opts := getLoggerOpts(l.module, api.CRITICAL)
	if l.loadCustomLogger() {
		l.customLogger.Fatal(args...)
		return
	}
	l.log(opts, api.CRITICAL, args...)
	l.deflogger.Fatal(args...)
}

// Fatalf is CRITICAL log formatted followed by a call to os.Exit(1).
func (l *Log) Fatalf(format string, args ...interface{}) {
	opts := getLoggerOpts(l.module, api.CRITICAL)
	if l.loadCustomLogger() {
		l.customLogger.Fatalf(format, args...)
		return
	}
	l.logf(opts, api.CRITICAL, format, args...)
	l.deflogger.Fatalf(format, args...)
}

// Fatalln is CRITICAL log ln followed by a call to os.Exit(1).
func (l *Log) Fatalln(args ...interface{}) {
	opts := getLoggerOpts(l.module, api.CRITICAL)
	if l.loadCustomLogger() {
		l.customLogger.Fatalln(args...)
		return
	}
	l.logln(opts, api.CRITICAL, args...)
	l.deflogger.Fatalln(args...)
}

// Panic is CRITICAL log followed by a call to panic()
func (l *Log) Panic(args ...interface{}) {
	opts := getLoggerOpts(l.module, api.CRITICAL)
	if l.loadCustomLogger() {
		l.customLogger.Panic(args...)
		return
	}
	l.log(opts, api.CRITICAL, args...)
	l.deflogger.Panic(args...)
}

// Panicf is CRITICAL log formatted followed by a call to panic()
func (l *Log) Panicf(format string, args ...interface{}) {
	opts := getLoggerOpts(l.module, api.CRITICAL)
	if l.loadCustomLogger() {
		l.customLogger.Panicf(format, args...)
		return
	}
	l.logf(opts, api.CRITICAL, format, args...)
	l.deflogger.Panicf(format, args...)
}

// Panicln is CRITICAL log ln followed by a call to panic()
func (l *Log) Panicln(args ...interface{}) {
	opts := getLoggerOpts(l.module, api.CRITICAL)
	if l.loadCustomLogger() {
		l.customLogger.Panicln(args...)
		return
	}
	l.logln(opts, api.CRITICAL, args...)
	l.deflogger.Panicln(args...)
}

// Print calls go log.Output.
// Arguments are handled in the manner of fmt.Print.
func (l *Log) Print(args ...interface{}) {
	if l.loadCustomLogger() {
		l.customLogger.Print(args...)
		return
	}
	l.deflogger.Print(args...)
}

// Printf calls go log.Output.
// Arguments are handled in the manner of fmt.Printf.
func (l *Log) Printf(format string, args ...interface{}) {
	if l.loadCustomLogger() {
		l.customLogger.Printf(format, args...)
		return
	}
	l.deflogger.Printf(format, args...)
}

// Println calls go log.Output.
// Arguments are handled in the manner of fmt.Println.
func (l *Log) Println(args ...interface{}) {
	if l.loadCustomLogger() {
		l.customLogger.Println(args...)
		return
	}
	l.deflogger.Println(args...)
}

// Debug calls go log.Output.
// Arguments are handled in the manner of fmt.Print.
func (l *Log) Debug(args ...interface{}) {
	opts := getLoggerOpts(l.module, api.DEBUG)
	if !opts.levelEnabled {
		return
	}
	if l.loadCustomLogger() {
		l.customLogger.Debug(args...)
		return
	}
	l.log(opts, api.DEBUG, args...)
}

// Debugf calls go log.Output.
// Arguments are handled in the manner of fmt.Printf.
func (l *Log) Debugf(format string, args ...interface{}) {
	opts := getLoggerOpts(l.module, api.DEBUG)
	if !opts.levelEnabled {
		return
	}
	if l.loadCustomLogger() {
		l.customLogger.Debugf(format, args...)
		return
	}
	l.logf(opts, api.DEBUG, format, args...)
}

// Debugln calls go log.Output.
// Arguments are handled in the manner of fmt.Println.
func (l *Log) Debugln(args ...interface{}) {
	opts := getLoggerOpts(l.module, api.DEBUG)
	if !opts.levelEnabled {
		return
	}
	if l.loadCustomLogger() {
		l.customLogger.Debugln(args...)
		return
	}
	l.logln(opts, api.DEBUG, args...)
}

// Info calls go log.Output.
// Arguments are handled in the manner of fmt.Print.
func (l *Log) Info(args ...interface{}) {
	opts := getLoggerOpts(l.module, api.INFO)
	if !opts.levelEnabled {
		return
	}
	if l.loadCustomLogger() {
		l.customLogger.Info(args...)
		return
	}
	l.log(opts, api.INFO, args...)
}

// Infof calls go log.Output.
// Arguments are handled in the manner of fmt.Printf.
func (l *Log) Infof(format string, args ...interface{}) {
	opts := getLoggerOpts(l.module, api.INFO)
	if !opts.levelEnabled {
		return
	}
	if l.loadCustomLogger() {
		l.customLogger.Infof(format, args...)
		return
	}
	l.logf(opts, api.INFO, format, args...)
}

// Infoln calls go log.Output.
// Arguments are handled in the manner of fmt.Println.
func (l *Log) Infoln(args ...interface{}) {
	opts := getLoggerOpts(l.module, api.INFO)
	if !opts.levelEnabled {
		return
	}
	if l.loadCustomLogger() {
		l.customLogger.Infoln(args...)
		return
	}
	l.logln(opts, api.INFO, args...)
}

// Warn calls go log.Output.
// Arguments are handled in the manner of fmt.Print.
func (l *Log) Warn(args ...interface{}) {
	opts := getLoggerOpts(l.module, api.WARNING)
	if !opts.levelEnabled {
		return
	}
	if l.loadCustomLogger() {
		l.customLogger.Warn(args...)
		return
	}
	l.log(opts, api.WARNING, args...)
}

// Warnf calls go log.Output.
// Arguments are handled in the manner of fmt.Printf.
func (l *Log) Warnf(format string, args ...interface{}) {
	opts := getLoggerOpts(l.module, api.WARNING)
	if !opts.levelEnabled {
		return
	}
	if l.loadCustomLogger() {
		l.customLogger.Warnf(format, args...)
		return
	}
	l.logf(opts, api.WARNING, format, args...)
}

// Warnln calls go log.Output.
// Arguments are handled in the manner of fmt.Println.
func (l *Log) Warnln(args ...interface{}) {
	opts := getLoggerOpts(l.module, api.WARNING)
	if !opts.levelEnabled {
		return
	}
	if l.loadCustomLogger() {
		l.customLogger.Warnln(args...)
		return
	}
	l.logln(opts, api.WARNING, args...)
}

// Error calls go log.Output.
// Arguments are handled in the manner of fmt.Print.
func (l *Log) Error(args ...interface{}) {
	opts := getLoggerOpts(l.module, api.ERROR)
	if !opts.levelEnabled {
		return
	}
	if l.loadCustomLogger() {
		l.customLogger.Error(args...)
		return
	}
	l.log(opts, api.ERROR, args...)
}

// Errorf calls go log.Output.
// Arguments are handled in the manner of fmt.Printf.
func (l *Log) Errorf(format string, args ...interface{}) {
	opts := getLoggerOpts(l.module, api.ERROR)
	if !opts.levelEnabled {
		return
	}
	if l.loadCustomLogger() {
		l.customLogger.Errorf(format, args...)
		return
	}
	l.logf(opts, api.ERROR, format, args...)
}

// Errorln calls go log.Output.
// Arguments are handled in the manner of fmt.Println.
func (l *Log) Errorln(args ...interface{}) {
	opts := getLoggerOpts(l.module, api.ERROR)
	if !opts.levelEnabled {
		return
	}
	if l.loadCustomLogger() {
		l.customLogger.Errorln(args...)
		return
	}
	l.logln(opts, api.ERROR, args...)
}

//ChangeOutput for changing output destination for the logger.
func (l *Log) ChangeOutput(output io.Writer) {
	l.deflogger.SetOutput(output)
}

func (l *Log) logf(opts *loggerOpts, level api.Level, format string, args ...interface{}) {
	//Format prefix to show function name and log level and to indicate that timezone used is UTC
	customPrefix := fmt.Sprintf(logLevelFormatter, l.getCallerInfo(opts), metadata.ParseString(level))
	err := l.deflogger.Output(2, customPrefix+fmt.Sprintf(format, args...))
	if err != nil {
		fmt.Printf("error from deflogger.Output %v\n", err)
	}
}

func (l *Log) log(opts *loggerOpts, level api.Level, args ...interface{}) {
	//Format prefix to show function name and log level and to indicate that timezone used is UTC
	customPrefix := fmt.Sprintf(logLevelFormatter, l.getCallerInfo(opts), metadata.ParseString(level))
	err := l.deflogger.Output(2, customPrefix+fmt.Sprint(args...))
	if err != nil {
		fmt.Printf("error from deflogger.Output %v\n", err)
	}
}

func (l *Log) logln(opts *loggerOpts, level api.Level, args ...interface{}) {
	//Format prefix to show function name and log level and to indicate that timezone used is UTC
	customPrefix := fmt.Sprintf(logLevelFormatter, l.getCallerInfo(opts), metadata.ParseString(level))
	err := l.deflogger.Output(2, customPrefix+fmt.Sprintln(args...))
	if err != nil {
		fmt.Printf("error from deflogger.Output %v\n", err)
	}
}

func (l *Log) loadCustomLogger() bool {
	l.once.Do(func() {
		if atomic.LoadInt32(&useCustomLogger) > 0 {
			l.customLogger = loggerProviderInstance.GetLogger(l.module)
			l.custom = true
		}
	})
	return l.custom
}

func (l *Log) getCallerInfo(opts *loggerOpts) string {

	if !opts.callerInfoEnabled {
		return ""
	}

	const MAXCALLERS = 6  // search MAXCALLERS frames for the real caller
	const SKIPCALLERS = 3 // skip SKIPCALLERS frames when determining the real caller
	const NOTFOUND = "n/a"

	fpcs := make([]uintptr, MAXCALLERS)

	n := runtime.Callers(SKIPCALLERS, fpcs)
	if n == 0 {
		return fmt.Sprintf(callerInfoFormatter, NOTFOUND)
	}

	frames := runtime.CallersFrames(fpcs[:n])
	loggerFrameFound := false
	for f, more := frames.Next(); more; f, more = frames.Next() {
		pkgPath, fnName := filepath.Split(f.Function)

		if f.Func == nil || f.Function == "" {
			fnName = NOTFOUND // not a function or unknown
		}

		if hasLoggerFnPrefix(pkgPath, fnName) {
			loggerFrameFound = true

		} else if loggerFrameFound {
			return fmt.Sprintf(callerInfoFormatter, fnName)
		}
	}

	return fmt.Sprintf(callerInfoFormatter, NOTFOUND)
}

func hasLoggerFnPrefix(pkgPath string, fnName string) bool {
	const (
		loggingAPIPath = "github.com/hyperledger/fabric-sdk-go/pkg/core/logging/"
		loggingAPIPkg  = "api" // Go < 1.12
		modlogFnPrefix = "modlog.(*Log)."
		loggingPath    = "github.com/hyperledger/fabric-sdk-go/pkg/common/"
		loggingPkg     = "logging"
		logBridgePath  = "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/sdkpatch/logbridge"
		logBridgePkg   = "logbridge"
	)

	switch pkgPath {
	case loggingAPIPath:
		return strings.HasPrefix(fnName, modlogFnPrefix) || strings.HasPrefix(fnName, loggingAPIPkg)
	case loggingPath:
		return strings.HasPrefix(fnName, loggingPkg)
	case logBridgePath:
		return strings.HasPrefix(fnName, logBridgePkg)
	}

	return false
}
