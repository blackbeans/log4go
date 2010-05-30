// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
package elog

import (
	"os"
	"fmt"
	"time"
	"runtime"
	"container/vector"
)

// Version information
const (
	ELOG_VERSION = "eLog-v2.0.0"
	ELOG_MAJOR   = 2
	ELOG_MINOR   = 0
	ELOG_BUILD   = 0
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
	LogWrite(rec *LogRecord) (n int, err os.Error)
	Good() bool
}

// This is the standard writer that prints to standard output
type ConsoleLogWriter struct{}

// This creates a new ConsoleLogWriter
func NewConsoleLogWriter() *ConsoleLogWriter { return new(ConsoleLogWriter) }

// This is the ConsoleLogWriter's output method
func (slw *ConsoleLogWriter) LogWrite(rec *LogRecord) (n int, err os.Error) {
	return fmt.Fprint(os.Stdout, "[", rec.Created.Format("01/02/06 15:04:05"), "] [", levelStrings[rec.Level], "] ", rec.Message, "\n")
}

// The standard output logger should always be writable
func (slw *ConsoleLogWriter) Good() bool { return true }

/****** Logger ******/

// If LogRecord is the blood of the package, is the heart.
type Logger struct {
	// All filters have an entry in each of the following
	filterLevels     map[string]int
	filterLogWriters map[string]LogWriter
}

// Create a new logger
func NewLogger() *Logger {
	l := new(Logger)
	l.filterLevels = make(map[string]int)
	l.filterLogWriters = make(map[string]LogWriter)
	return l
}

// Add the standard filter
func (l *Logger) AddFilter(name string, level int, writer LogWriter) {
	if writer == nil || !writer.Good() {
		return
	}
	l.filterLevels[name] = level
	l.filterLogWriters[name] = writer
}

// Create a new logger with the standard stdout
func NewConsoleLogger(level int) *Logger {
	l := NewLogger()
	l.AddFilter("stdout", level, new(ConsoleLogWriter))
	return l
}

// Send a log message manually
func (l *Logger) Log(level int, source, message string) {
	// Create a vector long enough to not require resizing
	var logto vector.StringVector
	logto.Resize(0, len(l.filterLevels))

	// Determine if any logging will be done
	for filt := range l.filterLevels {
		if level >= l.filterLevels[filt] {
			logto.Push(filt)
		}
	}

	// Only log if a filter requires it
	if len(logto) > 0 {
		// Make the log record
		rec := newLogRecord(level, source, message)

		// Dispatch the logs
		for _,filt := range logto {
			l.filterLogWriters[filt].LogWrite(rec)
		}
	}
}

// Send a formatted log message easily
func (l *Logger) intLogf(level int, format string, args ...interface{}) {
	// Create a vector long enough to not require resizing
	var logto vector.StringVector
	logto.Resize(0, len(l.filterLevels))

	// Determine if any logging will be done
	for filt := range l.filterLevels {
		if level >= l.filterLevels[filt] {
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
			l.filterLogWriters[filt].LogWrite(rec)
		}
	}
}

// Send a formatted log message easily
func (l *Logger) Logf(level int, format string, args ...interface{}) {
	l.intLogf(level, format, args)
}

// Utility for finest log messages
func (l *Logger) Finest(format string, args ...interface{}) {
	l.intLogf(FINEST, format, args)
}

// Utility for fine log messages
func (l *Logger) Fine(format string, args ...interface{}) {
	l.intLogf(FINE, format, args)
}

// Utility for debug log messages
func (l *Logger) Debug(format string, args ...interface{}) {
	l.intLogf(DEBUG, format, args)
}

// Utility for trace log messages
func (l *Logger) Trace(format string, args ...interface{}) {
	l.intLogf(TRACE, format, args)
}

// Utility for info log messages
func (l *Logger) Info(format string, args ...interface{}) {
	l.intLogf(INFO, format, args)
}

// Utility for warn log messages (returns an os.Error for easy function returns)
func (l *Logger) Warn(format string, args ...interface{}) os.Error {
	l.intLogf(WARNING, format, args)
	return os.NewError(fmt.Sprintf(format, args))
}

// Utility for error log messages (returns an os.Error for easy function returns)
func (l *Logger) Error(format string, args ...interface{}) os.Error {
	l.intLogf(ERROR, format, args)
	return os.NewError(fmt.Sprintf(format, args))
}

// Utility for critical log messages (returns an os.Error for easy function returns)
func (l *Logger) Critical(format string, args ...interface{}) os.Error {
	l.intLogf(CRITICAL, format, args)
	return os.NewError(fmt.Sprintf(format, args))
}
