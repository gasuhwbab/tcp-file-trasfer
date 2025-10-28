package handshake

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/gasuhwbab/tcp-file-transfer/internal/proto"
)

type ServerParams struct {
	SupportedFeatures uint32
	MaxFrame          uint32
	MaxChunck         uint32
	MaxWindow         uint32
	IdleTimeout       uint16
	HeartBeatSec      uint16
	HandshakeRead     time.Duration
	HandshakeWrite    time.Duration
}

func Accpet(conn net.Conn, p ServerParams) (*Approved, error) {
	if p.HandshakeRead <= 0 {
		p.HandshakeRead = 10 * time.Second
	}
	if p.HandshakeWrite <= 0 {
		p.HandshakeWrite = 10 * time.Second
	}
	if err := conn.SetReadDeadline(time.Now().Add(p.HandshakeRead)); err != nil {
		return nil, fmt.Errorf("error to set read deadline: %v", err)
	}
	frame, err := proto.ReadFrame(conn)
	if err != nil {
		return nil, fmt.Errorf("error to read frame %v", err)
	}
	if frame.Typ != proto.TypeHello {
		return nil, fmt.Errorf("bad type got %d, want %d", frame.Typ, proto.TypeHello)
	}

	var hello proto.Hello
	if err := hello.UnmarshalBinary(frame.Payload); err != nil {
		return nil, fmt.Errorf("error to unmarshal hello frame payload: %v", err)
	}

	requiredFlags := uint32(hello.FlagsRequired)
	if (requiredFlags &^ p.SupportedFeatures) != 0 {
		return nil, errors.New("error unsupported required features")
	}

	wantFeatures := uint32(hello.FlagsRequired) | uint32(hello.FlagsOptional)
	accpetedFeatures := p.SupportedFeatures & wantFeatures
	disabledFeatures := wantFeatures &^ accpetedFeatures

	maxFrame := min(hello.MaxFrame, p.MaxFrame)
	if maxFrame == 0 || maxFrame < uint32(proto.HdrLen)+1 {
		maxFrame = uint32(proto.HdrLen) + 1
	}
	maxChunck := min(hello.MaxChunck, p.MaxChunck)
	if maxChunck == 0 {
		maxChunck = min(1<<16, maxFrame-uint32(proto.HdrLen))
	}
	if maxChunck+uint32(proto.HdrLen) > maxFrame {
		maxChunck = maxFrame - uint32(proto.HdrLen)
	}
	maxWindow := min(hello.MaxWindow, p.MaxWindow)

	idle := hello.IdleTimeout
	if idle == 0 {
		idle = p.IdleTimeout
	}
	hb := hello.HeartBeatSec
	if hb == 0 {
		hb = p.HeartBeatSec
	}

	helloAck := proto.HelloAck{
		ProtoMinorAccepted:  hello.ProtoMinor,
		FeaturesAccepted:    accpetedFeatures,
		MaxFrameAccpeted:    maxFrame,
		MaxChunckAccepted:   maxChunck,
		MaxWindowAccepted:   maxWindow,
		IdleTimeoutAccepted: idle,
		HeartBeatAccpeted:   hb,
		ConnNonceEcho:       hello.ConnNonce,
		DisableMask:         disabledFeatures,
	}

	payload := helloAck.MarshalBinary()

	if err := conn.SetWriteDeadline(time.Now().Add(p.HandshakeWrite)); err != nil {
		return nil, fmt.Errorf("error to set write timeout: %v", err)
	}

	if err := proto.WriteFrame(proto.NewFrame(proto.TypeHelloAck, payload), conn); err != nil {
		return nil, fmt.Errorf("error to write frame: %v", err)
	}
	if err := conn.SetDeadline(time.Time{}); err != nil {
		return nil, fmt.Errorf("error to cancel deadlines (SetDeadline(time.time{})): %v", err)
	}

	approved := &Approved{
		ProtoMinor:   helloAck.ProtoMinorAccepted,
		Features:     helloAck.FeaturesAccepted,
		MaxFrame:     helloAck.MaxFrameAccpeted,
		MaxChunck:    helloAck.MaxChunckAccepted,
		MaxWindow:    helloAck.MaxWindowAccepted,
		IdleTimeout:  helloAck.IdleTimeoutAccepted,
		HeartBeatSec: helloAck.HeartBeatAccpeted,
		ConnNonce:    hello.ConnNonce,
	}
	return approved, nil
}
