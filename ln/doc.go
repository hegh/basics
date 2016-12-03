// Package ln provides flexible, easy to use logging with a minimal number of
// features to keep it easy to understand.
//
// An output line will look something like this:
//  I1203 10:04:59.846813 FuncName(filename.go:65) Message
//
// I: The logging level (Info in this case). Other built-in values are W for
//    Warning, E for Error, and F for Fatal.
// 1203: The date, MMDD (December 3rd).
// 10:04:59.846813: Timestamp, hh:mm:ss.micros
// FuncName: The name of the function that logged the message.
// filename.go: The name of the file that logged the message.
// 65: The line number that logged the message.
// Message: The message that was logged.
//
// Usage without format strings:
//   ln.V(1).Print("debug message")
//   ln.Info("info message")
//   ln.Warning("warning message")
//   ln.Error("error message")
//   ln.Fatal("fatal message")
//
// Usage with format strings:
//   ln.V(1).Printf("debug %s", "message")
//   ln.Info.Printf("info %v", "message")
//   ln.Warning.Printf("warning %q", "message")
//   ln.Error.Printf("error %s", "message")
//   ln.Fatal.Printf("fatal %s", "message")
//
// Setting the debug level:
//   ln.Verbosity = 5
//   ln.PackageVerbosity["main"] = 2
//   delete(ln.PackageVerbosity, "test")
//
// Setting up output locations:
//   ln.Info.LogTo(infoFile)
//   ln.Warning.LogTo(warningFile, ln.Info)
//   ln.Error.LogTo(errorFile, ln.Warning)
//   ln.Fatal.LogTo(os.Stderr, ln.Error)
//
// Now ln.Fatal("msg") goes to stderr, errorFile, warningFile, and infoFile.
// ln.Error("msg") goes to errorFile, warningFile, and infoFile.
// ln.Warning("msg") goes to warningFile and infoFile, and
// ln.Info("msg") and ln.V(0).Print("msg") go to infoFile.
//
// Setting up output to go through a testing.T:
//   ln.Info = ln.MakeLogger("I", ln.PrintLogger{t.Log}, nil)
//   ln.Warning = ln.MakeLogger("W", ln.PrintLogger{t.Log}, nil)
//   ln.Error = ln.MakeLogger("E", ln.PrintLogger{t.Error}, nil)
//   ln.Fatal = ln.MakeLogger("F", ln.PrintLogger{t.Fatal}, nil)
package ln
