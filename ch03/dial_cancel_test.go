/*
 * @Author: your name
 * @Date: 2021-07-16 01:25:37
 * @LastEditTime: 2021-07-16 01:26:31
 * @LastEditors: Please set LastEditors
 * @Description: In User Settings Edit/
 * @FilePath: \godemo\dial_cancel_test.go
 */
package ch03

import (
	"context"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestDialContextCancel(t *testing.T) {
	// 1
	ctx, cancel := context.WithCancel(context.Background())
	sync := make(chan struct{})
	// 2
	go func() {
		defer func() { sync <- struct{}{} }()
		var d net.Dialer
		d.Control = func(_, _ string, _ syscall.RawConn) error {
			time.Sleep(time.Second)
			return nil
		}
		conn, err := d.DialContext(ctx, "tcp", "10.0.0.1:80")
		if err != nil {
			t.Log(err)
			return
		}
		conn.Close()
		t.Error("connection did not time out")
	}()
	// 3
	cancel()
	<-sync
	// 4
	if ctx.Err() != context.Canceled {
		t.Errorf("expected canceled context; actual: %q", ctx.Err())
	}
}
