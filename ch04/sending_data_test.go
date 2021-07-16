package ch04

import (
	"log"
	"net"
	"testing"
	"time"
)

// 通过连接发送字符串“hello世界”
var (
	n int
	i = 7 // maximum number of retries
)

func TestSendingData(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		t.Fatalf("error:%v", err)
		return
	}
	defer conn.Close()

	for ; i > 0; i-- {
		//2
		n, err = conn.Write([]byte("hello world"))
		if err != nil {
			// 4
			if nErr, ok := err.(net.Error); ok && nErr.Temporary() {
				log.Println("temporary error:", nErr)
				time.Sleep(10 * time.Second)
				continue
			}
			t.Fatalf("error:%v\n", err)
			return
		}
		break
	}
	if i == 0 {
		t.Fatalf("%s\n", "temporary write failure threshold exceeded")
		return
	}
	log.Printf("wrote %d bytes to %s\n", n, conn.RemoteAddr())
}
