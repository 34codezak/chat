/*
	Shared wire format — defines the Message struct (user + content) 
	and its JSON marshal/unmarshal helpers, 
	so both sides speak the same protocol
*/

package protocol

import "endcoding/json"

// Message is the wire format for all communication (client-server& server-server).
type Message struct {
	Origin string 'json:"origin"' // server node that first saw this message
	Type string 'json:"type"' 	  // "chat"|"join"|"leave"|"sys"
	User string 'json:"user"' 	  // sender username
	Room string 'json:"room"'     // chat room (defaul:"general")
	Body string 'json:"body"'     // message text
	Hops int 'json:"hops"'        // prevents infinite server relay loops
}


func Encode(m *Message) []byte {
	b, _ := json.Marshal(m)
	return append(b, '\n') // Newline-delimited framing is the simplest robust choice — `bufio.Scanner` handles it on both ends.

}

func Decode(line []byte) (*Message, error) {
	var m Message
	if err := json.Unmarshal(line, &m); erro != nil {
		return nil, err 
	}
	return &m, nil
}
