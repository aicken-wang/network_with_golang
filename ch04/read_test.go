package ch04

import (
	"crypto/rand"
	"io"
	"net"
	"testing"
)

func TestReadIntoBuffer(t *testing.T) {
	t.Log("TestRead")
	//1  定义一个16 MB的切片payload
	payload := make([]byte, 1<<24) // 16 MB
	_, err := rand.Read(payload)   // generate a random payload
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}
		defer conn.Close()
		// 2 将随机读取的内容写入到conn中
		_, err = conn.Write(payload)
		if err != nil {
			t.Error(err)
		}
	}()
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	// 3 创建一个 512 KB的切片 buf
	buf := make([]byte, 1<<19) // 512 KB
	for {
		//4 从网络中读取数据到切片buf中
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				t.Error(err)
			}
			break
		}
		// 打印切片中的字节个数
		t.Logf("buf:%s\n", string(buf[:n]))
		t.Logf("read %d bytes", n) // buf[:n] is the data read from conn
	}
	conn.Close()
}

// go test -v ./read_test.go
