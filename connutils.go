package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

func readPacket(conn net.Conn) (packet []byte, err error) {
	buffer := new(bytes.Buffer)
	cpyMin := func(amount int) (err error) {
		if buffer.Len() < amount {
			_, err = io.CopyN(buffer, conn, int64(amount-buffer.Len()))
		}
		return err
	}
	err = cpyMin(4)
	if err != nil {
		return nil, err
	}
	headerBuffer := buffer.Next(4)
	length := binary.BigEndian.Uint32(headerBuffer)
	fmt.Println("decoded packet length", length, int(length))

	err = cpyMin(int(length))
	if err != nil {
		return nil, err
	}
	buff := buffer.Next(int(length))
	return buff, nil
}

func writePacket(conn net.Conn, buff []byte) (err error) {
	length := len(buff)
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, uint32(length))
	_, err = conn.Write(header)
	if err != nil {
		return err
	}
	_, err = conn.Write(buff)
	return err
}
