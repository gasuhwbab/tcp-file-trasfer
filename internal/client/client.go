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

	fileName := filepath.Base(filePath)
	fileNameLen := uint32(len([]byte(fileName)))

	fileNameFrame := &proto.Frame{
		Magic:      proto.WireMagic,
		Version:    proto.Version,
		Typ:        proto.TypeFileName,
		HdrLen:     proto.HdrLen,
		PayloadLen: fileNameLen,
		Payload:    []byte(fileName),
	}

	if err = proto.WriteFrame(fileNameFrame, conn); err != nil {
		log.Printf("error to write fileNameFrame %v", err)
		return
	}

	chunck := make([]byte, 1024)
	for {
		n, err := file.Read(chunck)
		if err == io.EOF {
			doneFrame := &proto.Frame{
				Magic:      proto.WireMagic,
				Version:    proto.Version,
				Typ:        proto.TypeDone,
				HdrLen:     proto.HdrLen,
				PayloadLen: 0,
				Payload:    nil,
			}
			err := proto.WriteFrame(doneFrame, conn)
			if err != nil {
				log.Printf("error to write doneFrame")
				return
			}
			break
		}
		if err != nil {
			log.Printf("error to read data from file %v", err)
			return
		}
		dataFrame := &proto.Frame{
			Magic:      proto.WireMagic,
			Version:    proto.Version,
			Typ:        proto.TypeData,
			HdrLen:     proto.HdrLen,
			PayloadLen: uint32(n),
			Payload:    chunck[:n],
		}
		err = proto.WriteFrame(dataFrame, conn)
		if err != nil {
			log.Printf("error to write dataFrame %v", err)
			return
		}
	}
	log.Printf("File successfully sent")
}
