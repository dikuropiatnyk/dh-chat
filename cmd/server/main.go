package main

import (
	"io"
	"log"
	"net"
)

type DHServer struct {
	addrress string
	listener net.Listener
}

func NewDHServer(addrress string) *DHServer {
	return &DHServer{addrress: addrress}
}

func (s *DHServer) Start() {
	listner, err := net.Listen("tcp", s.addrress)
	if err != nil {
		log.Fatalln("Bootup error:", err)
		return
	}
	log.Println("DHServer is starting at", s.addrress)
	defer listner.Close()
	s.listener = listner
	s.AcceptConnections()
}

func (s *DHServer) AcceptConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println("Connection error:", err)
			continue
		}
		go s.HandleConnection(conn)
	}
}

func (s *DHServer) HandleConnection(conn net.Conn) {
	defer conn.Close()
	log.Println("Received connection from", conn.RemoteAddr())
	buffer := make([]byte, 2048)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err.Error() == io.EOF.Error() {
				log.Println("Connection closed by client")
				break
			}
			log.Println("Read error: ", err)
		}
		log.Printf("%s => %s\n", conn.RemoteAddr(), string(buffer[:n]))
	}
}

func main() {
	server := NewDHServer("localhost:8080")
	server.Start()
}
