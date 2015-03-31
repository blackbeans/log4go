# Installation #

To install log4go, run one of the following commands:
```
goinstall log4go.googlecode.com/hg       # new installations
goinstall -u log4go.googlecode.com/hg    # to update
```

# Quick Start #

First, you need to import the log4go package into your program.  Here is an example of some import statements and how you might include log4go
```
import (
	"os"
	"fmt"
	"container/vector"
	"log4go.googlecode.com/hg"
) 
```

The package is "log4go", and as such that is the prefix for everything imported as above.  I recommend using the shorthand l4g to make code slightly simpler, as in the following:
```
import l4g "log4go.googlecode.com/hg"
```

## Adding logging to your code ##
A standard console logger is defined by default.  As of v2.0.0, it is configured at the DEBUG filter level (only DEBUG or higher levels will be displayed), and writes to standard output with the default format.  Thus, anywhere in your code, you can use lines of code like the following:
```
// Formatted logging can be done at any of the logging levels (Finest, Fine, Debug, Trace, Info, Warning, Error, Critical)
l4g.Trace("Received message: %s (%d)", msg, length)
...
// Warnings, Errors, and Criticals provide an os.Error that you can use for a return
return l4g.Error("Unable to open file: %s", err)
...
// The wrapper functions can also behave like Sprint if the first argument isn't a string
l4g.Debug(portno, clientid, client)
...
// Use a closure so that if DEBUG isn't logged, it doesn't take any time
l4g.Debug(func()string{ decodeRaw(raw) })
```

When you start to want to log to a file in addition to the console:
```
l4g.AddFilter("file", l4g.NewFileLogWriter("myapp.log", false))
```

You can also make your life really easy by coping the `examples/example.xml` file and letting that configure your logging setup every time you run your application:
```
l4g.LoadConfiguration("logging.xml")
```

## Using the `log` interface ##

If you have legacy code that is using the standard `log` package, you can use the following instead to ease migration:
```
import log "log4go.googlecode.com/hg"
```

# Further Reading #

Here are some other pages with information to help you out:
  * The writers are documented in [LogWriters](LogWriters.md)