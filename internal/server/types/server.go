package types

import (
	"io"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/dikuropiatnyk/dh-chat/internal/constants"
	"github.com/dikuropiatnyk/dh-chat/internal/server/actions"
	"github.com/dikuropiatnyk/dh-chat/pkg/communication"
)

type DHServer struct {
	addrress    string
	listener    net.Listener
	waitingPool map[string]*DHClient
	mut         sync.Mutex
}

func NewDHServer() *DHServer {
	return &DHServer{addrress: constants.SERVER_ADDRESS, waitingPool: make(map[string]*DHClient)}
}

func (s *DHServer) Start() {
	listner, err := net.Listen(constants.SERVER_CONNECTION_TYPE, s.addrress)
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
	buffer := make([]byte, constants.BUFFER_SIZE)
	// First reading from the connection to get the user name and the interlocutor
	userData, err := communication.ReadMessage(conn, buffer)
	if err != nil {
		return
	}
	// Split the user data into the user name and the interlocutor
	// The user data is in the format "userName;interlocutor"
	userDataSlice := strings.Split(userData, constants.DATA_SEPARATOR)
	userName, interlocutor := userDataSlice[0], userDataSlice[1]
	user := &DHClient{userAddress: conn.RemoteAddr(), name: userName, interlocutor: interlocutor}
	defer close(user.readChannel)

	// s.mut.Lock()
	// Check if the interlocutor is in the waiting pool
	availableUser, ok := s.waitingPool[interlocutor]
	if ok && availableUser.interlocutor == userName {
		user.readChannel, user.writeChannel = availableUser.writeChannel, availableUser.readChannel
		communication.SendMessage(conn, constants.INTERLOCUTOR_FOUND)
		user.writeChannel <- constants.INTERLOCUTOR_FOUND

		// Synchronize the chat between the current user and the interlocutor
		err := user.syncWithInterlocutor(conn, buffer)
		if err != nil {
			log.Println("Chat synchronization error:", err)
			return
		}
		// Remove the interlocutor from the waiting pool
		delete(s.waitingPool, interlocutor)
	} else {
		user.readChannel, user.writeChannel = make(chan string, 2), make(chan string, 2)
		s.waitingPool[userName] = user
		communication.SendMessage(conn, constants.NO_INTERLOCUTOR)
		// Set up a blocking waiter until the interlocutor is found, which is unblocked
		// by the interlocutor gorouting
		chat_secrets := <-user.readChannel
		communication.SendMessage(conn, chat_secrets)
		err := user.syncWithInterlocutor(conn, buffer)
		if err != nil {
			log.Println("Chat synchronization error:", err)
			return
		}
	}
	// s.mut.Unlock()

	ioReadChannel := make(chan string)
	errorChannel := make(chan error)

	go actions.ReadFromConnection(conn, buffer, ioReadChannel, errorChannel)

	// Here comes the actual chatting!
	for {
		select {
		case interlocutorMessage := <-user.readChannel:
			log.Printf("A[%s=>%s]: %s", user.interlocutor, user.name, interlocutorMessage)
			err := communication.SendMessage(conn, interlocutorMessage)
			if err != nil {
				log.Println("Couldn't send the message:", err)
				continue
			}
		case userMessage := <-ioReadChannel:
			log.Printf("B[%s=>%s]: %s", user.name, user.interlocutor, userMessage)
			user.writeChannel <- userMessage
		case err := <-errorChannel:
			if err.Error() == io.EOF.Error() {
				log.Println("Connection closed by client")
				return
			}
		}
	}
}
