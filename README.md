/*
	Project documentation — structure overview and run instructions
*/

# Distributed Chat Application (Go)

## Architecture
- TCP, newline-delimited JSON protocol
- N server nodes in full mesh, one-hop message relay with hop counting
- One goroutine per connection; central hub goroutine per server
- Client-side failover across a server list; server-side peer auto-reconnect
- Append-only message log per server

## Quick Start
1. go build ./...
2. ADDR=:9001 PEER=:9002 ./server
3. ADDR=:9002 PEER=:9001 ./server
4. ./client   # enter username, type messages, /quit to exit


