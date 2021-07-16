// 清单3-3：建立与127.0.0.1的连接
package ch03

import (
	"io"
	"net"
	"testing"
)

func TestDial(t *testing.T) {
	// 在一个随机端口上创建一个侦听器
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	// 1 启动一个goroutine 接收新的连接
	go func() {

		defer func() {
			done <- struct{}{}
		}()

		for {
			//	2
			conn, err := listener.Accept()
			if err != nil {
				t.Log(err)
				return
			}

			// 3
			go func(c net.Conn) {
				defer func() {
					c.Close()
					done <- struct{}{}
				}()

				buf := make([]byte, 1024)
				for {
					// 4 读取数据
					n, err := c.Read(buf)
					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}
						return
					}
					t.Logf("received: %q", buf[:n])
				}
			}(conn)
		}
	}()
	// 5
	// 6
	// 7
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	// 8
	conn.Close()
	<-done
	// 9
	listener.Close()
	<-done
}
