A small collection of low-level Golang packages to augment or replace bits of
the standard library.

# Components

## errors - Errors with stack traces and causes

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


## ln - A logging package with a natural interface

A new logging package that is a bit more flexible and easier to use than the
standard library `log` package.

An output log line will look something like this:

    I1203 10:04:59.846813 FuncName(filename.go:65) Message

The `I` indicates this is at Info level. It could also be a `W`, `E`, or `F` for
Warning, Error, or Fatal, respectively (the prefix is configurable).

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

The package also provides built-in `Warning`, `Error`, and `Fatal` loggers.

### Verbosity control

    ln.Verbosity = 5
    ln.PackageVerbosity["main"] = 2
    ln.V(3).Print("Message")

Verbosity is controlled by a global `Verbosity` level, and a package-specific
`PackageVerbosity` map. Package verbosity takes precedence.

A package name can be a short name like `http`, or a long name like `net/http`.
Short names can be ambiguous, so long names take precedence.

### Output locations

    ln.Info.LogTo(infoFile)
    ln.Warning.LogTo(ln.Info)
    ln.Error.LogTo(ln.Warning)
    ln.Fatal.LogTo(ln.Error, os.Stderr)

Output can be sent to any number of `io.Writer` implementations. A logger is
itself a writer, so you can direct loggers at each other (just don't make a
cycle).

### Send to a testing.T

    ln.Info.LogTo(ln.PrintWriter{t.Log})
    ln.Warning.LogTo(ln.Info)
    ln.Error.LogTo(ln.PrintWriter{t.Error})
    ln.Fatal.LogTo(ln.PrintWriter{t.Fatal})

For convenience, a `PrintWriter` struct is provided that can turn one of the
functions on a `testing.T` into an output location for log messages.

I don't actually recommend redirecting `Error` to `t.Error` unless you want error
messages to fail your test. Fatal I would probably leave alone, unless you are
writing a death test. In that case you may be better off redefining Fatal with
your own trigger function:

    ln.Fatal = ln.MakeLogger("F", ln.PrintWriter{t.Log}, func() {
      fatalCalled = true
    })

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
