package client

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
)

type Client struct {
	addr string
}

func NewClient(host string, port int) *Client {
	return &Client{addr: host + ":" + strconv.Itoa(port)}
}

func (client *Client) SendMessage(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error to open file %v\n", err)
		return
	}
	defer file.Close()

	conn, err := net.Dial("tcp", client.addr)
	if err != nil {
		log.Printf("Error to create connection %v\n", err)
		return
	}
	defer conn.Close()

	fileName := filepath.Base(filePath)
	fileNameLen := uint32(len([]byte(fileName)))

	if err := binary.Write(conn, binary.BigEndian, fileNameLen); err != nil {
		log.Printf("Error to write fileNameLen %v\n", err)
		return
	}
	if _, err := conn.Write([]byte(fileName)); err != nil {
		log.Printf("Error to write fileName %v\n", err)
	}

	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error to read file %v\n", err)
			return
		}
		if _, err := conn.Write(buf[:n]); err != nil {
			log.Printf("Error to write to connection %v\n", err)
			return
		}
	}
	log.Println("File successfully sent")
}
