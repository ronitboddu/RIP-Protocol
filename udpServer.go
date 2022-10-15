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
var m = sync.RWMutex{}
var down_routers []string

const INF = 999999999

type RouteEntry struct {
	Dest string
	Next string
	Cost int
}

var routing_table = make(map[string]RouteEntry)
var original_routingTable = make(map[string]RouteEntry)
var localAddr string

func sendResponse(conn *net.UDPConn, addr *net.UDPAddr) {
	m.Lock()
	s := conv_string(routing_table)
	n, err := conn.WriteToUDP([]byte(s), addr)
	if err != nil {
		fmt.Printf("Couldn't send response %v", err)
	}
	fmt.Println("Sent", n, "bytes", conn.LocalAddr(), "->", addr)
	fmt.Println()
	m.Unlock()
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
	for k, v := range routing_table {
		original_routingTable[k] = v
	}
	fmt.Println("Routing Table:")
	printRouting_table(routing_table)
	createServer(localAddr)
	time.Sleep(10 * time.Second)
	for {
		for key, _ := range routing_table {
			if key != localAddr {
				go actClient(key)
			}
		}
		time.Sleep(2 * time.Second)
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
		time.Sleep(1 * time.Second)
	}
}

func actClient(addr string) {
	conn, err := net.Dial("udp", addr+":4321")
	if err != nil {
		fmt.Printf("Some error %v", err)
		return
	}
	recieveFromServer(conn)
	conn.Close()
}

func printRouting_table(routing_table map[string]RouteEntry) {
	m.Lock()
	for _, element := range routing_table {
		str := "Dest:" + element.Dest + " | Next:" + element.Next + " | Cost:" + strconv.Itoa(element.Cost)
		if element.Cost == INF {
			str += " (Unreachable)"
		}
		fmt.Println(str)
	}
	fmt.Println()
	m.Unlock()
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

func recieveFromServer(conn net.Conn) {
	//m.Lock()
	p := make([]byte, 2048)
	var err error
	fmt.Fprintf(conn, "")
	remoteAddr := conn.RemoteAddr().String()
	_, err = bufio.NewReader(conn).Read(p)
	if err == nil {
		str := string(p[:])
		fmt.Println("Recieving routing table from " + remoteAddr[:len(remoteAddr)-5])
		recieved_table := conv_route(str)
		printRouting_table(recieved_table)
		updateRoutingTable(recieved_table, remoteAddr[:len(remoteAddr)-5], down_routers)
		fmt.Println("Routing Table:")
		printRouting_table(routing_table)

	} else {
		if !contains(down_routers, remoteAddr[:len(remoteAddr)-5]) {
			down_routers = append(down_routers, remoteAddr[:len(remoteAddr)-5])
		}
		//fmt.Printf("here is Some error %v\n", err)
		fmt.Printf("%v is Unreachable\n", remoteAddr[:len(remoteAddr)-5])
		poisonReverse(remoteAddr[:len(remoteAddr)-5])
		fmt.Println("Routing Table:")
		printRouting_table(routing_table)
		time.Sleep(2 * time.Second)
	}
	//m.Unlock()
}

func updateRoutingTable(m map[string]RouteEntry, next string, down_routers []string) {
	for key, element := range m {
		if key == element.Next && element.Cost == INF {
			down_routers = append(down_routers, key)
			routing_table[key] = RouteEntry{key, key, INF}
		}
		if !contains(down_routers, key) {
			if _, ok := routing_table[key]; ok {
				if routing_table[key].Cost > element.Cost+routing_table[next].Cost {
					routing_table[key] = RouteEntry{key, next, element.Cost + routing_table[next].Cost}
				}
			} else {
				routing_table[key] = RouteEntry{key, next, element.Cost + routing_table[next].Cost}
			}
		}

	}
}

func poisonReverse(addr string) {
	for k, v := range original_routingTable {
		routing_table[k] = v
	}
	for key, element := range routing_table {
		if key == addr && element.Next == addr {
			routing_table[key] = RouteEntry{key, key, INF}
			original_routingTable[key] = RouteEntry{key, key, INF}
		} else if element.Next == addr {
			// if val, ok := original_routingTable[key]; ok {
			// 	routing_table[key] = RouteEntry{key, key, val.Cost}
			// } else {
			routing_table[key] = RouteEntry{key, element.Next, INF}
			// }

		}
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
