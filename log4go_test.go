// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"os"
	"fmt"
	"runtime"
	"io/ioutil"
	"crypto/md5"
	"encoding/hex"
	t "testing"
)

func TestELog(test *t.T) {
	//func newLogRecord(lv int, src string, msg string) *LogRecord {}
	fmt.Printf("Testing %s", L4G_VERSION)
	lr := newLogRecord(CRITICAL, "source", "message")
	if lr.Level != CRITICAL {
		test.Errorf("Incorrect level: %d should be %d", lr.Level, CRITICAL)
	}
	if lr.Source != "source" {
		test.Errorf("Incorrect source: %s should be %s", lr.Source, "source")
	}
	if lr.Message != "message" {
		test.Errorf("Incorrect message: %s should be %s", lr.Source, "message")
	}
}

func TestConsoleLogWriter(test *t.T) {
	slw := NewConsoleLogWriter()
	rec := newLogRecord(CRITICAL, "source", "message")

	if slw == nil {
		test.Fatalf("Invalid return: slw should not be nil")
	}

	//func (slw *ConsoleLogWriter) LogWrite(rec *LogRecord) (n int, err os.Error) { }
	if n, err := slw.LogWrite(rec); n != 35 && err == nil {
		test.Errorf("Invalid return: slw.LogWrite returned (%d,%s)", n, err)
	}

	//func (slw *ConsoleLogWriter) Good() bool { return true }
	if slw.Good() == false {
		test.Fatalf("Invalid return: slw should always be good")
	}
}

func TestFileLogWriter(test *t.T) {
	//func NewFileLogWriter(fname string, append bool) *FileLogWriter {}
	flw := NewFileLogWriter("logtest.log", false)
	rec := newLogRecord(CRITICAL, "source", "message")

	if flw == nil {
		test.Fatalf("Invalid return: flw should not be nil")
	}

	//func (flw *FileLogWriter) Good() bool {}
	if flw.Good() == false {
		test.Fatalf("Invalid return: flw should be Good")
	}

	//func (flw *FileLogWriter) LogWrite(rec *LogRecord) (n int, err os.Error) {}
	if n, err := flw.LogWrite(rec); n != 50 && err == nil {
		test.Fatalf("Invalid return: flw.LogWrite returned (%d,%s)", n, err)
	}

	//func (flw *FileLogWriter) Close() {}
	flw.Close()

	// delete the file
	os.Remove("logtest.log")
}

func TestXMLLogWriter(test *t.T) {
	//func NewXMLLogWriter(fname string, append bool) *XMLLogWriter {}
	xlw := NewXMLLogWriter("logtest.log", false)
	rec := newLogRecord(CRITICAL, "source", "message")

	if xlw == nil {
		test.Fatalf("Invalid return: xlw should not be nil")
	}

	//func (xlw *XMLLogWriter) Good() bool {}
	if xlw.Good() == false {
		test.Fatalf("Invalid return: xlw should be Good")
	}

	//func (xlw *XMLLogWriter) LogWrite(rec *LogRecord) (n int, err os.Error) {}
	if n, err := xlw.LogWrite(rec); n != 139 && err == nil {
		test.Fatalf("Invalid return: xlw.LogWrite returned (%d,%s)", n, err)
	}

	//func (xlw *XMLLogWriter) Close() {}
	xlw.Close()

	// delete the file
	os.Remove("logtest.log")
}

func TestLogger(test *t.T) {
	//func NewLogger() *Logger {}
	l := NewLogger()
	if l == nil {
		test.Fatalf("NewLogger should never return nil")
	}

	//func NewConsoleLogger(level int) *Logger {}
	sl := NewConsoleLogger(WARNING)
	if sl == nil {
		test.Fatalf("NewConsoleLogger should never return nil")
	}
	if lw, exist := sl.filterLogWriters["stdout"]; lw == nil || exist != true {
		test.Fatalf("NewConsoleLogger produced invalid logger (DNE or nil)")
	}
	if sl.filterLevels["stdout"] != WARNING {
		test.Fatalf("NewConsoleLogger produced invalid logger (incorrect level)")
	}
	if len(sl.filterLevels) != 1 || len(sl.filterLogWriters) != 1 {
		test.Fatalf("NewConsoleLogger produced invalid logger (incorrect map count)")
	}

	//func (l *Logger) AddFilter(name string, level int, writer LogWriter) {}
	l.AddFilter("stdout", DEBUG, NewConsoleLogWriter())
	if lw, exist := l.filterLogWriters["stdout"]; lw == nil || exist != true {
		test.Fatalf("AddFilter produced invalid logger (DNE or nil)")
	}
	if l.filterLevels["stdout"] != DEBUG {
		test.Fatalf("AddFilter produced invalid logger (incorrect level)")
	}
	if len(l.filterLevels) != 1 || len(l.filterLogWriters) != 1 {
		test.Fatalf("AddFilter produced invalid logger (incorrect map count)")
	}

	//func (l *Logger) Warn(format string, args ...interface{}) os.Error {}
	if err := l.Warn("%s %d %#v", "Warning:", 1, []int{}); err.String() != "Warning: 1 []int{}" {
		test.Errorf("Warn returned invalid error: %s", err)
	}

	//func (l *Logger) Error(format string, args ...interface{}) os.Error {}
	if err := l.Error("%s %d %#v", "Error:", 10, []string{}); err.String() != "Error: 10 []string{}" {
		test.Errorf("Error returned invalid error: %s", err)
	}

	//func (l *Logger) Critical(format string, args ...interface{}) os.Error {}
	if err := l.Critical("%s %d %#v", "Critical:", 100, []int64{}); err.String() != "Critical: 100 []int64{}" {
		test.Errorf("Critical returned invalid error: %s", err)
	}

	// Already tested or basically untestable
	//func (l *Logger) Log(level int, source, message string) {}
	//func (l *Logger) Logf(level int, format string, args ...interface{}) {}
	//func (l *Logger) intLogf(level int, format string, args ...interface{}) string {}
	//func (l *Logger) Finest(format string, args ...interface{}) {}
	//func (l *Logger) Fine(format string, args ...interface{}) {}
	//func (l *Logger) Debug(format string, args ...interface{}) {}
	//func (l *Logger) Trace(format string, args ...interface{}) {}
	//func (l *Logger) Info(format string, args ...interface{}) {}
}

func TestLogOutput(test *t.T) {
	const (
		expected = "5d1d02513aa1d227bb762faa6b545fc1"
	)

	l := NewLogger()

	// Delete and open the output log
	os.Remove("_output.log")
	l.AddFilter("file", FINEST, NewFileLogWriter("_output.log", false).SetFormat("[%L] '%M'"))

	// Send some log messages
	l.Log(CRITICAL, "testsrc1", l.Critical("This message is a test %d", 1).String())
	l.Logf(ERROR, "This message is a test %s", l.Error(func() string { return "2" }))
	l.Logf(WARNING, "This message is a test %s", l.Warn(3))
	l.Info("This message is a test%d", 4)
	l.Trace("This message is a test%d", 5)
	l.Debug("This message is a test%d", 6)
	l.Fine("This message is a test%d", 7)
	l.Finest("This message is a test%d", 8)
	l.Finest(9, "This message is a test")
	l.Finest(func() string { return "This message is a test0" })

	l.Close()

	contents, err := ioutil.ReadFile("_output.log")
	if err != nil {
		test.Fatalf("Could not read output log: %s", err)
	}

	sum := md5.New()
	sum.Write(contents)
	sumstr := hex.EncodeToString(sum.Sum())
	if sumstr != expected {
		test.Fatalf("Checksum does not match: %s (expecting %s)", sumstr, expected)
	}
	os.Remove("_output.log")
}

func TestLogWrapperOutput(test *t.T) {
	const (
		expected = "5d1d02513aa1d227bb762faa6b545fc1"
	)

	Close()

	// Delete and open the output log
	os.Remove("_output.log")
	AddFilter("file", FINEST, NewFileLogWriter("_output.log", false).SetFormat("[%L] '%M'"))

	// Send some log messages
	Log(CRITICAL, "testsrc1", Critical("This message is a test %d", 1).String())
	Logf(ERROR, "This message is a test %s", Error(func() string { return "2" }))
	Logf(WARNING, "This message is a test %s", Warn(3))
	Info("This message is a test%d", 4)
	Trace("This message is a test%d", 5)
	Debug("This message is a test%d", 6)
	Fine("This message is a test%d", 7)
	Finest("This message is a test%d", 8)
	Finest(9, "This message is a test")
	Finest(func() string { return "This message is a test0" })

	Close()

	contents, err := ioutil.ReadFile("_output.log")
	if err != nil {
		test.Fatalf("Could not read output log: %s", err)
	}

	sum := md5.New()
	sum.Write(contents)
	sumstr := hex.EncodeToString(sum.Sum())
	if sumstr != expected {
		test.Fatalf("Checksum does not match: %s (expecting %s)", sumstr, expected)
	}
	os.Remove("_output.log")
}

func TestCountMallocs(test *t.T) {
	const N = 1

	// Console logger
	sl := NewConsoleLogger(INFO)
	mallocs := 0 - runtime.MemStats.Mallocs
	for i := 0; i < N; i++ {
		sl.Log(WARNING, "here", "This is a WARNING message")
	}
	mallocs += runtime.MemStats.Mallocs
	fmt.Printf("mallocs per sl.Log((WARNING, \"here\", \"This is a log message\"): %d\n", mallocs/N)

	// Console logger formatted
	mallocs = 0 - runtime.MemStats.Mallocs
	for i := 0; i < N; i++ {
		sl.Logf(WARNING, "%s is a log message with level %d", "This", WARNING)
	}
	mallocs += runtime.MemStats.Mallocs
	fmt.Printf("mallocs per sl.Logf(WARNING, \"%%s is a log message with level %%d\", \"This\", WARNING): %d\n", mallocs/N)

	// Console logger (not logged)
	sl = NewConsoleLogger(INFO)
	mallocs = 0 - runtime.MemStats.Mallocs
	for i := 0; i < N; i++ {
		sl.Log(DEBUG, "here", "This is a DEBUG log message")
	}
	mallocs += runtime.MemStats.Mallocs
	fmt.Printf("mallocs per unlogged sl.Log((WARNING, \"here\", \"This is a log message\"): %d\n", mallocs/N)

	// Console logger formatted (not logged)
	mallocs = 0 - runtime.MemStats.Mallocs
	for i := 0; i < N; i++ {
		sl.Logf(DEBUG, "%s is a log message with level %d", "This", DEBUG)
	}
	mallocs += runtime.MemStats.Mallocs
	fmt.Printf("mallocs per unlogged sl.Logf(WARNING, \"%%s is a log message with level %%d\", \"This\", WARNING): %d\n", mallocs/N)
}

func TestXMLConfig(test *t.T) {
	const (
		configfile = "example.xml"
	)

	fd, err := os.Create(configfile)
	if err != nil {
		test.Fatalf("Could not open %s for writing: %s", configfile, err)
	}

	fmt.Fprintln(fd, "<logging>")
	fmt.Fprintln(fd, "  <filter enabled=\"true\">")
	fmt.Fprintln(fd, "    <tag>stdout</tag>")
	fmt.Fprintln(fd, "    <type>console</type>")
	fmt.Fprintln(fd, "    <!-- level is (:?FINEST|FINE|DEBUG|TRACE|INFO|WARNING|ERROR) -->")
	fmt.Fprintln(fd, "    <level>DEBUG</level>")
	fmt.Fprintln(fd, "  </filter>")
	fmt.Fprintln(fd, "  <filter enabled=\"true\">")
	fmt.Fprintln(fd, "    <tag>file</tag>")
	fmt.Fprintln(fd, "    <type>file</type>")
	fmt.Fprintln(fd, "    <level>FINEST</level>")
	fmt.Fprintln(fd, "    <property name=\"filename\">test.log</property>")
	fmt.Fprintln(fd, "    <!--")
	fmt.Fprintln(fd, "       %T - Time (15:04:05 MST)")
	fmt.Fprintln(fd, "       %t - Time (15:04)")
	fmt.Fprintln(fd, "       %D - Date (2006/01/02)")
	fmt.Fprintln(fd, "       %d - Date (01/02/06)")
	fmt.Fprintln(fd, "       %L - Level (FNST, FINE, DEBG, TRAC, WARN, EROR, CRIT)")
	fmt.Fprintln(fd, "       %S - Source")
	fmt.Fprintln(fd, "       %M - Message")
	fmt.Fprintln(fd, "       It ignores unknown format strings (and removes them)")
	fmt.Fprintln(fd, "       Recommended: \"[%D %T] [%L] (%S) %M\"")
	fmt.Fprintln(fd, "    -->")
	fmt.Fprintln(fd, "    <property name=\"format\">[%D %T] [%L] (%S) %M</property>")
	fmt.Fprintln(fd, "    <property name=\"rotate\">false</property> <!-- true enables log rotation, otherwise append -->")
	fmt.Fprintln(fd, "    <property name=\"maxsize\">0M</property> <!-- \\d+[KMG]? Suffixes are in terms of 2**10 -->")
	fmt.Fprintln(fd, "    <property name=\"maxlines\">0K</property> <!-- \\d+[KMG]? Suffixes are in terms of thousands -->")
	fmt.Fprintln(fd, "    <property name=\"daily\">true</property> <!-- Automatically rotates when a log message is written after midnight -->")
	fmt.Fprintln(fd, "  </filter>")
	fmt.Fprintln(fd, "  <filter enabled=\"true\">")
	fmt.Fprintln(fd, "    <tag>xmllog</tag>")
	fmt.Fprintln(fd, "    <type>xml</type>")
	fmt.Fprintln(fd, "    <level>TRACE</level>")
	fmt.Fprintln(fd, "    <property name=\"filename\">trace.xml</property>")
	fmt.Fprintln(fd, "    <property name=\"rotate\">true</property> <!-- true enables log rotation, otherwise append -->")
	fmt.Fprintln(fd, "    <property name=\"maxsize\">100M</property> <!-- \\d+[KMG]? Suffixes are in terms of 2**10 -->")
	fmt.Fprintln(fd, "    <property name=\"maxrecords\">6K</property> <!-- \\d+[KMG]? Suffixes are in terms of thousands -->")
	fmt.Fprintln(fd, "    <property name=\"daily\">false</property> <!-- Automatically rotates when a log message is written after midnight -->")
	fmt.Fprintln(fd, "  </filter>")
	fmt.Fprintln(fd, "  <filter enabled=\"false\"><!-- enabled=false means this logger won't actually be created -->")
	fmt.Fprintln(fd, "    <tag>donotopen</tag>")
	fmt.Fprintln(fd, "    <type>socket</type>")
	fmt.Fprintln(fd, "    <level>FINEST</level>")
	fmt.Fprintln(fd, "    <property name=\"endpoint\">192.168.1.255:12124</property> <!-- recommend UDP broadcast -->")
	fmt.Fprintln(fd, "    <property name=\"protocol\">udp</property> <!-- tcp or udp -->")
	fmt.Fprintln(fd, "  </filter>")
	fmt.Fprintln(fd, "</logging>")
	fd.Close()

	log := NewLogger()
	log.LoadConfiguration(configfile)

	// Make sure we got 2 loggers
	if len(log.filterLevels) != 3 || len(log.filterLogWriters) != 3 {
		test.Fatalf("XMLConfig: Expected 3 filters, found %d (%d)", len(log.filterLevels), len(log.filterLogWriters))
	}

	// Make sure they're the right type
	if _, ok := log.filterLogWriters["stdout"].(*ConsoleLogWriter); !ok {
		test.Errorf("XMLConfig: Expected stdout to be *ConsoleLogWriter, found %T", log.filterLogWriters["stdout"])
	}
	if _, ok := log.filterLogWriters["file"].(*FileLogWriter); !ok {
		test.Fatalf("XMLConfig: Expected file to be *FileLogWriter, found %T", log.filterLogWriters["file"])
	}
	if _, ok := log.filterLogWriters["xmllog"].(*XMLLogWriter); !ok {
		test.Fatalf("XMLConfig: Expected xmllog to be *XMLLogWriter, found %T", log.filterLogWriters["xmllog"])
	}

	// Make sure levels are set
	if lvl := log.filterLevels["stdout"]; lvl != DEBUG {
		test.Errorf("XMLConfig: Expected stdout to be set to level %d, found %d", DEBUG, lvl)
	}
	if lvl := log.filterLevels["file"]; lvl != FINEST {
		test.Errorf("XMLConfig: Expected file to be set to level %d, found %d", FINEST, lvl)
	}
	if lvl := log.filterLevels["xmllog"]; lvl != TRACE {
		test.Errorf("XMLConfig: Expected xmllog to be set to level %d, found %d", TRACE, lvl)
	}

	// Make sure the FLW is open and points to the right file
	if ok := log.filterLogWriters["file"].Good(); !ok {
		test.Errorf("XMLConfig: Expected file to have opened %s successfully, but wasn't", "test.log")
	}
	if fname := log.filterLogWriters["file"].(*FileLogWriter).file.Name(); fname != "test.log" {
		test.Errorf("XMLConfig: Expected file to have opened %s, found %s", "test.log", fname)
	}

	// Make sure the XLW is open and points to the right file
	if ok := log.filterLogWriters["xmllog"].Good(); !ok {
		test.Errorf("XMLConfig: Expected xmllog to have opened %s successfully, but wasn't", "trace.xml")
	}
	if fname := log.filterLogWriters["xmllog"].(*XMLLogWriter).file.Name(); fname != "trace.xml" {
		test.Errorf("XMLConfig: Expected xmllog to have opened %s, found %s", "trace.xml", fname)
	}

	log.Close()
	os.Remove("test.log")
	os.Remove("trace.xml")

	// Move XML log file
	os.Rename(configfile, "examples/"+configfile) // Keep this so that an example with the documentation is available
}

func BenchmarkConsoleLog(b *t.B) {
	sl := NewConsoleLogger(INFO)
	for i := 0; i < b.N; i++ {
		sl.Log(WARNING, "here", "This is a log message")
	}
}

func BenchmarkConsoleNotLogged(b *t.B) {
	sl := NewConsoleLogger(INFO)
	for i := 0; i < b.N; i++ {
		sl.Log(DEBUG, "here", "This is a log message")
	}
}

func BenchmarkConsoleUtilLog(b *t.B) {
	sl := NewConsoleLogger(INFO)
	for i := 0; i < b.N; i++ {
		sl.Info("%s is a log message", "This")
	}
}


func BenchmarkConsoleUtilNotLog(b *t.B) {
	sl := NewConsoleLogger(INFO)
	for i := 0; i < b.N; i++ {
		sl.Debug("%s is a log message", "This")
	}
}

func BenchmarkFileLog(b *t.B) {
	sl := NewLogger()
	b.StopTimer()
	sl.AddFilter("file", INFO, NewFileLogWriter("benchlog.log", false))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Log(WARNING, "here", "This is a log message")
	}
	b.StopTimer()
	os.Remove("benchlog.log")
}

func BenchmarkFileNotLogged(b *t.B) {
	sl := NewLogger()
	b.StopTimer()
	sl.AddFilter("file", INFO, NewFileLogWriter("benchlog.log", false))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Log(DEBUG, "here", "This is a log message")
	}
	b.StopTimer()
	os.Remove("benchlog.log")
}

func BenchmarkFileUtilLog(b *t.B) {
	sl := NewLogger()
	b.StopTimer()
	sl.AddFilter("file", INFO, NewFileLogWriter("benchlog.log", false))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Info("%s is a log message", "This")
	}
	b.StopTimer()
	os.Remove("benchlog.log")
}

func BenchmarkFileUtilNotLog(b *t.B) {
	sl := NewLogger()
	b.StopTimer()
	sl.AddFilter("file", INFO, NewFileLogWriter("benchlog.log", false))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Debug("%s is a log message", "This")
	}
	b.StopTimer()
	os.Remove("benchlog.log")
}

// Benchmark results (darwin amd64 6g)
//elog.BenchmarkConsoleLog           100000       22819 ns/op
//elog.BenchmarkConsoleNotLogged    2000000         879 ns/op
//elog.BenchmarkConsoleUtilLog        50000       34380 ns/op
//elog.BenchmarkConsoleUtilNotLog   1000000        1339 ns/op
//elog.BenchmarkFileLog              100000       26497 ns/op
//elog.BenchmarkFileNotLogged       2000000         821 ns/op
//elog.BenchmarkFileUtilLog           50000       33945 ns/op
//elog.BenchmarkFileUtilNotLog      1000000        1258 ns/op
