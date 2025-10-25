package client

import (
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gasuhwbab/tcp-file-transfer/internal/proto"
)

type Client struct {
	addr string
}

func NewClient(host string, port int) *Client {
	return &Client{addr: host + ":" + strconv.Itoa(port)}
}

func (client *Client) SendMessageByFrame(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("error to open file %v", err)
		return
	}
	defer file.Close()

	conn, err := net.Dial("tcp", client.addr)
	if err != nil {
		log.Printf("error to connect to server %v", err)
		return
	}
	defer conn.Close()

	// Frame TypeHello
	if err = proto.WriteFrame(proto.NewFrame(proto.TypeHello, nil), conn); err != nil {
		log.Printf("Error to writeHelloFrame %v", err)
		return
	}

	// Frame TypeFileName
	if err = proto.WriteFrame(
		proto.NewFrame(proto.TypeFileName, []byte(filepath.Base(filePath))),
		conn,
	); err != nil {
		log.Printf("Error to writeFileNameFrame %v", err)
		return
	}

	chunck := make([]byte, 1<<16)
	for {
		n, err := file.Read(chunck)
		if err == io.EOF {
			// Frame TypeDone
			if n > 0 {
				if err := proto.WriteFrame(proto.NewFrame(proto.TypeData, chunck[:n]), conn); err != nil {
					log.Printf("error to write data frame %v", err)
					return
				}
			}
			if err := proto.WriteFrame(proto.NewFrame(proto.TypeDone, nil), conn); err != nil {
				log.Printf("error to write doneFrame")
				return
			}
			break
		}
		if err != nil {
			log.Printf("error to read data from file %v", err)
			return
		}
		// Frame TypeData
		if err := proto.WriteFrame(proto.NewFrame(proto.TypeData, chunck[:n]), conn); err != nil {
			log.Printf("error to write data frame %v", err)
			return
		}
	}
	log.Printf("File successfully sent")
}
