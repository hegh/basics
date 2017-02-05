package ln

import (
	"bytes"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/hegh/basics/errors"
)

// If the file or package gets renamed, update these constants.
const (
	// The path used to import the package.
	longPackageName = "github.com/hegh/basics/ln"

	// The default alias used to reference the package after importing it.
	shortPackageName = "ln"

	// The name of this file.
	fileName = "ln_test.go"
)

// Matches a log line and extracts subgroups (see the constants below).
//
// Matches the first line out of multiple log lines.
//
// Example message:
//  X1202 10:04:59.846813 TestCall(ln_test.go:65) test info message\n
var matcher = regexp.MustCompile(
	`^(.)\d{4} (\d{2}:\d{2}:\d{2}\.\d{6}) ([^(]+)\(([^:]+):(\d+)\) (.*)`)

// The subgroup indices from the regexp above.
const (
	prefixIdx     = 1 // I, W, E, ...
	timestampIdx  = 2 // Timestamp without date
	funcNameIdx   = 3
	fileNameIdx   = 4
	lineNumberIdx = 5
	logMessageIdx = 6
)

type sink struct {
	data     *bytes.Buffer
	triggers int
}

func newSink() *sink {
	return &sink{
		data: bytes.NewBuffer(nil),
	}
}

func (s *sink) trigger()                          { s.triggers++ }
func (s *sink) Write(p []byte) (n int, err error) { return s.data.Write(p) }
func (s *sink) String() string                    { return s.data.String() }

type sync struct {
	*sink
	syncs   int
	syncErr error // Returned from Sync().
}

func (s *sync) Sync() error {
	s.syncs++
	return s.syncErr
}

// TestCaller verifies the caller function returns reasonable output.
func TestCaller(t *testing.T) {
	file, line, fnc, ok := caller(0)
	if !ok {
		t.Fatal("failed to gather callsite info")
	}

	if file != fileName {
		t.Errorf("got %q want %q for file of caller", file, fileName)
	}

	// Give the line number a little leeway for future edits.
	if line < 50 || line > 100 {
		t.Errorf("got %d want something around 30-70 for line of caller", line)
	}

	if fnc != "TestCaller" {
		t.Errorf("got %q want %q for function of caller", fnc, "TestCaller")
	}
}

// TestPackageName verifies the packageName function returns reasonable output.
func TestPackageName(t *testing.T) {
	long, short, ok := packageName(0)
	if !ok {
		t.Fatal("failed to get package name")
	}

	// Sometimes the long package name includes the full path, and sometimes it is
	// just the path used in an import line. Luckily, they both end with the long
	// package name constant.
	wantLong := longPackageName
	if !strings.HasSuffix(long, wantLong) {
		t.Errorf("got %q want something ending in %q for long package name", long, wantLong)
	}

	wantShort := shortPackageName
	if short != wantShort {
		t.Errorf("got %q want %q for short package name", short, wantShort)
	}
}

// TestCall verifies the Logger can be called like a function.
func TestCall(t *testing.T) {
	s := newSink()
	l := MakeLogger("X", s, s.trigger)

	msg := "test info message"

	// These next two lines must be adjacent.
	file, line, fnc, ok := caller(0)
	l(msg)

	if !ok {
		t.Fatal("Failed to gather callsite info with runtime.Caller(0)")
	}

	m := matcher.FindStringSubmatch(s.String())
	if m == nil {
		t.Fatalf("got %q which does not match expected line format", s.String())
	}
	if m[prefixIdx] != "X" {
		t.Errorf("got %q want %q for prefix", m[prefixIdx], "X")
	}
	if m[funcNameIdx] != fnc {
		t.Errorf("got %q want %q for function", m[funcNameIdx], fnc)
	}
	if m[fileNameIdx] != file {
		t.Errorf("got %q want %q for file", m[fileNameIdx], file)
	}
	if lineStr := strconv.Itoa(line + 1); m[lineNumberIdx] != lineStr {
		t.Errorf("got %q want %q for line", m[lineNumberIdx], lineStr)
	}
	if m[logMessageIdx] != msg {
		t.Errorf("got %q want %q for message", m[logMessageIdx], msg)
	}
	if s.triggers != 1 {
		t.Errorf("got %d want %d for trigger count", s.triggers, 1)
	}
}

// TestTZ verifies that changing the TZ changes the timestamps.
func TestTZ(t *testing.T) {
	s := newSink()
	l := MakeLogger("X", s, nil)

	TZ = time.FixedZone("zone1", 3600)
	l("msg")
	line1 := s.String()
	m1 := matcher.FindStringSubmatch(line1)
	if m1 == nil {
		t.Errorf("got %q which does not match expected line format", line1)
	}

	s.data = new(bytes.Buffer)
	TZ = time.FixedZone("zone2", -3600)
	l("msg")
	line2 := s.String()
	m2 := matcher.FindStringSubmatch(line2)
	if m2 == nil {
		t.Errorf("got %q which does not match expected line format", line2)
	}

	if t.Failed() {
		t.FailNow()
	}

	ts1 := m1[timestampIdx]
	ts2 := m2[timestampIdx]
	if ts1 == ts2 {
		t.Errorf("got equal timestamps %q for different timezones", ts1)
	}
}

// TestPrint verifies the Print method of the Logger produces the right output.
func TestPrint(t *testing.T) {
	s := newSink()
	l := MakeLogger("X", s, s.trigger)

	msg := "test info message"

	// These next two lines must be adjacent.
	file, line, fnc, ok := caller(0)
	l.Print(msg)

	if !ok {
		t.Fatal("Failed to gather callsite info with runtime.Caller(0)")
	}

	m := matcher.FindStringSubmatch(s.String())
	if m == nil {
		t.Fatalf("got %q which does not match expected line format", s.String())
	}
	if m[prefixIdx] != "X" {
		t.Errorf("got %q want %q for prefix", m[prefixIdx], "X")
	}
	if m[funcNameIdx] != fnc {
		t.Errorf("got %q want %q for function", m[funcNameIdx], fnc)
	}
	if m[fileNameIdx] != file {
		t.Errorf("got %q want %q for file", m[fileNameIdx], file)
	}
	if lineStr := strconv.Itoa(line + 1); m[lineNumberIdx] != lineStr {
		t.Errorf("got %q want %q for line", m[lineNumberIdx], lineStr)
	}
	if m[logMessageIdx] != msg {
		t.Errorf("got %q want %q for message", m[logMessageIdx], msg)
	}
	if s.triggers != 1 {
		t.Errorf("got %d want %d for trigger count", s.triggers, 1)
	}
}

// TestPrint verifies the Printf method of the Logger produces the right output.
func TestPrintf(t *testing.T) {
	s := newSink()
	l := MakeLogger("X", s, s.trigger)

	msg := "test info message"

	// These next two lines must be adjacent.
	file, line, fnc, ok := caller(0)
	l.Printf("%s", msg)

	if !ok {
		t.Fatal("Failed to gather callsite info with runtime.Caller(0)")
	}

	m := matcher.FindStringSubmatch(s.String())
	if m == nil {
		t.Fatalf("got %q which does not match expected line format", s.String())
	}
	if m[prefixIdx] != "X" {
		t.Errorf("got %q want %q for prefix", m[prefixIdx], "X")
	}
	if m[funcNameIdx] != fnc {
		t.Errorf("got %q want %q for function", m[funcNameIdx], fnc)
	}
	if m[fileNameIdx] != file {
		t.Errorf("got %q want %q for file", m[fileNameIdx], file)
	}
	if lineStr := strconv.Itoa(line + 1); m[lineNumberIdx] != lineStr {
		t.Errorf("got %q want %q for line", m[lineNumberIdx], lineStr)
	}
	if m[logMessageIdx] != msg {
		t.Errorf("got %q want %q for message", m[logMessageIdx], msg)
	}
	if s.triggers != 1 {
		t.Errorf("got %d want %d for trigger count", s.triggers, 1)
	}
}

// TestPrint verifies the Write method of the Logger writes directly to the
// Logger's Writer.
func TestWrite(t *testing.T) {
	s := newSink()
	l := MakeLogger("X", s, s.trigger)

	want := "test info message"
	l.Write([]byte(want))

	got := s.String()
	if got != want {
		t.Errorf("got %q want %q for output of Write", got, want)
	}
	if s.triggers != 1 {
		t.Errorf("got %d want %d for trigger count", s.triggers, 1)
	}
}

// TestLogTo verifies we can redirect the output of a logger to multiple
// Writers.
func TestLogTo(t *testing.T) {
	// Set up two loggers, each with its own sink.
	s1 := newSink()
	l1 := MakeLogger("A", s1, s1.trigger)

	// The second logger writes to both its own sink and the first logger.
	s2 := newSink()
	l2 := MakeLogger("B", s2, s2.trigger)
	l2.LogTo(l1, s2)

	msg := "msg"
	l2(msg)

	// Verify the message got written to s1, with l2's prefix, and that both
	// triggers got called.
	m := matcher.FindStringSubmatch(s1.String())
	if m == nil {
		t.Fatalf("got %q which does not match expected line format", s1.String())
	}
	if m[prefixIdx] != "B" {
		t.Errorf("got %q want %q for prefix", m[prefixIdx], "B")
	}
	if m[logMessageIdx] != msg {
		t.Errorf("got %q want %q for message", m[logMessageIdx], msg)
	}
	if s1.triggers != 1 {
		t.Errorf("got %d want %d for s1 trigger count", s1.triggers, 1)
	}
	if s2.triggers != 1 {
		t.Errorf("got %d want %d for s2 trigger count", s2.triggers, 1)
	}
}

// TestVerbosity verifies the Verbosity var controls the logger returned by V.
func TestVerbosity(t *testing.T) {
	Info = MakeLogger("I", os.Stderr, nil)

	Verbosity = 0
	if l := V(1); l.String() != NilLogger().String() {
		t.Errorf("got %q want %q for V(1) with Verbosity = %d", l, NilLogger(), Verbosity)
	}

	Verbosity = 1
	if l := V(1); l.String() != Info.String() {
		t.Errorf("got %q want %q for V(1) with Verbosity = %d", l, Info, Verbosity)
	}

	Verbosity = 2
	if l := V(1); l.String() != Info.String() {
		t.Errorf("got %q want %q for V(1) with Verbosity = %d", l, Info, Verbosity)
	}
}

// TestPackageVerbosity verifies that PackageVerbosity overrides Verbosity.
func TestPackageVerbosity(t *testing.T) {
	Info = MakeLogger("I", os.Stderr, nil)

	// Verify we can lower the verbosity.
	Verbosity = 1
	PackageVerbosity[shortPackageName] = 0
	if l := V(1); l.String() != NilLogger().String() {
		t.Errorf("got %q want %q for V(1) with PackageVerbosity = %d", l, NilLogger(), PackageVerbosity[shortPackageName])
	}

	// Verify we can raise the verbosity.
	Verbosity = 0
	PackageVerbosity[shortPackageName] = 1
	if l := V(1); l.String() != Info.String() {
		t.Errorf("got %q want %q for V(1) with PackageVerbosity = %d", l, Info, PackageVerbosity[shortPackageName])
	}
}

// SetTrigger verifies we can change the trigger on a logger.
func TestSetTrigger(t *testing.T) {
	s := newSink()
	l := MakeLogger("X", s, nil)

	l("1")
	l.SetTrigger(s.trigger)
	l("2")
	l.SetTrigger(nil)
	l("3")

	if s.triggers != 1 {
		t.Errorf("got %d want %d for trigger count", s.triggers, 1)
	}
}

// TestAbortMe verifies the AbortMe function sends a SIGABRT to the current
// process.
func TestAbortMe(t *testing.T) {
	defer func() { signal.Reset(syscall.SIGABRT) }()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGABRT)

	AbortMe()
	select {
	case <-sigs:
	case <-time.After(5 * time.Second):
		t.Errorf("Timed out waiting for SIGABRT")
	}
}

// TestSyncWriter verifies that Sync is called on those writers that have it.
func TestSyncWriter(t *testing.T) {
	s1 := newSink()
	s2 := &sync{
		sink: newSink(),
	}
	l := MakeLogger("X", s1, nil)
	l.LogTo(s1, s2)

	_, err := l("test")
	if err != nil {
		t.Errorf("unexpected error from l()", err)
	}
	if s2.syncs != 1 {
		t.Errorf("got %d want %d for s2.syncs", s2.syncs, 1)
	}

	s2.syncErr = errors.New("an error")
	_, err = l("test")
	if s2.syncs != 2 {
		t.Errorf("got %d want %d for s2.syncs", s2.syncs, 2)
	}
	if err != s2.syncErr {
		t.Errorf("got %q want %q for error from l()", err, s2.syncErr)
	}
}

func TestReplaceErrors(t *testing.T) {
	err := errors.Errorf("message")
	s := []interface{}{"hello", err, 1}
	replaceErrors(s)
	if s[0].(string) != "hello" {
		t.Errorf("got %q want %q for s[0] after replaceErrors", s[0], "hello")
	}
	if s[2].(int) != 1 {
		t.Errorf("got %q want %q for s[2] after replaceErrors", s[2], 1)
	}

	re := regexp.MustCompile(`(?ms:^message$.*ln\.TestReplaceErrors\(\)$.*ln_test.go.*$.*)`)
	if !re.MatchString(s[1].(string)) {
		t.Errorf("expected %q to match\n%v", re, s[1])
	}
}
