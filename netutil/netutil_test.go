package netutil

import (
	"net"
	"testing"

	"math/rand"

	"fmt"

	"github.com/xiaonanln/goworld/gwlog"
)

type testEchoTcpServer struct {
}

func (ts *testEchoTcpServer) ServeTCPConnection(conn net.Conn) {
	buf := make([]byte, 1024*1024, 1024*1024)
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			conn.Write(buf[:n])
		}

		if err != nil {
			if IsTemporaryNetError(err) {
				continue
			} else {
				gwlog.Error("read error: %s", err.Error())
				break
			}
		}
	}
}

func TestRawConnection(t *testing.T) {
	PORT := 4001
	go func() {
		ServeTCP(fmt.Sprintf("localhost:%d", PORT), &testEchoTcpServer{})
	}()

	_conn, err := net.Dial("tcp", "localhost:4001")
	if err != nil {
		t.Errorf("connect error: %s", err)
	}
	conn := NewRawConnection(_conn)
	var b byte
	for b = 0; b < 255; b++ {
		conn.SendByte(b)
		_b, err := conn.RecvByte()
		if err != nil || b != _b {
			t.Errorf("send byte but recv wrong")
		}
	}
	conn.Close()
}

func TestBinaryConnection(t *testing.T) {
	PORT := 4002
	go func() {
		ServeTCP(fmt.Sprintf("localhost:%d", PORT), &testEchoTcpServer{})
	}()

	_conn, err := net.Dial("tcp", "localhost:4002")
	if err != nil {
		t.Errorf("connect error: %s", err)
	}
	conn := NewBinaryConnection(_conn)
	for i := 0; i < 100; i++ {
		var v uint64 = uint64(rand.Int63())
		conn.SendUint64(v)
		rv, err := conn.RecvUint64()
		if err != nil || rv != v {
			t.Errorf("send %v but recv %v", v, rv)
		}
	}
}

func TestPacketConnection(t *testing.T) {
	PORT := 4003
	go func() {
		ServeTCP(fmt.Sprintf("localhost:%d", PORT), &testEchoTcpServer{})
	}()

	_conn, err := net.Dial("tcp", "localhost:4002")
	if err != nil {
		t.Errorf("connect error: %s", err)
	}

	conn := NewPacketConnection(_conn)

	for i := 0; i < 100; i++ {
		var PAYLOAD_LEN uint32 = uint32(rand.Intn(4096 + 1))
		gwlog.Info("Testing with payload %v", PAYLOAD_LEN)

		packet := conn.NewPacket()
		for j := uint32(0); j < PAYLOAD_LEN; j++ {
			packet.AppendByte(byte(rand.Intn(256)))
		}
		if packet.payloadLen != PAYLOAD_LEN {
			t.Errorf("payload should be %d, but is %d", PAYLOAD_LEN, packet.payloadLen)
		}
		conn.SendPacket(packet)
		recvPacket, err := conn.RecvPacket()
		if err != nil {
			t.Error(err)
		}
		if packet.payloadLen != recvPacket.payloadLen {
			t.Errorf("send packet len %d, but recv len %d", packet.payloadLen, recvPacket.payloadLen)
		}
		for i := uint32(0); i < packet.payloadLen; i++ {
			if packet.Payload()[i] != recvPacket.Payload()[i] {
				t.Errorf("send packet and recv packet mismatch on byte index %d", i)
			}
		}
	}
}
