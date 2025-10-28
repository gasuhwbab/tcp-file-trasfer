package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
)

func main() {
	connNonce, err := rand.Int(rand.Reader, big.NewInt(1<<16))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(binary.BigEndian.Uint16(connNonce.Bytes()))
}
