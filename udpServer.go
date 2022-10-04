package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

type RouteEntry struct {
	dest string
	next string
	cost int
}

var routing_table = make(map[string]RouteEntry)

func sendResponse(conn *net.UDPConn, addr *net.UDPAddr) {
	_, err := conn.WriteToUDP([]byte("From server: Hello I got your message "), addr)
	if err != nil {
		fmt.Printf("Couldn't send response %v", err)
	}
}

func main() {
	args := os.Args[1:]
	var addr_slice []string
	var localAddr string
	var cost int
	for i := 0; i < len(args); i++ {
		if args[i] == "-localaddr" {
			localAddr = args[i+1]
			routing_table[localAddr] = RouteEntry{localAddr, localAddr, 0}
		}
		if args[i] == "-addr" {
			addr_slice = append(addr_slice, strings.TrimSpace(args[i+1]))
			cost, _ = strconv.Atoi(args[i+2])
			routing_table[strings.TrimSpace(args[i+1])] = RouteEntry{strings.TrimSpace(args[i+1]),
				strings.TrimSpace(args[i+1]), cost}
		}
	}
	createServer(localAddr)
	time.Sleep(5 * time.Second)
	for key, _ := range routing_table {
		if key != localAddr {
			go actClient(key)
		}
	}
	wg.Wait()
}

func createServer(localAddr string) {
	//fmt.Println("in server")
	p := make([]byte, 2048)
	addr := net.UDPAddr{
		Port: 4321,
		IP:   net.ParseIP(localAddr),
	}
	//time.Sleep(5 * time.Second)
	ser, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return
	}
	//fmt.Println("after listen")
	wg.Add(1)
	go readUDP(ser, p)
	//fmt.Println("end servers")
}

func readUDP(ser *net.UDPConn, p []byte) {
	defer wg.Done()
	for {
		_, remoteaddr, err := ser.ReadFromUDP(p)
		fmt.Printf("Read a message from %v %s \n", remoteaddr, p)
		if err != nil {
			fmt.Printf("Some error  %v", err)
			continue
		}
		go sendResponse(ser, remoteaddr)

	}
}

func actClient(addr string) {
	p := make([]byte, 2048)
	//fmt.Println("here")
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
