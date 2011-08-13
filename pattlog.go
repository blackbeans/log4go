// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"fmt"
	"bytes"
	"time"
	"io"
	"sync"
)

const (
	FORMAT_DEFAULT = "[%D %T] [%L] (%S) %M"
	FORMAT_SHORT   = "[%t %d] [%L] %M"
	FORMAT_ABBREV  = "[%L] %M"
)

var (
	// TimeConversionFunction specifies what function to call to
	// convert from seconds into a *time.Time.  Change this to
	// time.SecondsToUTC for UTC stamped logs
	TimeConversionFunction = time.SecondsToLocalTime
)

// This could really use Dmitry's "co" package's thread-local storage.
var formatCache = struct {
	*sync.Mutex
	buf               *bytes.Buffer
	tm                *time.Time
	LastUpdateSeconds int64
	t, T, d, D        string
}{
	Mutex: new(sync.Mutex),
	buf:   new(bytes.Buffer),
}

// Known format codes:
// %T - Time (15:04:05 MST)
// %t - Time (15:04)
// %D - Date (2006/01/02)
// %d - Date (01/02/06)
// %L - Level (FNST, FINE, DEBG, TRAC, WARN, EROR, CRIT)
// %S - Source
// %M - Message
// Ignores unknown formats
// Recommended: "[%D %T] [%L] (%S) %M"
func FormatLogRecord(format string, rec *LogRecord) string {
	if rec == nil {
		return "<nil>"
	}
	if len(format) == 0 {
		return ""
	}

	formatCache.Lock()
	defer formatCache.Unlock()

	out := formatCache.buf
	out.Truncate(0)

	secs := rec.Created / 1e9
	stale := secs != formatCache.LastUpdateSeconds

	if stale {
		formatCache.tm = TimeConversionFunction(secs)
	}
	tm := formatCache.tm

	// Split the string into pieces by % signs
	pieces := bytes.Split([]byte(format), []byte{'%'})

	// Iterate over the pieces, replacing known formats
	for i, piece := range pieces {
		if i > 0 && len(piece) > 0 {
			switch piece[0] {
			case 'T':
				if stale || formatCache.T == "" {
					formatCache.T = fmt.Sprintf("%02d:%02d:%02d %s", tm.Hour, tm.Minute, tm.Second, tm.Zone)
				}
				out.WriteString(formatCache.T)
			case 't':
				if stale || formatCache.t == "" {
					formatCache.t = fmt.Sprintf("%02d:%02d", tm.Hour, tm.Minute)
				}
				out.WriteString(formatCache.t)
			case 'D':
				if stale || formatCache.D == "" {
					formatCache.D = fmt.Sprintf("%04d/%02d/%02d", tm.Year, tm.Month, tm.Day)
				}
				out.WriteString(formatCache.D)
			case 'd':
				if stale || formatCache.d == "" {
					formatCache.d = fmt.Sprintf("%02d/%02d/%02d", tm.Month, tm.Day, tm.Year%100)
				}
				out.WriteString(formatCache.d)
			case 'L':
				out.WriteString(levelStrings[rec.Level])
			case 'S':
				out.WriteString(rec.Source)
			case 'M':
				out.WriteString(rec.Message)
			}
			if len(piece) > 1 {
				out.Write(piece[1:])
			}
		} else if len(piece) > 0 {
			out.Write(piece)
		}
	}
	out.WriteByte('\n')

	return out.String()
}

// This is the standard writer that prints to standard output.
type FormatLogWriter chan *LogRecord

// This creates a new FormatLogWriter
func NewFormatLogWriter(out io.Writer, format string) FormatLogWriter {
	records := make(FormatLogWriter, LogBufferLength)
	go records.run(out, format)
	return records
}

func (w FormatLogWriter) run(out io.Writer, format string) {
	for rec := range w {
		fmt.Fprint(out, FormatLogRecord(format, rec))
	}
}

// This is the FormatLogWriter's output method.  This will block if the output
// buffer is full.
func (w FormatLogWriter) LogWrite(rec *LogRecord) {
	w <- rec
}

// Close stops the logger from sending messages to standard output.  Attempts to
// send log messages to this logger after a Close have undefined behavior.
func (w FormatLogWriter) Close() {
	close(w)
}
