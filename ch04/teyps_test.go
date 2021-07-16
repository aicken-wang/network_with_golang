package ch04

import (
	"bytes"
	"encoding/binary"
	"net"
	"reflect"
	"testing"
)

// 创建测试有效负载测试
func TestPayloads(t *testing.T) {
	//  1
	b1 := Binary("Clear is better than clever.")
	b2 := Binary("Don't panic.")
	//2
	s1 := String("Errors are values.")
	//3
	payloads := []Payload{&b1, &s1, &b2}
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		for _, p := range payloads {
			//4
			_, err = p.WriteTo(conn)
			if err != nil {
				t.Error(err)
				break
			}
		}
	}()
	// 完成测试payload测试
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < len(payloads); i++ {
		// 2
		actual, err := decode(conn)
		if err != nil {
			t.Fatal(err)
		}
		// 3
		if expected := payloads[i]; !reflect.DeepEqual(expected, actual) {
			t.Errorf("value mismatch: %v != %v", expected, actual)
			continue
		}
		//4
		t.Logf("[%T] %[1]q", actual)
	}
}

// 测试最大payload大小
func TestMaxPayloadSize(t *testing.T) {
	buf := new(bytes.Buffer)
	err := buf.WriteByte(BinaryType)
	if err != nil {
		t.Fatal(err)
	}
	// 1
	err = binary.Write(buf, binary.BigEndian, uint32(1<<30)) // 1 GB
	if err != nil {
		t.Fatal(err)
	}
	var b Binary
	_, err = b.ReadFrom(buf)
	// 2
	if err != ErrMaxPayloadSize {
		t.Fatalf("expected ErrMaxPayloadSize; actual: %v", err)
	}
}
