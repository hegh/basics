package ln

import (
	"os"
	"syscall"
	"time"
)

// AbortMe sends SIGABRT to this process. May not succeed.
//
// The process may not terminate immediately (or at all) on SIGABRT.
func AbortMe() error {
	me, err := os.FindProcess(os.Getpid())
	if err != nil {
		return err
	}
	if err = me.Signal(syscall.SIGABRT); err != nil {
		return err
	}
	return nil
}

// Terminate is the default trigger attached to the Fatal logger.
//
// It first tries to send SIGABRT to this process using AbortMe. If that
// fails, or if the process does not die after a few seconds, then it forces
// termination with os.Exit(1).
//
// This function will not return.
func Terminate() {
	defer os.Exit(1)
	if err := AbortMe(); err != nil {
		Error.Printf("AbortMe: failed: %v", err)
		return
	}

	// Sleep a moment to give the SIGABRT time to kill the process.
	time.Sleep(30 * time.Second)
}
