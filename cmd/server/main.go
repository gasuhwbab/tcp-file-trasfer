package main

import (
	"flag"
	"log"

	"github.com/gasuhwbab/tcp-file-transfer/internal/server"
)

func main() {
	host := flag.String("host", "localhost", "Host to listen server")
	port := flag.Int("port", 8080, "Port to connect server")
	flag.Parse()

	server := server.NewServer(*host, *port)
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
