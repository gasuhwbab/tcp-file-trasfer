package server

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
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
		go server.handleConnection(conn)
	}
}

func (server *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	var receivedFileNameLen uint32
	if err := binary.Read(conn, binary.BigEndian, &receivedFileNameLen); err != nil {
		log.Printf("Error to read receivedFileNameLen %v\n", err)
		return
	}

	receivedFileName := make([]byte, receivedFileNameLen)
	if _, err := io.ReadFull(conn, receivedFileName); err != nil {
		log.Printf("Error to read receivedFileName %v\n", err)
		return
	}
	reseivedFileNameStr := filepath.Join("received", string(receivedFileName))

	if err := os.Mkdir("received", 0750); err != nil && !os.IsExist(err) {
		log.Printf("Error to creato directory %v\n", err)
		return
	}

	file, err := os.Create(reseivedFileNameStr)
	if err != nil {
		log.Printf("Error to create file %v\n", err)
		return
	}
	defer file.Close()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error to read data to buffer %v\n", err)
			return
		}

		if _, err := file.Write(buf[:n]); err != nil {
			log.Printf("Error to write data to file %v\n", err)
			return
		}
	}
	log.Println("File received and saved")
}
