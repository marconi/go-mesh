package main

import (
	// "bytes"
	"fmt"
	// "github.com/marconi/go-mesh/gomesh/utils"
	// "reflect"
	// "encoding/binary"
	// "net"
	// "os"
	// "errors"
	// "strings"
)

func main() {
	// guid := utils.GenPeerId()

	// buff := new(bytes.Buffer)
	// buff.Write(guid[:])
	// buff.WriteByte(byte(0x00))
	// buff.WriteByte(byte(5))
	// buff.WriteByte(byte(0))

	// p_len := make([]byte, 3)
	// binary.LittleEndian.PutUint16(p_len, uint16(5))
	// buff.Write(p_len)

	// err := binary.Write(buff, binary.LittleEndian, uint16(5))
	// if err != nil {
	//  fmt.Println("binary.Write failed: ", err)
	// }

	// p_len_buf := new(bytes.Buffer)
	// err := binary.Write(p_len_buf, binary.LittleEndian, pi)

	// buff.WriteByte(p_len)

	// messages := []byte{0x00, 0x01, 0x02, 0x40, 0x80, 0x81}
	// for m := range messages {
	//  m := byte(m)

	//  buff.WriteByte(m)

	//  m_type := reflect.TypeOf(m)
	//  fmt.Println(m, m_type, int(m_type.Size()), "bytes")
	//  fmt.Println("-------------------------")
	// }

	// buff.Write([]byte(pong))

	// buff.WriteByte(byte(5))
	// buff.WriteByte(byte(0))
	// buff.WriteByte(byte(0))

	// fmt.Println("raw: ", buff)
	// fmt.Println("buff:", buff.Bytes(), buff.Len(), "bytes")

	// raw_msg := buff.Bytes()
	// guidx := utils.FormatGuid(raw_msg[:16])
	// p_type := raw_msg[16]
	// ttl := raw_msg[17]
	// hops := raw_msg[18]
	// p_lenx := raw_msg[19:]

	// fmt.Println(guidx, p_type, ttl, hops, p_lenx)

	// p_lenxx := binary.LittleEndian.Uint16(p_lenx)
	// fmt.Println(p_lenxx)

	// buf := new(bytes.Buffer)
	// p_len := make([]byte, 3)
	// binary.LittleEndian.PutUint16(p_len, uint16(4096))
	// buf.Write(p_len)

	// bites := buf.Bytes()
	// fmt.Println("raw: ", buf)
	// fmt.Println("buff:", bites, buf.Len(), "bytes")

	// p_len_x := binary.LittleEndian.Uint16(bites[:3])
	// fmt.Println(p_len_x)

	s := "hello"
	fmt.Println(s)
	s = "World"
	// fmt.Println(s)
}
