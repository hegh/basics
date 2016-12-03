package ln

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	// Verbosity provides control over whether the Logger returned by V will do
	// anything.
	Verbosity = 0

	// PackageVerbosity provides overrides to the Verbosity based on the package
	// name.
	//
	// Mapping of package name onto verbosity level. Short package name is fine,
	// but if you need to disambiguate you can use the full package path.
	// Note that package main is always just 'main' and should never have a path.
	PackageVerbosity = make(map[string]int)

	// Info logs messages at Info level.
	Info = MakeLogger("I", os.Stderr, nil)

	// Warning logs messages at Warning level.
	Warning = MakeLogger("W", os.Stderr, nil)

	// Error logs messages at Error level.
	Error = MakeLogger("E", os.Stderr, nil)

	// Fatal logs messages at Fatal level, and then terminates the program.
	Fatal = MakeLogger("F", os.Stderr, Terminate)

	nilLogger = Logger(func(a ...interface{}) (int, error) {
		return 0, nil
	})
)

// V returns the Info logger if the given level is less than or equal to the
// current Verbosity. Otherwise it returns the nil logger, which throws away
// everything logged to it.
func V(level int) Logger {
	v := Verbosity
	if pv, ok := packageVerbosity(1); ok {
		v = pv
	}

	if level <= v {
		return Info
	}
	return nilLogger
}

// Logger is the main interface to this package. It annotates messages and
// writes them to an io.Writer.
//
// Calling the Logger as a function acts the same as calling the Print method
// with the same parameters.
//
// A nil Logger is perfectly valid, and will act like /dev/null for all of its
// output. The only thing you can't do with it is call it like a function.
type Logger func(a ...interface{}) (n int, err error)

// Print writes the parameters to the Logger, formatted as if they were passed
// through fmt.Print.
func (l Logger) Print(a ...interface{}) (int, error) {
	lg := l.getLogger()
	if lg == nil {
		return 0, nil
	}

	message := assemble(1, lg.prefix, fmt.Sprint(a...))
	return lg.Write(message)
}

// Printf writes a formatted result to the Logger, using the same formatting
// rules as fmt.Printf.
func (l Logger) Printf(format string, a ...interface{}) (int, error) {
	lg := l.getLogger()
	if lg == nil {
		return 0, nil
	}

	message := assemble(1, lg.prefix, fmt.Sprintf(format, a...))
	return lg.Write(message)
}

// Write is a low-level function that forwards its parameter directly to the
// io.Writer associated with the Logger.
//
// If the Logger has an associated trigger function, it is called after writing
// the value.
func (l Logger) Write(p []byte) (int, error) {
	lg := l.getLogger()
	if lg == nil {
		return 0, nil
	}
	return lg.Write(p)
}

// LogTo changes the io.Writer associated with the Logger.
//
// The Logger will write to all of the associated writers, which can be other
// Loggers.
func (l Logger) LogTo(w ...io.Writer) {
	lg := l.getLogger()
	if lg == nil {
		return
	}
	lg.w = io.MultiWriter(w...)
}

// SetTrigger changes the trigger that gets called when anything is written to
// the Logger.
func (l Logger) SetTrigger(trigger func()) {
	lg := l.getLogger()
	if lg == nil {
		return
	}
	lg.trigger = trigger
}

// String returns the prefix of the Logger, or "(nil)".
func (l Logger) String() string {
	return l.getLogger().String()
}

// Get the logger holding the data associated with the given Logger.
//
// Watch out for nil, which will happen for the nil logger.
func (l Logger) getLogger() *logger {
	if l == nil {
		return nil
	}

	var lg *logger
	l(op{
		op: func(lgx *logger) (int, error) {
			lg = lgx
			return 0, nil
		},
	})
	return lg
}

// NilLogger returns a Logger that does nothing, as cheaply as possible.
func NilLogger() Logger {
	return nilLogger
}

// PrintWriter provides an io.Writer interface backed by a Print function like
// the logging functions on testing.T (they do not return anything).
//
// The byte slice passed to the Write function is converted to a string and
// forwarded to the given Print function.
//
// To use a Logger backed by a testing.T, set it up like this:
//   Info.LogTo(PrintWriter{t.Log})
type PrintWriter struct {
	P func(...interface{})
}

// Write converts the given byte slice to a string and prints it to the Print
// function backing the PrintWriter.
func (w PrintWriter) Write(p []byte) (int, error) {
	w.P(string(p))
	return len(p), nil
}

// Holds the data associated with a Logger.
type logger struct {
	prefix  string
	w       io.Writer
	trigger func()
}

// Write writes the given message to the io.Writer associated with the logger.
//
// If the logger has a trigger function, calls it after writing the message.
func (l *logger) Write(p []byte) (n int, err error) {
	n, err = l.w.Write(p)
	if t := l.trigger; t != nil {
		t()
	}
	return
}

// String returns the logger's prefix, or "(nil)".
//
// Primarily intended for debugging.
func (l *logger) String() string {
	if l == nil {
		return "(nil)"
	}
	return l.prefix
}

// This is a bit of magic that lets Logger be a function instead of a struct.
//
// If the one and only parameter to the Logger function is an op, then it calls
// op.op with the logger captured by the Logger closure, allowing methods on
// Logger to do what they do.
type op struct {
	op func(lg *logger) (n int, err error)
}

// MakeLogger builds a new Logger.
//
// prefix: The logger name prefix, which is prefixed to every line
//   formatted and written by the logger.
// w: The io.Writer that the logger sends all of its messages to.
// trigger: If not nil, this function is called after each message goes through
//   the logger. Primarily intended for the Fatal logger implementation, which
//   triggers the program to terminate after logging a message.
func MakeLogger(prefix string, w io.Writer, trigger func()) Logger {
	lg := &logger{
		prefix:  prefix,
		w:       w,
		trigger: trigger,
	}
	var l Logger = func(a ...interface{}) (int, error) {
		if len(a) == 1 {
			if o, ok := a[0].(op); ok {
				// Implement the magic that lets Logger be a function rather than a
				// struct, but still have associated data.
				return o.op(lg)
			}
		}
		message := assemble(1, lg.prefix, fmt.Sprint(a...))
		return lg.Write(message)
	}
	return l
}

// Formats a log message.
//
// skip: How many stack frames to go back (0 = caller of assemble) when
//   gathering callsite information to include in the message.
// prefix: The logger name prefix. Prefixed to the formatted log message.
// msg: The message to log.
//
// Returns the formatted message, including a newline, as a byte slice that can
// be passed directly to an io.Writer.
func assemble(skip int, prefix string, msg string) []byte {
	now := time.Now()
	file, lineNum, fnc, ok := caller(skip + 1)
	var line string
	if ok {
		line = strconv.Itoa(lineNum)
	} else {
		fnc = "????"
		file = "???"
		line = "??"
	}

	return []byte(fmt.Sprintf("%s%s %s %s(%s:%s) %s\n",
		prefix, now.Format("0102"),
		now.Format("15:04:05.000000"),
		fnc, file, line,
		msg))
}

// Returns the file name (without path), line, and function name (witchout path)
// of the caller.
//
// skip: How many stack frames to go back (0 = caller of caller).
func caller(skip int) (file string, line int, fnc string, ok bool) {
	file, line, fnc, ok = fullCaller(skip + 1)
	if !ok {
		return
	}

	file = path.Base(file)
	dot := strings.LastIndex(fnc, ".")
	if dot != -1 {
		fnc = fnc[dot+1:]
	}
	return
}

// Returns the full file name, line, and full function name of the caller.
//
// skip: How many stack frames to go back (0 = caller of fullCaller).
func fullCaller(skip int) (file string, line int, fnc string, ok bool) {
	var pc uintptr
	pc, file, line, ok = runtime.Caller(skip + 1)
	if !ok {
		return
	}

	f := runtime.FuncForPC(pc)
	fnc = f.Name()
	return
}

// PackageName returns both the long and short names of the package of the
// caller.
//
// skip = 0 is the caller of packageName.
func packageName(skip int) (long, short string, ok bool) {
	// Full function name looks like path/to/pkg.Func
	var f string
	_, _, f, ok = fullCaller(skip + 1)
	if !ok {
		return
	}

	// Strip down to path/to/pkg
	dot := strings.LastIndex(f, ".")
	if dot == -1 {
		return
	}
	long = f[:dot]

	// Strip down to pkg
	slash := strings.LastIndex(long, "/")
	if slash == -1 {
		return
	}
	short = long[slash+1:]
	return
}

// PackageVerbosity returns the package-specific verbosity, or false if not set.
func packageVerbosity(skip int) (v int, ok bool) {
	if len(PackageVerbosity) == 0 {
		return
	}

	long, short, ok := packageName(skip + 1)
	v, ok = PackageVerbosity[long]
	if ok {
		return
	}

	v, ok = PackageVerbosity[short]
	return
}
