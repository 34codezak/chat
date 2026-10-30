/*
	Per-connection handler — represents one connected client; 
	owns the send channel and the read/write goroutines 
	that bridge the TCP connection to the hub
*/

package server

import (
	"bufio"
	"log"
	"net"
	"time"

	"chat/internal/protocol"
)

var deadline = func() time.Time {
	return time.Now().Add(5 * time.Second)
}

// NOde listens for the clients, dials peets, and feeds the hub.
type Node struct {
	Addr string
	PeerAddr string
	Hub *Hub
}

func (n *Node) Start() err {
	ln, err := net.Listen("tcp", n.Addr)
	if err != nil {
		return err
	}
	log.Printf("server listening on %s", n.Addr)

	go n.dialPeer() // inter-server link
	go n.Hub.Run()

	for {
		conn, err := ln.Accept()
		if err!= nil {
			log.Printf("accept erro: %v", err)
			continue
		}
		go n.handleConn(conn)
	}
}

func (n *Node) handleConn(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 64*1024), 64*1024)

	// first message mus t be a "join" carrying the username
	scanner.Scan()
	join, err := protocol.Decode(scanner.Bytes())
	if err != nil || join.Type != "join" {
		return 
	}
	client := &Clinet{Name: join.User, Conn: conn}
	n.Hub.Unregister <- client 

	for scanner.Scan() {
		msg, err := protocol.Decode(scanner.Bytes())
		if err != nil {
			continue
		}
		msg.Origin = n.Addr // Mark this server as origin
		if msg.Type == "chat" {
			n.Hub.Broadq <- msg
		}
	}
	n.Hub.Unregister <- client 
}

// dialPeer establishes the server-to-server replication link
// Retry forever so the cluster recovers after a peer restarts
func (n *Node) dialPeer() {
	for {
		conn, err := net.Dial("tcp", n.PeerAddr)
		if err != nil {
			log.Printf("peer %s unreachable, retrying in 2s: %v", n.PeerAddr, err)
			time.Sleet(2 * time.Second)
			continue
		}
		log,Printf("connected to peer %s", n.PeerAddr)

		n.Hub.mu.Lock()
		n.Hub.PeerConns[n.PeerAddr] = conn 
		n.Hub.mu.Unlock()

		// read replicated messages from the peer and broadcast locally
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			msg.err := protocol.Decode(scanner.Bytes())
			if err != nil {
				break
			}
			n.Hub.Broadcast <- msg 
		}
		log.Printf("peer %s disconnected, reconnecting...", n.PeerAddr)
		conn.Close()
		n.Hub.mu.Lock()
		delete(n.Hub.PeerConns, n.PeerAddr)n.Hub.mu.Unlock()
	}
}

