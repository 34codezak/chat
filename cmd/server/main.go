/*
	Server entrypoint — creates the Hub, starts its run loop,
	and listens on :8080 for incoming connections
*/

package main

import (
	"log"
	"os"

	"github.com/34codezak/chat/internal/server"
)

func main() {
	addr := os.Getenv("ADDR")
	peer := os.Getenv("PEER")
	if addr == "" {
		log.Fatal("usage: ADDR=:9001 PEER=localhost:9002 ./server")
	}
	node := &server.Node{Addr: addr, PeerAddr: peer, Hub: server.NewHub(addr + ".log")}
	log.Fatal(node.Start())
}
