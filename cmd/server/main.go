package main

import (
	"log"
	"net"
	"strings"
	"time"

	"github.com/dikuropiatnyk/dh-chat/pkg/communication"
)

type DHServer struct {
	addrress    string
	listener    net.Listener
	waitingPool map[string]DHUser
}

type DHUser struct {
	userAddress  net.Addr
	name         string
	interlocutor string
	readChannel  chan string
	writeChannel chan string
}

func NewDHServer(addrress string) *DHServer {
	return &DHServer{addrress: addrress, waitingPool: make(map[string]DHUser)}
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
	// First reading from the connection to get the user name and the interlocutor
	userData, err := communication.ReadMessage(conn, buffer)
	if err != nil {
		log.Println("Couldn't get a user info:", err)
		return
	}
	// Split the user data into the user name and the interlocutor
	// The user data is in the format "userName;interlocutor"
	userDataSlice := strings.Split(userData, ":")
	userName, interlocutor := userDataSlice[0], userDataSlice[1]

	// Check if the interlocutor is in the waiting pool
	availableUser, ok := s.waitingPool[interlocutor]
	if ok && availableUser.interlocutor == userName {
		// If the interlocutor is in the waiting pool and the
		// interlocutor of the interlocutor is the current user
		// then we can start the chat
		// Create a new user
		user := DHUser{
			userAddress:  conn.RemoteAddr(),
			name:         userName,
			interlocutor: interlocutor,
			readChannel:  availableUser.writeChannel,
			writeChannel: availableUser.readChannel,
		}
		communication.SendMessage(conn, "INTERLOCUTOR_FOUND")
		user.writeChannel <- "INTERLOCUTOR_FOUND"
		time.Sleep(100 * time.Second)

	} else {
		// If the interlocutor is not in the waiting pool or
		// the interlocutor of the interlocutor is not the current user
		// then add the user to the waiting pool
		user := DHUser{
			userAddress:  conn.RemoteAddr(),
			name:         userName,
			interlocutor: interlocutor,
			readChannel:  make(chan string),
			writeChannel: make(chan string),
		}
		s.waitingPool[userName] = user
		communication.SendMessage(conn, "NO_INTERLOCUTOR")
		// Set up a blocking waiter until the interlocutor is found, which is unblocked
		// by the interlocutor gorouting
		chat_secrets := <-user.readChannel
		communication.SendMessage(conn, chat_secrets)
		time.Sleep(100 * time.Second)
	}
	// for {
	// 	n, err := conn.Read(buffer)
	// 	if err != nil {
	// 		if err.Error() == io.EOF.Error() {
	// 			log.Println("Connection closed by client")
	// 			break
	// 		}
	// 		log.Println("Read error: ", err)
	// 	}
	// 	log.Printf("%s => %s\n", conn.RemoteAddr(), string(buffer[:n]))
	// }
}

func main() {
	server := NewDHServer("localhost:8080")
	server.Start()
}
