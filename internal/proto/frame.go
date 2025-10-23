package proto

import (
	"encoding/binary"
	"errors"
	"io"
)

var (
	WireMagic uint32 = binary.BigEndian.Uint32([]byte("TXT1"))
	Version   uint8  = 1
	HdrLen    uint8  = 11
)

const (
	TypeHello    uint8 = 1
	TypeFileName uint8 = 2
	TypeData     uint8 = 3
	TypeDone     uint8 = 4
)

// Errors
var (
	ErrNilFrame   = errors.New("nil frame")
	ErrBadMagic   = errors.New("bad magic")
	ErrBadVersion = errors.New("bad version")
	ErrBadHdrLen  = errors.New("bad header length")
)

type Frame struct {
	Magic      uint32
	Version    uint8
	Typ        uint8
	HdrLen     uint8
	PayloadLen uint32
	Payload    []byte
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
	if _, err := w.Write(hdr); err != nil {
		return err
	}
	if _, err := w.Write(frame.Payload); err != nil {
		return err
	}
	return nil
}

func ReadFrame(r io.Reader) (*Frame, error) {
	hdr := make([]byte, HdrLen)
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return nil, err
	}

	magic := binary.BigEndian.Uint32(hdr[0:4])
	version := hdr[4]
	typ := hdr[5]
	hdrLen := hdr[6]
	payloadLen := binary.BigEndian.Uint32(hdr[7:])

	var payload []byte
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
