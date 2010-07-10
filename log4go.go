// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

// Enhanced Logging
//
// This is inspired by the logging functionality in Java.  Essentially, you create a Logger
// object and create output filters for it.  You can send whatever you want to the Logger,
// and it will filter that based on your settings and send it to the outputs.  This way, you
// can put as much debug code in your program as you want, and when you're done you can filter
// out the mundane messages so only the import ones show up.
//
// Utility functions are provided to make life easier. Here is some example code to get started:
//
// log := elog.NewLogger()
// log.AddFilter("stdout", elog.DEBUG, new(elog.ConsoleLogWriter))
// log.AddFilter("log",    elog.FINE,  elog.NewFileLogWriter("example.log", true))
// log.Info("The time is now: %s", time.LocalTime().Format("15:04:05 MST 2006/01/02"))
//
// The first two lines can be combined with the utility NewConsoleLogger:
//
// log := elog.NewConsoleLogger(elog.DEBUG)
// log.AddFilter("log",    elog.FINE,  elog.NewFileLogWriter("example.log", true))
// log.Info("The time is now: %s", time.LocalTime().Format("15:04:05 MST 2006/01/02"))
//
// Usage notes:
// - The ConsoleLogWriter does not display the source to standard output, but the FileLogWriter does.
// - The utility functions (Info, Debug, Warn, etc) derive their source from the calling function
//
// Future work: (please let me know if you think I should work on any of these particularly)
// - Log file rotation
// - Logging configuration files ala log4j
// - Have the ability to remove filters?
// - Have GetInfoChannel, GetDebugChannel, etc return a chan string that allows for another method of logging
// - Add an XML filter type
package log4go

import (
	"os"
	"fmt"
	"time"
	"strings"
	"runtime"
	"container/vector"
)

// Version information
const (
	L4G_VERSION = "log4go-v2.0.2"
	L4G_MAJOR   = 2
	L4G_MINOR   = 0
	L4G_BUILD   = 2
)

/****** Constants ******/

// These are the integer logging levels used by the logger
const (
	FINEST = iota
	FINE
	DEBUG
	TRACE
	INFO
	WARNING
	ERROR
	CRITICAL
)

// Logging level strings
var (
	levelStrings = [...]string{"FNST", "FINE", "DEBG", "TRAC", "INFO", "WARN", "EROR", "CRIT"}
)

/****** LogRecord ******/

// This is the lifeblood of the package; it contains all of the pertinent information for each message
type LogRecord struct {
	Level   int        // The log level
	Created *time.Time // The time at which the log message was created
	Source  string     // The message source
	Message string     // The log message
}

func newLogRecord(lv int, src string, msg string) *LogRecord {
	lr := new(LogRecord)
	lr.Created = time.LocalTime()
	lr.Level = lv
	lr.Source = src
	lr.Message = msg
	return lr
}

/****** LogWriter ******/

// This is an interface for anything that should be able to write logs
type LogWriter interface {
	// This will be called to log a LogRecord message.
	// If necessary. this function should be *INTERNALLY* synchronzied,
	// and should spawn a separate goroutine if it could hang the program or take a long time.
	// TODO: This may be changed to have an Init() call that returns a
	// channel similar to <-chan *LogRecord for a more go-like internal setup
	LogWrite(rec *LogRecord) (n int, err os.Error)

	// This should return, at any given time, if the LogWriter is still in a good state.
	// A good state is defined as having the ability to dispatch a log message immediately.
	// if a LogWriter is not in a good state, the log message is simply not dispatched.
	Good() bool

	// This should clean up anything lingering about the LogWriter, as it is called before
	// the LogWriter is removed.  If possible, this should guarantee that all LogWrites
	// have been completed.
	Close()
}

/****** Logger ******/

// If LogRecord is the blood of the package, is the heart.
type Logger struct {
	// All filters have an entry in each of the following
	filterLevels     map[string]int
	filterLogWriters map[string]LogWriter
}

// Create a new logger
func NewLogger() *Logger {
	log := new(Logger)
	log.filterLevels = make(map[string]int)
	log.filterLogWriters = make(map[string]LogWriter)
	return log
}

// Closes all log writers in preparation for exiting the program.
// Calling this is not really imperative, unless you want to guarantee that all log messages are written.
func (log *Logger) Close() {
	// Close all open loggers
	for key := range log.filterLogWriters {
		log.filterLogWriters[key].Close()
		log.filterLogWriters[key] = nil, false
		log.filterLevels[key] = 0, false
	}
}

// Add the standard filter.
// This function is NOT INTERNALLY THREAD SAFE.  If you plan on
// calling this function from multiple goroutines, you will want
// to synchronize it yourself somehow.
// Returns self for chaining
func (log *Logger) AddFilter(name string, level int, writer LogWriter) *Logger {
	if writer == nil || !writer.Good() {
		return nil
	}
	log.filterLevels[name] = level
	log.filterLogWriters[name] = writer
	return log
}

// Create a new logger with the standard stdout
func NewConsoleLogger(level int) *Logger {
	log := NewLogger()
	log.AddFilter("stdout", level, new(ConsoleLogWriter))
	return log
}

/******* Logging *******/
// Send a formatted log message internally
func (log *Logger) intLogf(level int, format string, args ...interface{}) {
	// Create a vector long enough to not require resizing
	var logto vector.StringVector
	logto.Resize(0, len(log.filterLevels))

	// Determine if any logging will be done
	for filt := range log.filterLevels {
		if level >= log.filterLevels[filt] {
			logto.Push(filt)
		}
	}

	// Only log if a filter requires it
	if len(logto) > 0 {
		// Determine caller func
		pc, _, lineno, ok := runtime.Caller(2)
		src := ""
		if ok {
			src = fmt.Sprintf("%s:%d", runtime.FuncForPC(pc).Name(), lineno)
		}

		// Make the log record
		rec := newLogRecord(level, src, fmt.Sprintf(format, args))

		// Dispatch the logs
		for _,filt := range logto {
			log.filterLogWriters[filt].LogWrite(rec)
		}
	}
}

// Send a closure log message internally
func (log *Logger) intLogc(level int, closure func()string) {
	// Create a vector long enough to not require resizing
	var logto vector.StringVector
	logto.Resize(0, len(log.filterLevels))

	// Determine if any logging will be done
	for filt := range log.filterLevels {
		if level >= log.filterLevels[filt] {
			logto.Push(filt)
		}
	}

	// Only log if a filter requires it
	if len(logto) > 0 {
		// Determine caller func
		pc, _, lineno, ok := runtime.Caller(2)
		src := ""
		if ok {
			src = fmt.Sprintf("%s:%d", runtime.FuncForPC(pc).Name(), lineno)
		}

		// Make the log record from the closure's return
		rec := newLogRecord(level, src, closure())

		// Dispatch the logs
		for _,filt := range logto {
			log.filterLogWriters[filt].LogWrite(rec)
		}
	}
}

// Send a log message manually
func (log *Logger) Log(level int, source, message string) {
	// Create a vector long enough to not require resizing
	var logto vector.StringVector
	logto.Resize(0, len(log.filterLevels))

	// Determine if any logging will be done
	for filt := range log.filterLevels {
		if level >= log.filterLevels[filt] {
			logto.Push(filt)
		}
	}

	// Only log if a filter requires it
	if len(logto) > 0 {
		// Make the log record
		rec := newLogRecord(level, source, message)

		// Dispatch the logs
		for _,filt := range logto {
			lw := log.filterLogWriters[filt]
			if lw.Good() {
				lw.LogWrite(rec)
			}
		}
	}
}

// Send a formatted log message easily
func (log *Logger) Logf(level int, format string, args ...interface{}) {
	log.intLogf(level, format, args)
}

// Send a closure log message
func (log *Logger) Logc(level int, closure func()string) {
	log.intLogc(level, closure)
}

// Utility for finest log messages (see Debug() for parameter explanation)
func (log *Logger) Finest(arg0 interface{}, args ...interface{}) {
	const (
		lvl = FINEST
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		log.intLogf(lvl, first, args)
	case func()string:
		// Log the closure (no other arguments used)
		log.intLogc(lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		log.intLogf(lvl, fmt.Sprint(arg0) + strings.Repeat(" %v", len(args)), args)
	}
}

// Utility for fine log messages (see Debug() for parameter explanation)
func (log *Logger) Fine(arg0 interface{}, args ...interface{}) {
	const (
		lvl = FINE
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		log.intLogf(lvl, first, args)
	case func()string:
		// Log the closure (no other arguments used)
		log.intLogc(lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		log.intLogf(lvl, fmt.Sprint(arg0) + strings.Repeat(" %v", len(args)), args)
	}
}

// Utility for debug log messages
// When given a string as the first argument, this behaves like Logf but with the DEBUG log level (e.g. the first argument is interpreted as a format for the latter arguments)
// When given a closure of type func()string, this logs the string returned by the closure iff it will be logged.  The closure runs at most one time.
// When given anything else, the log message will be each of the arguments formatted with %v and separated by spaces (ala Sprint).
func (log *Logger) Debug(arg0 interface{}, args ...interface{}) {
	const (
		lvl = DEBUG
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		log.intLogf(lvl, first, args)
	case func()string:
		// Log the closure (no other arguments used)
		log.intLogc(lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		log.intLogf(lvl, fmt.Sprint(arg0) + strings.Repeat(" %v", len(args)), args)
	}
}

// Utility for trace log messages (see Debug() for parameter explanation)
func (log *Logger) Trace(arg0 interface{}, args ...interface{}) {
	const (
		lvl = TRACE
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		log.intLogf(lvl, first, args)
	case func()string:
		// Log the closure (no other arguments used)
		log.intLogc(lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		log.intLogf(lvl, fmt.Sprint(arg0) + strings.Repeat(" %v", len(args)), args)
	}
}

// Utility for info log messages (see Debug() for parameter explanation)
func (log *Logger) Info(arg0 interface{}, args ...interface{}) {
	const (
		lvl = INFO
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		log.intLogf(lvl, first, args)
	case func()string:
		// Log the closure (no other arguments used)
		log.intLogc(lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		log.intLogf(lvl, fmt.Sprint(arg0) + strings.Repeat(" %v", len(args)), args)
	}
}

// Utility for warn log messages (returns an os.Error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return
func (log *Logger) Warn(arg0 interface{}, args ...interface{}) os.Error {
	const (
		lvl = WARNING
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		log.intLogf(lvl, first, args)
		return os.NewError(fmt.Sprintf(first, args))
	case func()string:
		// Log the closure (no other arguments used)
		str := first()
		log.intLogf(lvl, "%s", str)
		return os.NewError(str)
	default:
		// Build a format string so that it will be similar to Sprint
		log.intLogf(lvl, fmt.Sprint(first) + strings.Repeat(" %v", len(args)), args)
		return os.NewError(fmt.Sprint(first) + fmt.Sprintf(strings.Repeat(" %v", len(args)), args))
	}
	return nil
}

// Utility for error log messages (returns an os.Error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return
func (log *Logger) Error(arg0 interface{}, args ...interface{}) os.Error {
	const (
		lvl = ERROR
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		log.intLogf(lvl, first, args)
		return os.NewError(fmt.Sprintf(first, args))
	case func()string:
		// Log the closure (no other arguments used)
		str := first()
		log.intLogf(lvl, "%s", str)
		return os.NewError(str)
	default:
		// Build a format string so that it will be similar to Sprint
		log.intLogf(lvl, fmt.Sprint(first) + strings.Repeat(" %v", len(args)), args)
		return os.NewError(fmt.Sprint(first) + fmt.Sprintf(strings.Repeat(" %v", len(args)), args))
	}
	return nil
}

// Utility for critical log messages (returns an os.Error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return
func (log *Logger) Critical(arg0 interface{}, args ...interface{}) os.Error {
	const (
		lvl = CRITICAL
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		log.intLogf(lvl, first, args)
		return os.NewError(fmt.Sprintf(first, args))
	case func()string:
		// Log the closure (no other arguments used)
		str := first()
		log.intLogf(lvl, "%s", str)
		return os.NewError(str)
	default:
		// Build a format string so that it will be similar to Sprint
		log.intLogf(lvl, fmt.Sprint(first) + strings.Repeat(" %v", len(args)), args)
		return os.NewError(fmt.Sprint(first) + fmt.Sprintf(strings.Repeat(" %v", len(args)), args))
	}
	return nil
}
