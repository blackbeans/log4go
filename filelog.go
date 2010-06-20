// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"os"
	"fmt"
	"time"
	"sync"
)

// This log writer sends output to a file
type FileLogWriter struct {
	// thread safety (TODO: goroutine+chans instead?)
	lock *sync.Mutex

	// The opened file
	filename string
	file  *os.File

	// The logging format
	format string

	// Rotate at linecount
	maxlines int
	maxlines_curlines int

	// Rotate at size
	maxsize int
	maxsize_cursize int

	// Rotate daily
	daily bool
	daily_opendate int

	// Keep old logfiles (.001, .002, etc)
	rotate bool
}

// This is the FileLogWriter's output method
func (flw *FileLogWriter) LogWrite(rec *LogRecord) (n int, err os.Error) {
	flw.lock.Lock()
	defer flw.lock.Unlock()

	// First, check if we've gone over any of our rotate triggers
	if flw.Good() {
		if (flw.maxlines > 0 && flw.maxlines_curlines >= flw.maxlines) ||
			(flw.maxsize > 0 && flw.maxsize_cursize >= flw.maxsize) ||
			(flw.daily && time.LocalTime().Day != flw.daily_opendate) {
			flw.intRotate()
		}
	}

	// Make sure the writer is (still) good
	if !flw.Good() {
		return -1, os.NewError("File was not opened successfully")
	}

	// Perform the write
	n,err = flw.file.Write([]byte(FormatLogRecord(flw.format, rec)))

	// Update the counts
	if err == nil {
		flw.maxlines_curlines++
		flw.maxsize_cursize += n
	}

	return n, err
}

func (flw *FileLogWriter) Good() bool {
	return flw.file != nil
}

func (flw *FileLogWriter) Close() {
	flw.file.Close()
	flw.file = nil
}

func NewFileLogWriter(fname string, rotate bool) *FileLogWriter {
	flw := new(FileLogWriter)
	flw.lock = new(sync.Mutex)
	flw.filename = fname
	flw.format = "[%D %T] [%L] (%S) %M"
	flw.file = nil
	flw.rotate = rotate

	flw.intRotate() // open the file for the first time

	return flw
}

// Request that the logs rotate
func (flw *FileLogWriter) Rotate() {
	flw.lock.Lock()
	defer flw.lock.Unlock()
	flw.intRotate()
}

// If this is called in a threaded context, it MUST be synchronized
func (flw *FileLogWriter) intRotate() {
	// Close any log file that may be open
	if flw.file != nil {
		flw.file.Close()
	}

	// If we are keeping log files, move it to the next available number
	if flw.rotate {
		_, err := os.Lstat(flw.filename)
		if err == nil { // file exists
			//fmt.Fprintf(os.Stderr, "FileLogWriter.intRotate: file %s exists, searching for extension\n", flw.filename)
			// Find the next available number
			num := 1
			fname := ""
			for ; err == nil && num < 999; num++ {
				fname = flw.filename + fmt.Sprintf(".%03d", num)
				_, err = os.Lstat(fname)
			}
			if err != nil {
				// Rename the file to its newfound home
				//fmt.Fprintf(os.Stderr, "FileLogWriter.intRotate: Moving %s to %s\n", flw.filename, fname)
				os.Rename(flw.filename, fname)
			} else {
				//fmt.Fprintf(os.Stderr, "FileLogWriter.intRotate: Cannot find free log number to rename %s\n", flw.filename)
			}
		}
	}

	// Open the log file
	fd, err := os.Open(flw.filename, os.O_WRONLY|os.O_APPEND|os.O_CREAT, 0660)
	if err != nil {
		//fmt.Fprintf(os.Stderr, "FileLogWriter.intRotate: %s\n", err)
		return
	}
	flw.file = fd

	// Set the daily open date to the current date
	flw.daily_opendate = time.LocalTime().Day

	// initialize rotation values
	flw.maxlines_curlines = 0
	flw.maxsize_cursize = 0
}

// Set the logging format (chainable)
func (flw *FileLogWriter) SetFormat(format string) *FileLogWriter {
	//fmt.Fprintf(os.Stderr, "FileLogWriter.SetFormat: %v\n", format)
	flw.format = format
	return flw
}

// Set rotate at linecount (chainable)
func (flw *FileLogWriter) SetRotateLines(maxlines int) *FileLogWriter {
	//fmt.Fprintf(os.Stderr, "FileLogWriter.SetRotateLines: %v\n", maxlines)
	flw.maxlines = maxlines
	return flw
}

// Set rotate at size (chainable)
func (flw *FileLogWriter) SetRotateSize(maxsize int) *FileLogWriter {
	//fmt.Fprintf(os.Stderr, "FileLogWriter.SetRotateSize: %v\n", maxsize)
	flw.maxsize = maxsize
	return flw
}

// Set rotate daily (chainable)
func (flw *FileLogWriter) SetRotateDaily(daily bool) *FileLogWriter {
	//fmt.Fprintf(os.Stderr, "FileLogWriter.SetRotateDaily: %v\n", daily)
	flw.daily = daily
	return flw
}

// Set keep old (chainable)
func (flw *FileLogWriter) SetRotate(rotate bool) *FileLogWriter {
	//fmt.Fprintf(os.Stderr, "FileLogWriter.SetRotate: %v\n", rotate)
	flw.rotate = rotate
	return flw
}
