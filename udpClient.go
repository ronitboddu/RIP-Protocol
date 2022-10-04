package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	args := os.Args[1:]
	var addr_slice []string
	for i := 0; i < len(args); i++ {
		if args[i] == "-addr" {
			addr_slice = append(addr_slice, strings.TrimSpace(args[i+1]))
		}
	}
	for i := 0; i < len(addr_slice); i++ {
		Client(addr_slice[i])
	}
}

func Client(addr string) {
	p := make([]byte, 2048)
	conn, err := net.Dial("udp", addr+":4321")
	if err != nil {
		fmt.Printf("Some error %v", err)
		return
	}
	fmt.Fprintf(conn, "Hi UDP Server, How are you doing?")
	_, err = bufio.NewReader(conn).Read(p)
	if err == nil {
		fmt.Printf("%s\n", p)
	} else {
		fmt.Printf("Some error %v\n", err)
	}
	conn.Close()
}
