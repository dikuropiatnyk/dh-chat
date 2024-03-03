package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

type DHUser struct {
	userAddress   net.Addr
	serverAddress net.Addr
}

func (u *DHUser) Connect() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatalln("Connection error:", err)
		return
	}
	defer conn.Close()
	u.userAddress = conn.LocalAddr()
	u.serverAddress = conn.RemoteAddr()
	log.Println("Connected to", u.serverAddress)
	u.SendData(conn)
}

func (u *DHUser) SendData(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)

	// Read user input and send it to the server
	fmt.Print("Enter text: ")
	input, err := reader.ReadBytes('\n')
	if err != nil {
		log.Fatalln("Read error:", err)
	}
	// Convert the string to a byte slice and send it to the server
	sent_bytes, err := conn.Write(input)
	if err != nil {
		log.Println("Write error:", err)
		return
	}
	log.Printf("Sent %d bytes to %s\n", sent_bytes, u.serverAddress)

}

func main() {
	user := DHUser{}
	user.Connect()
}
