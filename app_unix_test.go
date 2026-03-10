//go:build !windows
// +build !windows

package web

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func TestApplication_Start_Signal(t *testing.T) {
	app := New(WithAddress(":0")).(*Application)

	done := make(chan error, 1)
	go func() {
		done <- app.Start()
	}()

	time.Sleep(50 * time.Millisecond)

	if err := syscall.Kill(os.Getpid(), syscall.SIGINT); err != nil {
		t.Fatalf("failed to send signal: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Start returned error after signal: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Start did not return after signal in time")
	}
}
