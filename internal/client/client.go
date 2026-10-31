/*
	Client-side I/O — manages the TCP dial,
	exposes Send() to push a local user's message,
	and (in the read loop) prints incoming messages to the terminal
*/

package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/34codezak/chat/internal/protocol"
)

// Client holds the server list for failover.
type Client struct {
	Servers []string // e.g ["localhost:9001", "localhost:9002"]
}

func (c *Client) Run(username string) {
	conn := c.connectWithFailover()
	if conn == nil {
		fmt.Println("all servers unavailable, exiting")
		os.Exit(1)
	}

	// send join
	conn.Write(protocol.Encode(&protocol.Message{Type: "join", User: username}))

	// goroutine: read incoming messages and display them
	done := make(chan struct{})
	go func() {
		defer close(done)
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			msg, err := protocol.Decode(scanner.Bytes())
			if err != nil {
				continue
			}
			switch msg.Type {
			case "chat":
				fmt.Printf("\r[%s] %s:%s\n>", msg.Room, msg.User, msg.Body)
			case "sys":
				fmt.Printf("\r* %s\n>", msg.Body)
			}
		}
	}()

	// main loop: read user input and send
	input := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for input.Scan() {
		text := strings.TrimSpace(input.Text())
		if text == "/quit" {
			conn.Close()
			return
		}

		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		conn.Write(protocol.Encode(&protocol.Message{
			Type: "chat", User: username, Room: "general", Body: text,
		}))
		fmt.Print("> ")
	}
}

// connectWtihFailover tries each servr in order - the fault tolerance mechanism.
func (c *Client) connectWithFailover() net.Conn {
	for _, addr := range c.Servers {
		conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
		if err != nil {
			fmt.Printf("server %s unreachable: %v\n", addr, err)
			continue
		}
		fmt.Printf("connected to %s\n", addr)
		return conn
	}
	return nil
}
