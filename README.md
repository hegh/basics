A small collection of low-level Golang packages to augment or replace bits of
the standard library.

# Components

* [cache/lru](#lru---an-lru-cache) - An LRU cache with a read-through interface
* [io/writecounter](#writecounter---a-writer-that-counts-bytes-written) - A writer that counts bytes written
* [ln](#ln---a-logging-package-with-a-natural-interface) - A logging package with a natural interface
* [refcount](#refcount---for-refcounting-expensive-resources) - For refcounting expensive resources
* [sync/semaphore](#semaphore---a-simple-semaphore) - A semaphore implementation.
* [todo](#todo---filler-for-functions-that-havent-been-written-yet) - Filler for functions that haven't been written yet

Deprecated:

* [errors](#errors---errors-with-stack-traces-and-causes) - Errors with stack traces and causes

## ln - A logging package with a natural interface

A new logging package that is a bit more flexible and easier to use than the
standard library `log` package.

An output log line will look something like this:

    I1203 10:04:59.846813 FuncName(filename.go:65) Message

The `I` indicates this is at Info level. It could also be a `D`, `W`, `E`, or
`F` for Debug, Warning, Error, or Fatal, respectively (the prefix is
configurable).

The `1203` indicates this was logged on December 3rd.

The `10:04:59.846813` is the timestamp, on a 24-hour clock in local time
(the timezone is configurable).

The `FuncName(filename.go:65)` describes the caller that logged the message.

Finally, `Message` is the message that was logged.

The important parts of the interface are:

### How to log a message

    ln.Info("Message ", "concatenated ", "from ", "pieces")
    ln.Info.Print("This acts exactly like Info()")
    ln.Info.Printf("Message %s", "through a format string")

The package also provides built-in `Debug`, `Warning`, `Error`, and `Fatal`
loggers.

### Logging errors

    ln.Info.Printf("Error: %v", errors.New("message"))

Errors passed to the formatted output functions get passed through the
`basics/errors` package's `String` method, which expands stack traces when they
are available.

Note that this means a `string` is being passed to `fmt.Sprintf` instead of an
`error`, not that it should make any difference.

### Verbosity control

    ln.Verbosity = 5
    ln.PackageVerbosity["main"] = 2
    ln.V(3).Print("Message")

Verbosity is controlled by a global `Verbosity` level, and a package-specific
`PackageVerbosity` map. Package verbosity takes precedence.

A package name can be a short name like `http`, or a long name like `net/http`.
Short names can be ambiguous, so long names take precedence.

### Output locations

    ln.LogAllTo(logFile)
    ln.Info.LogTo(ln.Debug)
    ln.Warning.LogTo(ln.Info)
    ln.Error.LogTo(ln.Warning)
    ln.Fatal.LogTo(ln.Error, os.Stderr)

Output can be sent to any number of `io.Writer` implementations. A logger is
itself a writer, so you can direct loggers at each other (just don't make a
cycle).

By default, all loggers write to `os.Stderr`.

### Send to a testing.T

I recommend defining a `func init()` in each of your test files like this:

```
func setupLogging(t testing.TB) func() {
    snap := ln.Snapshot()
    ln.LogAllTo(ln.PrintWriter{t.Log})
    ln.Fatal.SetTrigger(t.FailNow)

    ln.Info.Printf("beginning %s", t.Name())
    return snap.Restore
}
```

And in each of your test functions, begin like:

```
func TestSomething(t *testing.T) {
    defer setupLogging(t)()

    // ... rest of test ...
}
```

`PrintWriter` turns a function like `testing.T.Log` into an output location for
log messages.

If you wanted, you could `ln.Error.LogTo(t.Error)` to fail the test if an error
message gets logged, but I don't recommend it.

For death tests, you might try:

    ln.Fatal = ln.MakeLogger("F", ln.PrintWriter{t.Log}, func() {
      fatalCalled = true
      panic(typeSpecificErrorTest{})
    })

If a fatal condition won't muck up the environment for other tests.

### Use UTC for log message timestamps

    ln.TZ = time.FixedZone("UTC", 0)

By default, the package uses the local timezone.

This was a conscious choice. It keeps short, local programs shorter by not
requiring them to switch the timezone to something easy to read, and it only
adds one line to larger programs.

### Recommended setup for larger programs

Large programs tend to have strong opinions on how to configure logging, and
they tend to want to configure it all early enough for init functions to use it.

When you are in control of the whole codebase, add a `preinit` package that is
included by anything that needs logging in `func init()`. In your `preinit`
package, configure logging.

If you are not in full control of your codebase and pieces outside your control
want to log during `init` then we will need to add special flag-based
configuration to `ln` itself. But I am hoping that this won't be necessary.


## lru - An LRU cache

A simple cost-based cache with a least-recently-used eviction strategy. Can be
used as either a read-through cache (it will call a getter for each missing
entry), or as a manual cache (you put entries into the cache, and it holds
some of them).

All you need to provide is a maximum cache cost. Optionally, you can also
provide a function to be called on eviction, and a function to be called to
retrieve missing entries (to make it a read-through cache).

See cache/lru/lru.go for usage instructions.

## refcount - For refcounting expensive resources

If you want to release a resource after all users of it are finished, and you
don't want to hold multiple copies of it open at the same time, reference
counting is a good strategy. This package provides a basic implementation.

For example, let's say you want to service concurrent data requests by reading
random-access sections from a set of files. You don't want to hold the same file
open from multiple routines, and you don't want to hold all of the files open at
the same time. This package will help you hold files open only as long as they
are actively being used.

This is meant to be an internal feature of your object, not exposed to users.

To use:

 1. Include a `*refcount.RefCount` in your object.
 2. Write `Opener` and `Closer` functions that manage the lifecycle of the
    resource that you want to control. The `Opener` will be called when the
    resource was not open, but is required to fulfill a request. The `Closer`
    will be closed when all concurrent requests have been fulfilled and the
    resource is no longer needed.
 3. Initialize the `*RefCount` object using the `Opener` and `Closer` in your
    `New` function.
 4. When a new request arrives, call `Increment` on the `RefCount` object. This
    may cause the `Opener` to be called. Associate the returned `io.Closer` with
    the request, so when it is complete its `Close` method will be closed.
 5. Make sure to call the `Close` method of the returned `io.Closer` when the
    request is finished. This may cause the `Closer` function to be called.

See the file comment in refcount/refcount.go for an example.

### How to use refcount and LRU cache together

This is a bit tricky, but you can combine the use of the `lru` package with an
existing use of `refcount` in a way that is invisible to your callers. This
example assumes the refcounted object is a file, kept open as long as there are
any open references to it. A file is refreshed in the cache any time a new
reference is requested for it.

Example:

```
var (
  cacheLock sync.Mutex
  fileCache *lru.Cache // Cache of open files by path.
)

func init() {
  fileCache = lru.New(16) // Max 16 open files.

  // On cache eviction, close the reference to the file.
  // If there are other references open, the file will stay open until they are
  // closed, but it won't be held open by the cache any longer.
  fileCache.OnEvict = func(_ lru.Key, value any) {
    value.(io.Closer).Close()
  }
}

type File struct {
  path string

  rc *refcount.RefCount
  file *os.File
}

func New(path string) *File {
  f := &File{
    path: path,
  }
  f.rc = refcount.New(f.open, f.close)
  return f
}

func (f *File) open() (err error) {
  c.file, err = os.Open(c.path)
  return
}

func (f *File) close() error {
  c.file.Close() // Ignore the error. This is assuming read-only file caching.
  c.file = nil
  return nil
}

func (f *File) GetReference() (closer io.Closer, err error) {
  // Get two open references. The first will be returned.
  // If the file is already open in the cache, this will be super cheap.
  closer, err = f.rc.Increment()
  if err != nil {
    return
  }
  defer func() {
    if err != nil {
      closer.Close() // Clean up on error.
      closer = nil
    }
  }()

  // The second closer goes into the cache.
  cachedCloser, err := f.rc.Increment()
  if err != nil {
    return
  }
  func() {
    cacheLock.Lock()
    defer cacheLock.Unlock()
    had := fileCache.Put(f.path, 1, cachedCloser)
    if had != nil {
      cachedCloser = had.(io.Closer)
    }
  }()

  // If there was already a cached closer, close it (we don't need three).
  // Do this outside of the synchronized subfunction so we don't hold the lock
  // open any longer than needed.
  if cachedCloser != nil {
    cachedCloser.Close()
  }
  return
}
```

## todo - Filler for functions that haven't been written yet

If you are writing new code, and want to test it as you go, the `todo` package
provides a couple of methods to fill in functions you haven't written yet.

### Error - Returns an error

    return todo.Error()

The returned error will look like this:

    TODO: implement func FunctionName at filename.go:123

### Panic - Panics

    todo.Panic()

The `panic` value is the same error that would have been returned by `Error`.

## writecounter - A writer that counts bytes written

That's pretty much it. Wrap your `io.Writer` in a `writecounter.Writer` and it
will simply count the number of bytes that you've written.

This can be especially useful when writing to files with to avoid repeated calls
to `Seek`. It also improves testability, because the lack of `Seek` calls means
your test code doesn't need to use real files.

As a bonus, I've found that the ability to write simple integers directly to the
writer as bytes is very handy, so it includes a wrapper around `binary.Write`.
You can just use `WriteValue` for those now. Set the byte order in
`Writer.ByteOrder` (default is BigEndian).

## errors - (Deprecated) Errors with stack traces and causes

Deprecated in favor of the new Golang `errors` package with the `%w` format
directive, although that is still missing stack traces. I may revisit this to
update it to work with the newer `errors`, but I've basically stopped using it
myself.

A drop-in replacement for the Golang `errors` package, and `fmt.Errorf`. As long
as you use the fuctions provided by this package, errors will have causes
and stack traces tracked across message formatting.

The important parts of the interface are:

### New - Build a new error

    errors.New("error description")

A drop-in replacement for the Golang standard library `errors.New` function.

Returns an `error` whose `Error` function will return the given message. The
error will have an attached stack trace that starts at the caller of `New`.

To print the error message with its stack trace:

    fmt.Println(errors.String(err))

### Errorf - Format and capture a cause

    errors.Errorf("function foo failed: %v", err)

A drop-in replacement for the Golang standard library `fmt.Errorf` function.

Returns an `error` whose `Error` function will return the given formatted
message. The error will have an attached stack trace that starts at the caller
of `Errorf`, and an attached cause equal to `err` (the last `error` passed in the
format args).

To print the error message with its stack trace and cause chain:

    fmt.Println(errors.String(err))

### NewTrace - Copy an error and capture a new stack trace

    errors.NewTrace(err, 0)

Attaches a new stack trace to an existing error, returning a new value and
leaving the original unmodified. The start of the new stack trace is controlled
by the second parameter, `skip`. When `skip` is 0, the trace begins with the
caller of `NewTrace`.

This is useful if you want to have a template error at the package level, and
return copies of it with unique stack traces.

To test whether a given error originated with your template error:

    template == errors.Original(err)

## Semaphore - A simple semaphore

A strict and a non-strict implementation of a basic semaphore, which feels like
something the standard library `sync` package should have provided.

The strict implementation disallows growing the size of the semaphore, and
detects attempts to acquire too many slots.

The non-strict implementation allows size changes.

Both implementations allow closing the semaphore to release all waiting routines
and signaling to them that their blocking operations failed.
