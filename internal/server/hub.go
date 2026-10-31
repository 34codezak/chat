/*
	Message router — central event loop that manages the set of connected nodes
	and fans out every incoming message to all other clients (the pub/sub core)
*/

package server

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"sync"

	"github.com/34codezak/chat/internal/protocol"
)

// Client represents one connected end-user
type Client struct {
	Name string
	Conn net.Conn
}

// Hub manages local clients, peer server, and the message log.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
	logFile *os.File

	// outbound channel: everything to be broadcasted locally + relayed to peers
	Broadcast  chan *protocol.Message
	Register   chan *Client
	Unregister chan *Client

	PeerConns map[string]net.Conn // peer address -> connection (set up by Node)
}

func NewHub(logPath string) *Hub {
	h := &Hub{
		clients:    make(map[*Client]struct{}),
		Broadcast:  make(chan *protocol.Message, 64),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		PeerConns:  make(map[string]net.Conn),
	}
	if f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); err == nil {
		h.logFile = f // message persistence (fault tolerance step)
	}
	return h
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.Register:
			h.mu.Lock()
			h.clients[c] = struct{}{}
			h.mu.Unlock()
			h.Broadcast <- &protocol.Message{Type: "sys", Body: c.Name + " joined"}
			log.Printf("client %s connected", c.Name)

		case c := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				_ = c.Conn.Close()
			}
			h.mu.Unlock()
			h.Broadcast <- &protocol.Message{Type: "sys", Body: c.Name + " left"}

		case msg := <-h.Broadcast:
			if msg == nil {
				continue
			}
			h.persist(msg)
			h.localBroadcast(msg)
			h.relayToPeers(msg)
		}
	}
}

func (h *Hub) persist(m *protocol.Message) {
	if h.logFile == nil || m == nil {
		return
	}
	b, _ := json.Marshal(m)
	_, _ = h.logFile.Write(append(b, '\n'))
}

func (h *Hub) localBroadcast(m *protocol.Message) {
	if m == nil {
		return
	}
	b := protocol.Encode(m)
	h.mu.RLock()
	for c := range h.clients {
		// never let one slow client block the hub
		if c == nil || c.Conn == nil {
			continue
		}
		_ = c.Conn.SetWriteDeadline(deadline())
		if _, err := c.Conn.Write(b); err != nil {
			go func(c *Client) { h.Unregister <- c }(c)
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) relayToPeers(m *protocol.Message) {
	if m == nil || m.Hops >= 1 { // do not re-relay a message that came from a peer
		return
	}
	fwd := *m
	fwd.Hops++
	b := protocol.Encode(&fwd)
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, conn := range h.PeerConns {
		if conn == nil {
			continue
		}
		_, _ = conn.Write(b)
	}
}
