package main

import (
	"bytes"
	"fmt"
	"net"
	"time"

	pineapplenet "github.com/beijian128/pineapple/internal/net"
	"github.com/beijian128/pineapple/internal/router"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		fmt.Printf("connect failed: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("connected to server")

	codec := pineapplenet.NewCodec()

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				fmt.Printf("read error: %v\n", err)
				return
			}
			reader := bytes.NewBuffer(buf[:n])
			for reader.Len() > 0 {
				msg, err := codec.Decode(reader)
				if err != nil {
					fmt.Printf("decode error: %v\n", err)
					return
				}
				fmt.Printf("received msg: id=%d, data=%s\n", msg.MsgID, string(msg.Data))
			}
		}
	}()

	time.Sleep(500 * time.Millisecond)

	heartbeatMsg := &pineapplenet.Message{
		MsgID: router.MsgIDHeartbeat,
		Data:  []byte{},
	}
	data, _ := codec.Encode(heartbeatMsg)
	_, _ = conn.Write(data)
	fmt.Println("sent heartbeat")

	time.Sleep(500 * time.Millisecond)

	loginMsg := &pineapplenet.Message{
		MsgID: router.MsgIDLoginRequest,
		Data:  []byte(`{"username":"test","password":"123456"}`),
	}
	data, _ = codec.Encode(loginMsg)
	_, _ = conn.Write(data)
	fmt.Println("sent login request")

	time.Sleep(2 * time.Second)
}
