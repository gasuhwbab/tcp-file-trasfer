package proto

import (
	"encoding/binary"
	"errors"
)

const (
	HelloLen    uint32 = 48
	HelloAckLen uint32 = 56
	ProtoMinor  uint8  = 1
)

const (
	FeatACK         uint32 = 1 << 0 // кумулятивные подтверждения
	FeatRESUME      uint32 = 1 << 1 // возобновление
	FeatCRC32       uint32 = 1 << 2 // CRC32 для DATA
	FeatHMAC        uint32 = 1 << 3 // HMAC для кадра
	FeatCOMPRESS    uint32 = 1 << 4 // сжатие
	FeatHEARTBEAT   uint32 = 1 << 5 // пинги при простое
	FeatTLSRequired uint32 = 1 << 6 // соединение должно быть под TLS
)

var (
	ErrBadHelloLen    = errors.New("bad HELLO len")
	ErrBadHelloAckLen = errors.New("bad HELLO_ACK len")
	ErrBadNonceEcho   = errors.New("bad nonce echo")
	ErrReserved       = errors.New("reserved bytes must be zero")
)

type Hello struct {
	ProtoMinor    uint8
	FlagsRequired uint8
	FlagsOptional uint8
	Pad0          uint8
	MaxFrame      uint32
	MaxChunck     uint32
	MaxWindow     uint32
	IdleTimeout   uint16
	HeartBeatSec  uint16
	ConnNonce     [16]byte
	Reserved      [12]byte
}

func (h *Hello) MarshalBinary() []byte {
	var payload [HelloLen]byte
	payload[0] = h.ProtoMinor
	payload[1] = h.FlagsRequired
	payload[2] = h.FlagsOptional
	payload[3] = 0
	binary.BigEndian.PutUint32(payload[4:8], h.MaxFrame)
	binary.BigEndian.PutUint32(payload[8:12], h.MaxChunck)
	binary.BigEndian.PutUint32(payload[12:16], h.MaxWindow)
	binary.BigEndian.PutUint16(payload[16:18], h.IdleTimeout)
	binary.BigEndian.PutUint16(payload[18:20], h.HeartBeatSec)
	copy(payload[20:36], h.ConnNonce[:])
	return payload[:]
}

func (h *Hello) UnmarshalBinary(payload []byte) error {
	if len(payload) != int(HelloLen) {
		return ErrBadHelloLen
	}
	h.ProtoMinor = payload[0]
	h.FlagsRequired = payload[1]
	h.FlagsOptional = payload[2]
	h.Pad0 = payload[3]
	h.MaxFrame = binary.BigEndian.Uint32(payload[4:8])
	h.MaxChunck = binary.BigEndian.Uint32(payload[8:12])
	h.MaxWindow = binary.BigEndian.Uint32(payload[12:16])
	h.IdleTimeout = binary.BigEndian.Uint16(payload[16:18])
	h.HeartBeatSec = binary.BigEndian.Uint16(payload[18:20])
	copy(h.ConnNonce[:], payload[20:36])
	copy(h.Reserved[:], payload[36:48])
	if h.Pad0 != 0 {
		return ErrReserved
	}
	for _, b := range h.Reserved {
		if b != 0 {
			return ErrReserved
		}
	}
	return nil
}

type HelloAck struct {
	ProtoMinorAccepted  uint8
	Pad0                uint8
	Pad1                uint16
	FeaturesAccepted    uint32
	MaxFrameAccpeted    uint32
	MaxChunckAccepted   uint32
	MaxWindowAccepted   uint32
	IdleTimeoutAccepted uint16
	HeartBeatAccpeted   uint16
	ConnNonceEcho       [16]byte
	DisableMask         uint32
	Reserved            [12]byte
}

func (h *HelloAck) MarshalBinary() []byte {
	var payload [HelloAckLen]byte
	payload[0] = h.ProtoMinorAccepted
	payload[1] = h.Pad0
	binary.BigEndian.PutUint16(payload[2:4], h.Pad1)
	binary.BigEndian.PutUint32(payload[4:8], h.FeaturesAccepted)
	binary.BigEndian.PutUint32(payload[8:12], h.MaxFrameAccpeted)
	binary.BigEndian.PutUint32(payload[12:16], h.MaxChunckAccepted)
	binary.BigEndian.PutUint32(payload[16:20], h.MaxWindowAccepted)
	binary.BigEndian.PutUint16(payload[20:22], h.IdleTimeoutAccepted)
	binary.BigEndian.PutUint16(payload[22:24], h.HeartBeatAccpeted)
	copy(payload[24:40], h.ConnNonceEcho[:])
	binary.BigEndian.PutUint32(payload[40:44], h.DisableMask)
	return payload[:]
}

func (h *HelloAck) UnmarshalBinary(payload []byte) error {
	if len(payload) != int(HelloAckLen) {
		return ErrBadHelloAckLen
	}
	h.ProtoMinorAccepted = payload[0]
	h.Pad0 = payload[1]
	h.Pad1 = binary.BigEndian.Uint16(payload[2:4])
	h.FeaturesAccepted = binary.BigEndian.Uint32(payload[4:8])
	h.MaxFrameAccpeted = binary.BigEndian.Uint32(payload[8:12])
	h.MaxChunckAccepted = binary.BigEndian.Uint32(payload[12:16])
	h.MaxWindowAccepted = binary.BigEndian.Uint32(payload[16:20])
	h.IdleTimeoutAccepted = binary.BigEndian.Uint16(payload[20:22])
	h.HeartBeatAccpeted = binary.BigEndian.Uint16(payload[22:24])
	copy(h.ConnNonceEcho[:], payload[24:40])
	h.DisableMask = binary.BigEndian.Uint32(payload[40:44])
	copy(h.Reserved[:], payload[44:56])
	if h.Pad0 != 0 || h.Pad1 != 0 {
		return ErrReserved
	}
	for _, b := range h.Reserved {
		if b != 0 {
			return ErrReserved
		}
	}
	return nil
}
