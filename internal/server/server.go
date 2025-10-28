package server

import (
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gasuhwbab/tcp-file-transfer/internal/handshake"
	"github.com/gasuhwbab/tcp-file-transfer/internal/proto"
)

type Server struct {
	addr string
	ln   net.Listener
}

func NewServer(host string, port int) *Server {
	return &Server{addr: host + ":" + strconv.Itoa(port)}
}

func (server *Server) Run() error {
	ln, err := net.Listen("tcp", server.addr)
	if err != nil {
		return err
	}
	server.ln = ln
	log.Printf("Server runned on %s\n", server.addr)
	for {
		conn, err := server.ln.Accept()
		if err != nil {
			log.Printf("Error to create connection %v\n", err)
		}
		go server.handleConnectionByFrame(conn)
	}
}

func (server *Server) handleConnectionByFrame(conn net.Conn) {
	defer conn.Close()

	p := handshake.ServerParams{
		SupportedFeatures: 0,
		MaxFrame:          proto.MaxPayloadLen + uint32(proto.HdrLen),
		MaxChunck:         proto.MaxPayloadLen,
		MaxWindow:         proto.MaxPayloadLen + uint32(proto.HdrLen),
	}

	if _, err := handshake.Accpet(conn, p); err != nil {
		log.Printf("error to accept handshake %v", err)
		return
	}

	fileNameFrame, err := proto.ReadFrame(conn)
	if err != nil {
		log.Printf("error to read fileNameFrame %v\n", err)
		return
	}
	if fileNameFrame.Typ != proto.TypeFileName {
		log.Printf("bad frame type\n")
		return
	}

	receivedFileName := filepath.Join("test_received", string(fileNameFrame.Payload))
	if err := os.Mkdir("test_received", 0750); err != nil && !os.IsExist(err) {
		log.Printf("error to create directory %v\n", err)
		return
	}
	file, err := os.Create(receivedFileName)
	if err != nil {
		log.Printf("error to create file %v\n", err)
		return
	}
	defer file.Close()

	for {
		readFrame, err := proto.ReadFrame(conn)
		if err != nil {
			log.Printf("error to read frame %v\n", err)
			return
		}
		if readFrame.Typ == proto.TypeDone {
			break
		}
		if readFrame.Typ == proto.TypeData {
			chunk := readFrame.Payload
			if _, err := file.Write(chunk); err != nil {
				log.Printf("error to write chunk to data\n")
				return
			}
		}
	}
	log.Printf("File successfully received\n")
}
