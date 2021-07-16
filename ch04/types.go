package ch04

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

/*
	The message struct implements a simple protocol (types.go).
	这个消息结构实现了一个简单的协议
*/
// 创建二进制文件类型
const (
	// 消息类型
	BinaryType uint8 = iota + 1
	// uint8 = iota + 1 字符串类型
	StringType
	//最大的有效载荷 即值的最大长度
	MaxPayloadSize uint32 = 10 << 20 // 10 MB
)

var ErrMaxPayloadSize = errors.New("maximum payload size exceeded")

// 接口定义
type Payload interface {
	fmt.Stringer
	io.ReaderFrom
	io.WriterTo
	Bytes() []byte
}

// Binary是一个字节切片
type Binary []byte

// 实现Bytes接口
func (m Binary) Bytes() []byte { return m }

// 实现String接口
func (m Binary) String() string { return string(m) }

// 实现WriteTo接口
func (m Binary) WriteTo(w io.Writer) (int64, error) {
	// 调用 Write方法写入 1 byte
	err := binary.Write(w, binary.BigEndian, BinaryType) // 1-byte type
	if err != nil {
		return 0, err
	}
	var n int64 = 1
	// BigEndian是ByteOrder的大端实现
	err = binary.Write(w, binary.BigEndian, uint32(len(m))) // 4-byte size
	if err != nil {
		return n, err
	}
	n += 4
	// 写入数据
	o, err := w.Write(m) // payload
	return n + int64(o), err
}

// 完成二进制文件类型的实现
func (m *Binary) ReadFrom(r io.Reader) (int64, error) {
	var typ uint8
	// eadFrom方法将1个字节从读取器读入类型变量 typ
	err := binary.Read(r, binary.BigEndian, &typ) // 1-byte type
	if err != nil {
		return 0, err
	}
	var n int64 = 1
	// 验证该类型是否为 BinaryType类型
	if typ != BinaryType {
		return n, errors.New("invalid Binary")
	}
	var size uint32
	// 将接下来的4个字节读入`size`变量
	err = binary.Read(r, binary.BigEndian, &size) // 4-byte size
	if err != nil {
		return n, err
	}
	n += 4
	// 防止DOS攻击耗尽主机内存
	if size > MaxPayloadSize {
		return n, ErrMaxPayloadSize
	}
	// 创建一个新的切片
	*m = make([]byte, size)
	// 填充二进制字节切片,返回读到的字节数和错误信息
	o, err := r.Read(*m) // payload
	if err != nil && err != io.EOF {
		fmt.Printf("error:%v\n", err)
	}
	return n + int64(o), err
}

// 创建字符串类型
type String string

// String实现的"Bytes"方法
func (m String) Bytes() []byte { return []byte(m) }

// String类型强制转换为string类型
// go语言是强类型，String别名必须通过string(m)转换
func (m String) String() string { return string(m) }

// 和Binary的WriteTo方法类似
func (m String) WriteTo(w io.Writer) (int64, error) {
	// 第一个字节是StringType
	err := binary.Write(w, binary.BigEndian, StringType) // 1-byte type
	if err != nil {
		return 0, err
	}
	var n int64 = 1
	// 它将String强制转换为字节切片
	err = binary.Write(w, binary.BigEndian, uint32(len(m))) // 4-byte size
	if err != nil {
		return n, err
	}
	n += 4
	// 写入writer
	o, err := w.Write([]byte(m)) // payload
	return n + int64(o), err
}

// 完成字符串类型的实现
func (m *String) ReadFrom(r io.Reader) (int64, error) {
	var typ uint8
	err := binary.Read(r, binary.BigEndian, &typ) // 1-byte type
	if err != nil {
		return 0, err
	}
	var n int64 = 1
	// 将typ变量与StringType进行比较
	if typ != StringType {
		return n, errors.New("invalid String")
	}
	var size uint32
	err = binary.Read(r, binary.BigEndian, &size) // 4-byte size
	if err != nil {
		return n, err
	}
	n += 4
	buf := make([]byte, size)
	o, err := r.Read(buf) // payload
	if err != nil {
		return n, err
	}
	//从r中Read的值强制转换为String
	*m = String(buf)
	return n + int64(o), nil
}

//将读取器中的字节解码为Binary类型或String类型
// 解码器的功能//将读取器中的字节解码为二进制类型或字符串类型(
func decode(r io.Reader) (Payload, error) {
	var typ uint8
	// 获取类型typ
	err := binary.Read(r, binary.BigEndian, &typ)
	if err != nil {
		return nil, err
	}
	// 创建一个Payload变量
	var payload Payload
	//从读取器中读取的是否为预期的类型常量BinaryType or StringType
	switch typ {
	case BinaryType:
		payload = new(Binary)
	case StringType:
		payload = new(String)
	default:
		return nil, errors.New("unknown type")
	}
	// 已经读到的字节与读取器连接起来
	_, err = payload.ReadFrom(io.MultiReader(bytes.NewReader([]byte{typ}), r))
	if err != nil {
		return nil, err
	}
	return payload, nil
}
