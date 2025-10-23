package main

import (
	"flag"

	"github.com/gasuhwbab/tcp-file-transfer/internal/client"
)

func main() {
	host := flag.String("host", "localhost", "Host to connect to server")
	port := flag.Int("port", 8080, "Port to connect to server")
	filePath := flag.String("filePath", "test/test.txt", "Path to file to send")

	client := client.NewClient(*host, *port)
	client.SendMessageByFrame(*filePath)
}
