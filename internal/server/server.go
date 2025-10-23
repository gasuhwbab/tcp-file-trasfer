package server

import (
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"

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

	fileNameFrame, err := proto.ReadFrame(conn)
	if err != nil {
		log.Printf("error to read fileNameFrame %v", err)
		return
	}
	if fileNameFrame.Typ != proto.TypeFileName {
		log.Printf("bad frame type")
		return
	}

	receivedFileName := filepath.Join("test_received", string(fileNameFrame.Payload))
	if err := os.Mkdir("test_received", 0750); err != nil && !os.IsExist(err) {
		log.Printf("error to create directory %v", err)
		return
	}
	file, err := os.Create(receivedFileName)
	if err != nil {
		log.Printf("error to create file %v", err)
		return
	}
	defer file.Close()

	for {
		readFrame, err := proto.ReadFrame(conn)
		if err != nil {
			log.Printf("error to read frame %v", err)
			return
		}
		if readFrame.Typ == proto.TypeDone {
			break
		}
		if readFrame.Typ == proto.TypeData {
			chunk := readFrame.Payload
			if _, err := file.Write(chunk); err != nil {
				log.Printf("error to write chunk to data")
				return
			}
		}
	}
	log.Printf("File successfully received")
}
