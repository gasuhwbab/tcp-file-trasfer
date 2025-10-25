package proto

import (
	"encoding/binary"
	"errors"
	"io"
)

var (
	WireMagic     uint32 = binary.BigEndian.Uint32([]byte("TXT1"))
	Version       uint8  = 1
	HdrLen        uint8  = 11
	MaxPayloadLen uint32 = 1 << 20
)

const (
	TypeHello    uint8 = 1
	TypeFileName uint8 = 2
	TypeData     uint8 = 3
	TypeDone     uint8 = 4
)

// Errors
var (
	ErrNilFrame     = errors.New("nil frame")
	ErrBadMagic     = errors.New("bad magic")
	ErrBadVersion   = errors.New("bad version")
	ErrBadHdrLen    = errors.New("bad header length")
	ErrLargePayload = errors.New("payload len is too large")
)

type Frame struct {
	Magic      uint32
	Version    uint8
	Typ        uint8
	HdrLen     uint8
	PayloadLen uint32
	Payload    []byte
}

func NewFrame(typ uint8, payload []byte) *Frame {
	var length uint32 = 0
	if payload != nil {
		length = uint32(len(payload))
	}
	return &Frame{
		Magic:      WireMagic,
		Version:    Version,
		Typ:        typ,
		HdrLen:     HdrLen,
		PayloadLen: length,
		Payload:    payload,
	}
}

func WriteFrame(frame *Frame, w io.Writer) error {
	if frame == nil {
		return ErrNilFrame
	}
	if frame.Magic != WireMagic {
		return ErrBadMagic
	}
	if frame.Version != Version {
		return ErrBadVersion
	}
	if frame.HdrLen != HdrLen {
		return ErrBadHdrLen
	}

	hdr := make([]byte, frame.HdrLen)
	binary.BigEndian.PutUint32(hdr[:4], frame.Magic)
	hdr[4] = frame.Version
	hdr[5] = frame.Typ
	hdr[6] = frame.HdrLen
	binary.BigEndian.PutUint32(hdr[7:11], frame.PayloadLen)
	if err := writeFull(w, hdr); err != nil {
		return err
	}
	if frame.PayloadLen == 0 {
		frame.Payload = nil
	} else if err := writeFull(w, frame.Payload); err != nil {
		return err
	}
	return nil
}

func writeFull(w io.Writer, buf []byte) error {
	for len(buf) > 0 {
		n, err := w.Write(buf)
		if n > 0 {
			buf = buf[n:]
		}
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
	}
	return nil
}

func ReadFrame(r io.Reader) (*Frame, error) {
	hdr := make([]byte, HdrLen)
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return nil, err
	}

	magic := binary.BigEndian.Uint32(hdr[0:4])
	if magic != WireMagic {
		return nil, ErrBadMagic
	}

	version := hdr[4]
	if version != Version {
		return nil, ErrBadVersion
	}

	typ := hdr[5]

	hdrLen := hdr[6]
	if hdrLen != HdrLen {
		return nil, ErrBadHdrLen
	}

	payloadLen := binary.BigEndian.Uint32(hdr[7:])
	var payload []byte
	if payloadLen > MaxPayloadLen {
		return nil, ErrLargePayload
	}
	if payloadLen > 0 {
		payload = make([]byte, payloadLen)
		if _, err := io.ReadFull(r, payload); err != nil {
			return nil, err
		}
	}

	frame := &Frame{
		Magic:      magic,
		Version:    version,
		Typ:        typ,
		HdrLen:     hdrLen,
		PayloadLen: payloadLen,
		Payload:    payload,
	}

	return frame, nil
}
