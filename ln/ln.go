package ln

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type any = interface{}

var (
	// TZ is the timezone to use for log messages.
	//
	// If it is nil, uses the default for time.Now().
	TZ *time.Location

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

	// Debug logs messages at Debug level.
	Debug = New("D", os.Stderr, nil)

	// Info logs messages at Info level.
	Info = New("I", os.Stderr, nil)

	// Warning logs messages at Warning level.
	Warning = New("W", os.Stderr, nil)

	// Error logs messages at Error level. Syncs after every write.
	Error = New("E", NewSyncWriter(os.Stderr), nil)

	// Fatal logs messages at Fatal level, and then terminates the program.
	Fatal = New("F", NewSyncWriter(os.Stderr), Terminate)

	nilLogger = Logger(func(a ...any) (int, error) {
		return 0, nil
	})
)

// LogAllTo sets up all loggers using the default prefixes & triggers, writing
// to the given writer.
//
// Does not set any writers to sync.
func LogAllTo(w io.Writer) {
	Debug = New("D", w, nil)
	Info = New("I", w, nil)
	Warning = New("W", w, nil)
	Error = New("E", w, nil)
	Fatal = New("F", w, Terminate)
}

// MakeLogger is deprecated in favor of `New`, and may be removed in the future.
func MakeLogger(prefix string, w io.Writer, trigger func()) Logger {
	return New(prefix, w, trigger)
}

// New returns a new Logger that writes to `w`.
//
// Every line of output will have the given `prefix`. Usually this is a single
// letter, but in can be anything.
//
// If not nil, the `trigger` function is called after each message is written.
// Primarily intended for the Fatal logger implementation, which triggers the
// program to terminate after logging a message.
//
// To sync after each write, wrap the writer in a `NewSyncWriter(w)` call.
//
// To write to multiple sinks, wrap with an `io.MultiWriter`.
func New(prefix string, w io.Writer, trigger func()) Logger {
	lg := &logger{
		prefix:  prefix,
		w:       w,
		trigger: trigger,
	}
	return newLogger(lg)
}

func newLogger(lg *logger) Logger {
	var l Logger = func(a ...any) (int, error) {
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

// Config holds the configuration settings for a collection of loggers, and
// provides simple snapshot/restore for the package settings.
type Config struct {
	TZ                                 *time.Location
	Verbosity                          int
	PackageVerbosity                   map[string]int
	Debug, Info, Warning, Error, Fatal Logger
}

// Restore sets the package settings to the values from the config.
//
// The PackageVerbosity map is cloned, so changes to the config are not
// reflected in the package post-restore, and vice-versa.
//
// May be called more than once.
func (c *Config) Restore() {
	TZ = c.TZ
	Verbosity = c.Verbosity
	PackageVerbosity = make(map[string]int, len(c.PackageVerbosity))
	for k, v := range c.PackageVerbosity {
		PackageVerbosity[k] = v
	}
	Debug, Info, Warning, Error, Fatal = c.Debug, c.Info, c.Warning, c.Error, c.Fatal
}

// Snapshot takes a snapshot of the current package settings, to allow for
// easy restoration.
//
// The PackageVerbosity map is cloned, so changes post-snapshot are not
// reflected in the snapshot.
func Snapshot() *Config {
	pv := make(map[string]int, len(PackageVerbosity))
	for k, v := range PackageVerbosity {
		pv[k] = v
	}
	return &Config{
		TZ:               TZ,
		Verbosity:        Verbosity,
		PackageVerbosity: pv,
		Debug:            Debug.Clone(),
		Info:             Info.Clone(),
		Warning:          Warning.Clone(),
		Error:            Error.Clone(),
		Fatal:            Fatal.Clone(),
	}
}

// LevelEnabled returns true if a log message at the given level would be
// passed through from the current file and with the current verbosity settings.
func LevelEnabled(level int) bool {
	v := Verbosity
	if pv, ok := packageVerbosity(1); ok {
		v = pv
	}
	return level <= v
}

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
// A nil Logger is valid, and will throw away its output, BUT it will panic if
// called like a function.
type Logger func(a ...any) (n int, err error)

// String returns the prefix of the Logger, or "?".
func (l Logger) String() string { return l.getLogger().String() }

// Clone returns a new Logger that is a copy of the receiver.
func (l Logger) Clone() Logger {
	lg := l.getLogger()
	if lg == nil {
		return NilLogger()
	}

	lg = lg.clone()
	return newLogger(lg)
}

// Print writes the parameters to the Logger, formatted as if they were passed
// through fmt.Print.
func (l Logger) Print(a ...any) (int, error) {
	lg := l.getLogger()
	if lg == nil {
		return 0, nil
	}

	message := assemble(1, lg.prefix, fmt.Sprint(a...))
	return lg.Write(message)
}

// Printf writes a formatted result to the Logger, using the same formatting
// rules as fmt.Printf.
func (l Logger) Printf(format string, a ...any) (int, error) {
	lg := l.getLogger()
	if lg == nil {
		return 0, nil
	}

	message := assemble(1, lg.prefix, fmt.Sprintf(format, a...))
	return lg.Write(message)
}

// LogTo changes the io.Writer associated with the Logger.
//
// The Logger will write to all of the associated writers, which can be other
// Loggers. If the list is empty, then the logger will not output anything.
//
// Has no effect on the nil logger.
//
// If you want to sync after writing a message, wrap your logger in
// `NewSyncLogger(w)`.
func (l Logger) LogTo(writers ...io.Writer) {
	lg := l.getLogger()
	if lg == nil {
		return
	}
	lg.w = io.MultiWriter(writers...)
}

// Write is a low-level function that forwards its parameter directly to the
// io.Writer associated with the Logger.
//
// If the io.Writer has a `Sync() error` function (like os.File) then that is
// called after writing.
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

// SetTrigger changes the trigger that gets called when anything is written to
// the Logger.
//
// No-op on the nil logger, which does not support triggers.
func (l Logger) SetTrigger(trigger func()) {
	lg := l.getLogger()
	if lg == nil {
		return
	}
	lg.trigger = trigger
}

// getLogger returns the logger holding the data associated with the given
// Logger.
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
//
// This logger CAN be called like a function.
func NilLogger() Logger {
	return nilLogger
}

// PrintWriter provides an io.Writer interface backed by a Print function like
// the logging functions on testing.T (they do not return anything).
//
// The byte slice passed to the Write function is converted to a string and
// forwarded to the given Print function, removing up to one trailing newline.
//
// To use a Logger backed by a testing.T, set it up like this:
//
//	Info.LogTo(PrintWriter{t.Log})
type PrintWriter struct {
	P func(...any)
}

// Write converts the given byte slice to a string and prints it to the Print
// function backing the PrintWriter.
func (w PrintWriter) Write(p []byte) (int, error) {
	w.P(string(bytes.TrimSuffix(p, []byte{'\n'})))
	return len(p), nil
}

// Holds the data associated with a Logger.
type logger struct {
	prefix  string
	w       io.Writer // Probably an io.MultiWriter. May be nil.
	trigger func()    // May be nil.
}

func (l *logger) clone() *logger {
	return &logger{
		prefix:  l.prefix,
		w:       l.w,
		trigger: l.trigger,
	}
}

// SyncableWriter is a writer than can Sync its output.
type SyncableWriter interface {
	io.Writer
	Sync() error
}

// SyncWriter is an io.Writer that syncs after each write.
type SyncWriter struct{ w SyncableWriter }

// NewSyncWriter wraps the given SyncableWriter in a SyncWriter.
//
// Writes through the returned SyncWriter will call Sync after writing.
func NewSyncWriter(w SyncableWriter) *SyncWriter { return &SyncWriter{w} }

// Write writes the given data following the contract specified by
// io.Writer.Write.
//
// On a successful write (no error), syncs the writer before returning. If the
// sync fails, that error is returned but the number of bytes written will be
// equal to `len(p)`.
func (w *SyncWriter) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	if err != nil {
		return
	}

	err = w.w.Sync()
	return
}

// Write writes the given message to the writer associated with the logger.
//
// If the logger has a trigger function, calls it after writing the message.
func (l *logger) Write(p []byte) (n int, err error) {
	defer func() {
		if t := l.trigger; t != nil {
			t()
		}
	}()

	n, err = l.w.Write(p)
	return
}

// String returns the logger's prefix, or "?".
//
// Primarily intended for debugging.
func (l *logger) String() string {
	if l == nil {
		return "?"
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

// assemble concatenates the parts to create a full log message.
//
// `skip` specifies how many stack frames to go back (0 = caller of assemble)
// when gathering callsite information to include in the message.
//
// Returns the formatted message, including a newline, as a byte slice.
func assemble(skip int, prefix string, msg string) []byte {
	now := time.Now()
	if tz := TZ; tz != nil {
		now = now.In(tz)
	}

	file, lineNum, fnc, ok := caller(skip + 1)
	var line string
	if ok {
		line = strconv.Itoa(lineNum)
	} else {
		fnc = "????"
		file = "???"
		line = "??"
	}

	return []byte(fmt.Sprintf("%s%s %s(%s:%s) %s\n",
		prefix, now.Format("0102 15:04:05.000000"),
		fnc, file, line,
		msg))
}

// caller returns the file name (without path), line, and function name
// (witchout path) of the caller.
//
// Jumps back `skip` frames (0 = caller of `caller`).
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

// fullCaller returns the full file name, line, and full function name of the
// caller.
//
// Jumps back `skip` frames (0 = caller of `fullCaller`).
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

// packageName returns both the long and short names of the package of the
// caller.
//
// skip = 0 is the caller of `packageName`.
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

// packageVerbosity returns the package-specific verbosity, or false if not set.
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

// ParsePackageVerbosity parses the given string of comma-separated
// `package=verbosity` strings and merges them into `PackageVerbosity`.
//
// Returns an error on encountering a parse error.
func ParsePackageVerbosity(s string) error {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	for _, part := range parts {
		pkg, v, ok := strings.Cut(part, "=")
		if !ok {
			return fmt.Errorf("'%s' in package verbosity '%s' not in 'pkg=verbosity' format", part, s)
		}

		verb, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return fmt.Errorf("'%s in package verbosity '%s': bad verbosity: %w", part, s, err)
		}
		PackageVerbosity[pkg] = int(verb)
	}
	return nil
}
