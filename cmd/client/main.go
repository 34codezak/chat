/*
	Client entrypoint — dials the server,
	then reads lines from stdin
	and forwards them via Client.Send
*/

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/34codezak/chat/internal/client"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	c := &client.Client{Servers: []string{"localhost:9001", "localhost:9002"}}
	c.Run(username)
}
