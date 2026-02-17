package main

import (
	"fmt"
	"log"
	"net"

	"github.com/elmq0022/kv-store/internal/executor"
	"github.com/elmq0022/kv-store/internal/resp"
	"github.com/elmq0022/kv-store/internal/storage"
)

func main() {
	var s storage.Storage = storage.NewInMemoryShardedStorage()
	var exe = executor.NewExecutor(s)

	ln, err := net.Listen("tcp", ":6379")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	fmt.Println("listening on :6379")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}
		go handleConn(conn, exe)
	}
}

func handleConn(conn net.Conn, exe *executor.Executor) {
	defer conn.Close()

	decoder := resp.NewDecoder(conn)
	encoder := resp.NewEncoder(conn)

	for {
		input, err := decoder.Decode()
		if err != nil {
			return
		}

		output, err := exe.Execute(input)
		if err != nil {
			encoder.Encode(resp.Value{Type: resp.TypeError, Bytes: []byte(err.Error())})
			return
		}

		if err := encoder.Encode(output); err != nil {
			return
		}
	}
}
