package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

type RouteEntry struct {
	Dest string
	Next string
	Cost int
}

var routing_table = make(map[string]RouteEntry)
var localAddr string

func sendResponse(conn *net.UDPConn, addr *net.UDPAddr) {
	s := conv_string(routing_table)
	n, err := conn.WriteToUDP([]byte(s), addr)
	if err != nil {
		fmt.Printf("Couldn't send response %v", err)
	}
	fmt.Println("Sent", n, "bytes", conn.LocalAddr(), "->", addr)
}

func main() {
	args := os.Args[1:]
	var addr_slice []string

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
	time.Sleep(10 * time.Second)
	for key, _ := range routing_table {
		if key != localAddr {
			go actClient(key)
		}
	}
	wg.Wait()
}

func createServer(localAddr string) {
	p := make([]byte, 2048)
	addr := net.UDPAddr{
		Port: 4321,
		IP:   net.ParseIP(localAddr),
	}
	ser, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return
	}
	wg.Add(1)
	go readUDP(ser, p)

}

func readUDP(ser *net.UDPConn, p []byte) {
	defer wg.Done()
	for {
		_, remoteaddr, err := ser.ReadFromUDP(p)
		fmt.Printf("Sending routing table to %v \n", remoteaddr)
		//fmt.Printf("Read a message from %v %s \n", remoteaddr, p)
		if err != nil {
			fmt.Printf("Some error  %v", err)
			continue
		}
		go sendResponse(ser, remoteaddr)
	}
}

func actClient(addr string) {
	p := make([]byte, 2048)
	conn, err := net.Dial("udp", addr+":4321")
	if err != nil {
		fmt.Printf("Some error %v", err)
		return
	}
	fmt.Fprintf(conn, "")
	_, err = bufio.NewReader(conn).Read(p)
	if err == nil {
		str := string(p[:])
		fmt.Println("Recieving routing table from " + conn.RemoteAddr().String())
		printRouting_table(conv_route(str))
	} else {
		fmt.Printf("Some error %v\n", err)
	}
	conn.Close()
}

func printRouting_table(routing_table map[string]RouteEntry) {
	for _, element := range routing_table {
		fmt.Println("Dest:" + element.Dest + " | Next:" + element.Next + " | Cost:" + strconv.Itoa(element.Cost))
	}
}

func conv_string(m map[string]RouteEntry) string {
	s := ""
	for key, element := range m {
		s += " dest " + key + " next " + element.Next + " cost " + strconv.Itoa(element.Cost) + "\n"
	}
	return s
}

func conv_route(s string) map[string]RouteEntry {
	r := make(map[string]RouteEntry)
	var err error
	arr := strings.Split(s, "\n")
	for i := 0; i < len(arr); i++ {
		var routeEntry RouteEntry
		temp := strings.Split(arr[i], " ")
		for j := 0; j < len(temp); j++ {
			if temp[j] == "dest" {
				routeEntry.Dest = temp[j+1]
			} else if temp[j] == "next" {
				routeEntry.Next = temp[j+1]
			} else if temp[j] == "cost" {
				routeEntry.Cost, err = strconv.Atoi(temp[j+1])
				if err != nil {
					log.Fatal(err)
				}
			}
		}
		if routeEntry.Dest != "" {
			r[routeEntry.Dest] = routeEntry
		}
	}
	return r
}
