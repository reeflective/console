//go:build unix

package console

import (
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

// TestMonitorSignalsCustom verifies that monitorSignals honors a customized
// Console.Signals set. SIGUSR1 is used because it is not part of the default
// trapped set and is not sent by the test harness.
func TestMonitorSignalsCustom(t *testing.T) {
	c := New("test")
	c.Signals = []os.Signal{syscall.SIGUSR1}

	ch := c.monitorSignals()
	defer signal.Stop(ch)

	if err := syscall.Kill(os.Getpid(), syscall.SIGUSR1); err != nil {
		t.Fatalf("failed to raise SIGUSR1: %v", err)
	}

	select {
	case got := <-ch:
		if got != syscall.SIGUSR1 {
			t.Fatalf("received %v, want SIGUSR1", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for the custom signal")
	}
}
