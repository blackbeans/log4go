# Table of Contents #


# Console Log Writer #
The console log writer is the simplest writer included in log4go.  You can create one in a few ways.  In none of the examples below does the text `"stdout"` control anything, it is simply the tag that uniquely names the log writer to the logger.  All example code is included in the repository under trunk/examples (or stable/examples).

## Default Logging ##
When you import the package, without doing anything else, you already have a console log writer set up at DEBUG level.  You can send messages through the wrapper interface.

## Manual Creation ##
You always have the option to manually create your Logger.  To manually create a console log writer, the code might look something like this:

```
// Excerpt from examples/ConsoleLogWriter_Manual.go
import l4g "log4go.googlecode.com/svn/stable"

func main() {
	log.AddFilter("stdout", l4g.DEBUG, l4g.NewConsoleLogWriter())
	log.Info("The time is now: %s", time.LocalTime().Format("15:04:05 MST 2006/01/02"))
}
```

## Utility Function and wrappers ##
The above code listing could have been simplified to the following:
```
// Simplified from examples/ConsoleLogWriter_Manual.go
import l4g "log4go.googlecode.com/svn/stable"

func main() {
	l4g.AddFilter("stdout", l4g.DEBUG, l4g.NewConsoleLogWriter()) // this is actually already done by default
	l4g.Info("The time is now: %s", time.LocalTime().Format("15:04:05 MST 2006/01/02"))
}
```

For the purposes of this example, the new logwriter is created and added to the global logger (thus avoiding the necessity to pass around a `*Logger`), even though the default global logger already has a `ConsoleLogWriter` at the `DEBUG` level.  All subsequent examples can also be simplified to not create a new logger instance and instead add the filters to the global logger with `AddFilter` instead of `(*Logger).AddFilter`.

## XML Configuration ##
If you are using the XML configuration file, the following is an example of how to create a console logger:
```
<!-- This is taken from examples/example.xml -->
<logging>
  <filter enabled="true">
    <tag>stdout</tag><!-- can be anything -->
    <type>console</type>
    <level>DEBUG</level>
  </filter>
</logging>
```

# File Log Writer #
The file writer is a much more complicated writer.  All it really needs to know is the name of the file to which it is logging and whether you want it to rotate logfiles when it finds one that exists, but it can do far more than that.

## Manual Creation ##
The following are two ways to create a file logger:
```
    // Excerpt taken from examples/FileLogWriter_Manual.go
    // Create the empty logger
    log := l4g.NewLogger()

    // Create a default logger that is logging messages of FINE or higher to filename, no rotation
    log.AddFilter("file", l4g.FINE, l4g.NewFileLogWriter(filename, false))

    // =OR= Can also specify manually via the following: (these are the defaults, this is equivalent to above)
    flw := l4g.NewFileLogWriter(filename, false)
    flw.SetFormat("[%D %T] [%L] (%S) %M")
    flw.SetRotate(false)
    flw.SetRotateSize(0)
    flw.SetRotateLines(0)
    flw.SetRotateDaily(false)
    log.AddFilter("file", l4g.FINE, flw)
```
If you are not rotating logs, the file will be opened in append mode so that you don't lose any log messages.  If it is in rotating mode, then any existing files are moved to the first available `filename.###` before opening the named file.  This behavior can be enabled or disabled using `(*FileLogWriter).SetRotate(bool)`.

The table below summarizes the utility methods for the file logger package and their purpose:
| _Method_ | _Functionality_ |
|:---------|:----------------|
| `SetRotate(bool)` | Turns on and off log rotation.  The below still trigger, the log simply reopens. |
| `SetRotateSize(int)` | Will rotate on the next write after writing the given number of bytes to the file. |
| `SetRotateLines(int)` | Will rotate on the next write after reaching/exceeding the number of lines written to file. |
| `SetRotateDaily(bool)` | Will rotate on the next write after the local date changes. |
| `SetFormat(string)` | Will format log messages according to the given format string (see below). |

Formatting:
| _Specifier_ | _Replaced by_ |
|:------------|:--------------|
| `%T` | Time (15:04:05 MST) |
| `%t` | Time (15:04) |
| `%D` | Date (2006/01/02) |
| `%d` | Date (01/02/06) |
| `%L` | Level (FNST, FINE, DEBG, TRAC, WARN, EROR, CRIT) |
| `%S` | Source |
| `%M` | Message |

The formatter ignores unknown format strings (and removes them).  The default format string is `"[%D %T] [%L] (%S) %M"`.

## XML Configuration ##
The easiest way to create the file writer is via the XML interface.  Only specify the properties you need, the rest will be filled in with the defaults.
```
<!-- Taken from examples/example.xml -->
<logging>
  <filter enabled="true">
    <tag>file</tag><!-- can be anything -->
    <type>file</type>
    <level>FINEST</level>
    <property name="filename">test.log</property>
    <property name="format">[%D %T] [%L] (%S) %M</property>
    <property name="rotate">false</property> <!-- true enables log rotation, otherwise truncation -->
    <property name="maxsize">0M</property> <!-- \d+[KMG]? Suffixes are in terms of thousands -->
    <property name="maxlines">0K</property> <!-- \d+[KMG]? Suffixes are in terms of 2**10 -->
    <property name="daily">false</property> <!-- Automatically rotates when a log message is written after midnight -->
  </filter>
</logging>
```
Notice that the maxsize and maxlines respect suffixes of K, M, and G (with no intervening space).  These are in powers of `2**10` for the size and in multiples of 1000 for number of lines.

# Socket Log Writer #
The socket writer is pretty simple.  Provide it with a transport (`tcp` or `udp`) and a destination (single host for TCP and broadcast for UDP would be the typical usage) and let it go.

## Manual Creation ##
Same deal as before.  The code is pretty self-explanatory.
```
    // Excerpt from examples/SocketLogWriter_Manual.go
    log := l4g.NewLogger()
    log.AddFilter("network", l4g.FINEST, l4g.NewSocketLogWriter("udp", "192.168.1.255:12124"))

    // Run `nc -u -l -p 12124` or similar before you run this to see the following message
    log.Info("The time is now: %s", time.LocalTime().Format("15:04:05 MST 2006/01/02"))
    
    // This makes sure the output stream buffer is written
    log.Close()
```

## XML configuration ##
As usual, probably the easiest way.
```
<!-- Taken from examples/example.xml -->
<logging>
  <filter enabled="true"><!-- enabled=false means this logger won't actually be created -->
    <tag>network</tag> <!-- can be anything -->
    <type>socket</type>
    <level>FINEST</level>
    <property name="endpoint">192.168.1.255:12124</property> <!-- recommend UDP broadcast -->
    <property name="protocol">udp</property> <!-- tcp or udp -->
  </filter>
</logging>
```

## Testing UDP ##
To test your UDP broadcast socket log writer there is a simple program in the examples directory:
```
6g SimpleNetLogServer.go && 6l -o SNLS SimpleNetLogServer.6 && ./SNLS -p <port>
```

# The Easy Way (and multiple loggers) #
The easiest way to handle logging in your programs is to combine the above with XML and the [Wrapper](Wrapper.md) functions.  An example of the code and XML is below.  Even if you do not choose to use the XML configuration, you can simplify all of the above examples by not creating a `*Logger` instance and instead using the `AddFilter` instead of `(*Logger).AddFilter`, and then (in any source file in your application file) you can log to the same global logger using the global logging wrapper as below.

```
	// Adapted from XMLConfigurationExample.go
	// Load the configuration (isn't this easy?)
	l4g.LoadConfiguration("example.xml")

	/* Static logging like this is also acceptable:
	l4g.AddFilter("file", l4g.ERROR, l4g.NewFileLogWriter("errors.log", false))
	l4g.AddFilter("net", l4g.DEBUG, l4g.NewSocketLogWriter("udp", "127.0.0.1:12120"))
	*/

	// And now we're ready!
	l4g.Finest("This will only go to those of you really cool UDP kids!  If you change enabled=true.")
	l4g.Debug("Oh no!  %d + %d = %d!", 2, 2, 2+2)
	l4g.Info("About that time, eh chaps?")
```
```
<!-- example.xml -->
<logging>
  <filter enabled="true">
    <tag>stdout</tag>
    <type>console</type>
    <!-- level is (:?FINEST|FINE|DEBUG|TRACE|INFO|WARNING|ERROR) -->
    <level>DEBUG</level>
  </filter>
  <filter enabled="true">
    <tag>file</tag>
    <type>file</type>
    <level>FINEST</level>
    <property name="filename">test.log</property>
    <!--
       %T - Time (15:04:05 MST)
       %t - Time (15:04)
       %D - Date (2006/01/02)
       %d - Date (01/02/06)
       %L - Level (FNST, FINE, DEBG, TRAC, WARN, EROR, CRIT)
       %S - Source
       %M - Message
       It ignores unknown format strings (and removes them)
       Recommended: "[%D %T] [%L] (%S) %M"
    -->
    <property name="format">[%D %T] [%L] (%S) %M</property>
    <property name="rotate">false</property> <!-- true enables log rotation, otherwise truncation -->
    <property name="maxsize">0M</property> <!-- \d+[KMG]? Suffixes are in terms of thousands -->
    <property name="maxlines">0K</property> <!-- \d+[KMG]? Suffixes are in terms of 2**10 -->
    <property name="daily">false</property> <!-- Automatically rotates when a log message is written after midnight -->
  </filter>
  <filter enabled="false"><!-- enabled=false means this logger won't actually be created -->
    <tag>donotopen</tag>
    <type>socket</type>
    <level>FINEST</level>
    <property name="endpoint">192.168.1.255:12124</property> <!-- recommend UDP broadcast -->
    <property name="protocol">udp</property> <!-- tcp or udp -->
  </filter>
</logging>
```