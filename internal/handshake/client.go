package handshake

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/gasuhwbab/tcp-file-transfer/internal/proto"
)

type ClientParams struct {
	ProtoMinor       uint8
	FlagsRequired    uint8
	FlagsOptional    uint8
	MaxFrame         uint32
	MaxChunck        uint32
	MaxWindow        uint32
	IdleTimeout      uint16
	HeartBeatSec     uint16
	ConnNonce        [16]byte
	HandshakeTimeout time.Duration
}

type Approved struct {
	ProtoMinor   uint8
	Features     uint32
	MaxFrame     uint32
	MaxChunck    uint32
	MaxWindow    uint32
	IdleTimeout  uint16
	HeartBeatSec uint16
	ConnNonce    [16]byte
}

func Do(conn net.Conn, p ClientParams) (*Approved, error) {
	if p.HandshakeTimeout <= 0 {
		p.HandshakeTimeout = 10 * time.Second
	}
	deadline := time.Now().Add(p.HandshakeTimeout)
	err := conn.SetDeadline(deadline)
	if err != nil {
		return nil, err
	}

	if p.ConnNonce != [16]byte{} {
		if _, err := rand.Read(p.ConnNonce[:]); err != nil {
			return nil, fmt.Errorf("error to create ConnNonce %v\n", err)
		}
	}

	hello := proto.Hello{
		ProtoMinor:    p.ProtoMinor,
		FlagsRequired: p.FlagsRequired,
		Pad0:          0,
		MaxFrame:      p.MaxFrame,
		MaxChunck:     p.MaxChunck,
		MaxWindow:     p.MaxWindow,
		IdleTimeout:   p.IdleTimeout,
		HeartBeatSec:  p.HeartBeatSec,
		ConnNonce:     p.ConnNonce,
	}

	helloPayload := hello.MarshalBinary()
	if err := proto.WriteFrame(proto.NewFrame(proto.TypeHello, helloPayload), conn); err != nil {
		return nil, fmt.Errorf("error to write hello frame %v", err)
	}

	helloAckFrame, err := proto.ReadFrame(conn)
	if err != nil {
		return nil, fmt.Errorf("error to read hello ack frame %v", err)
	}
	if helloAckFrame.Typ != proto.TypeHelloAck {
		return nil, fmt.Errorf("bad type: expected: %d, received: %d", proto.TypeHelloAck, helloAckFrame.Typ)
	}
	var helloAck proto.HelloAck
	if err := helloAck.UnmarshalBinary(helloAckFrame.Payload); err != nil {
		return nil, fmt.Errorf("error to unmarshal helloAck %v", err)
	}
	if helloAck.ConnNonceEcho != p.ConnNonce {
		return nil, proto.ErrBadNonceEcho
	}
	reqMask := uint32(p.FlagsRequired)
	if (reqMask & ^helloAck.FeaturesAccepted) != 0 {
		return nil, errors.New("error: required features are not accepted")
	}
	if helloAck.MaxChunckAccepted+uint32(proto.HdrLen) > helloAck.MaxFrameAccpeted {
		return nil, errors.New("error: chucnk size is too large")
	}
	if err := conn.SetDeadline(time.Time{}); err != nil {
		return nil, err
	}

	approved := &Approved{
		ProtoMinor:   helloAck.ProtoMinorAccepted,
		Features:     helloAck.FeaturesAccepted,
		MaxFrame:     helloAck.MaxFrameAccpeted,
		MaxChunck:    helloAck.MaxChunckAccepted,
		MaxWindow:    helloAck.MaxWindowAccepted,
		IdleTimeout:  helloAck.IdleTimeoutAccepted,
		HeartBeatSec: hello.HeartBeatSec,
		ConnNonce:    p.ConnNonce,
	}
	return approved, nil
}
