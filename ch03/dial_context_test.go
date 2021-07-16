package ch03

import (
	"context"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestDialContext(t *testing.T) {
	// 1
	dl := time.Now().Add(5 * time.Second)
	// 2
	ctx, cancel := context.WithDeadline(context.Background(), dl)
	// 3
	defer cancel()
	var d net.Dialer // DialContext is a method on a Dialer
	// 4
	d.Control = func(_, _ string, _ syscall.RawConn) error {
		// Sleep long enough to reach the context's deadline.
		time.Sleep(5*time.Second + time.Millisecond)
		return nil
	}
	// 5
	conn, err := d.DialContext(ctx, "tcp", "10.0.0.0:80")

	if err == nil {
		conn.Close()
		t.Fatal("connection did not time out")
	}

	nErr, ok := err.(net.Error)

	if !ok {
		t.Error(err)
	} else {
		if !nErr.Timeout() {
			t.Errorf("error is not a timeout: %v", err)
		}
	}
	// 6
	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("expected deadline exceeded; actual: %v", ctx.Err())
	}
}
