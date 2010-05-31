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
	"runtime"
	"container/vector"
)

// Version information
const (
	L4G_VERSION = "log4go-v1.0.1"
	L4G_MAJOR   = 1
	L4G_MINOR   = 0
	L4G_BUILD   = 1
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
func (log *Logger) AddFilter(name string, level int, writer LogWriter) {
	if writer == nil || !writer.Good() {
		return
	}
	log.filterLevels[name] = level
	log.filterLogWriters[name] = writer
}

// Create a new logger with the standard stdout
func NewConsoleLogger(level int) *Logger {
	log := NewLogger()
	log.AddFilter("stdout", level, new(ConsoleLogWriter))
	return log
}

/******* Logging *******/

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

// Send a formatted log message easily
func (log *Logger) Logf(level int, format string, args ...interface{}) {
	log.intLogf(level, format, args)
}

// Utility for finest log messages
func (log *Logger) Finest(format string, args ...interface{}) {
	log.intLogf(FINEST, format, args)
}

// Utility for fine log messages
func (log *Logger) Fine(format string, args ...interface{}) {
	log.intLogf(FINE, format, args)
}

// Utility for debug log messages
func (log *Logger) Debug(format string, args ...interface{}) {
	log.intLogf(DEBUG, format, args)
}

// Utility for trace log messages
func (log *Logger) Trace(format string, args ...interface{}) {
	log.intLogf(TRACE, format, args)
}

// Utility for info log messages
func (log *Logger) Info(format string, args ...interface{}) {
	log.intLogf(INFO, format, args)
}

// Utility for warn log messages (returns an os.Error for easy function returns)
func (log *Logger) Warn(format string, args ...interface{}) os.Error {
	log.intLogf(WARNING, format, args)
	return os.NewError(fmt.Sprintf(format, args))
}

// Utility for error log messages (returns an os.Error for easy function returns)
func (log *Logger) Error(format string, args ...interface{}) os.Error {
	log.intLogf(ERROR, format, args)
	return os.NewError(fmt.Sprintf(format, args))
}

// Utility for critical log messages (returns an os.Error for easy function returns)
func (log *Logger) Critical(format string, args ...interface{}) os.Error {
	log.intLogf(CRITICAL, format, args)
	return os.NewError(fmt.Sprintf(format, args))
}
