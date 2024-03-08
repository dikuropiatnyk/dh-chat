package types

import (
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/dikuropiatnyk/dh-chat/internal/constants"
	"github.com/dikuropiatnyk/dh-chat/internal/server/actions"
	"github.com/dikuropiatnyk/dh-chat/pkg/communication"
)

type DHServer struct {
	addrress    string
	listener    net.Listener
	waitingPool map[string]*DHClient
	mut         sync.RWMutex
}

func NewDHServer() *DHServer {
	return &DHServer{addrress: constants.SERVER_ADDRESS, waitingPool: make(map[string]*DHClient)}
}

func (s *DHServer) CheckWaitingPool(interlocutorName string) (*DHClient, bool) {
	s.mut.RLock()
	defer s.mut.RUnlock()
	interlocutor, ok := s.waitingPool[interlocutorName]
	return interlocutor, ok
}

func (s *DHServer) AddClientToWaitingPool(clientName string, client *DHClient) {
	s.mut.Lock()
	s.waitingPool[clientName] = client
	s.mut.Unlock()
	log.Printf("Added %s to the waiting pool\n", clientName)
}

func (s *DHServer) DeleteClientFromWaitingPool(clientName string) {
	s.mut.Lock()
	delete(s.waitingPool, clientName)
	s.mut.Unlock()
	log.Printf("Deleted %s from the waiting pool\n", clientName)
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
	defer actions.CloseConnection(conn)
	log.Println("Received connection from", conn.RemoteAddr())
	buffer := make([]byte, constants.BUFFER_SIZE)
	// First reading from the connection to get the client name and the interlocutor
	clientData, err := communication.ReadMessage(conn, buffer)
	if err != nil {
		return
	}
	// Split the client data into the client name and the interlocutor
	// The client data is in the format "clientName;interlocutor"
	clientDataSlice := strings.Split(clientData, constants.DATA_SEPARATOR)
	clientName, interlocutor := clientDataSlice[0], clientDataSlice[1]

	client := NewDHClient(conn.RemoteAddr(), clientName, interlocutor)
	defer client.Close()

	// Check if the interlocutor is in the waiting pool
	availableClient, ok := s.CheckWaitingPool(interlocutor)
	if ok && availableClient.interlocutor == clientName {
		client.readChannel, client.writeChannel = availableClient.writeChannel, availableClient.readChannel
		if err = communication.SendMessage(conn, constants.INTERLOCUTOR_FOUND); err != nil {
			log.Println("Couldn't send the message:", err)
			return
		}
		client.writeChannel <- constants.INTERLOCUTOR_FOUND

		// Synchronize the chat between the current client and the interlocutor
		err := client.SyncWithInterlocutor(conn, buffer)
		if err != nil {
			log.Println("Chat synchronization error:", err)
			return
		}
	} else {
		client.readChannel, client.writeChannel = make(chan string, 2), make(chan string, 2)
		s.AddClientToWaitingPool(clientName, client)
		if err = communication.SendMessage(conn, constants.NO_INTERLOCUTOR); err != nil {
			log.Println("Couldn't send the message:", err)
			return
		}
		// Set up a blocking waiter until the interlocutor is found, which is unblocked
		// by the interlocutor goroutine
		select {
		case chat_secrets, ok := <-client.readChannel:
			s.DeleteClientFromWaitingPool(clientName)
			if !ok {
				log.Println(ErrReadChannelClosed)
				return
			}
			if err = communication.SendMessage(conn, chat_secrets); err != nil {
				log.Println("Couldn't send the message:", err)
				return
			}
			err := client.SyncWithInterlocutor(conn, buffer)
			if err != nil {
				log.Println("Chat synchronization error:", err)
				return
			}
		// If the interlocutor doesn't show up in time, remove the client from the waiting pool
		case <-time.After(constants.INTERLOCUTOR_WAIT_TIME * time.Second):
			s.DeleteClientFromWaitingPool(clientName)
			if err = communication.SendMessage(conn, constants.INTERLOCUTOR_WAIT_TIMEOUT); err != nil {
				log.Println("Couldn't send the message:", err)
			}
			return
		}
	}

	ioReadChannel := make(chan string)
	errorChannel := make(chan error)

	go actions.ReadFromConnection(conn, buffer, ioReadChannel, errorChannel)

	// Here comes the actual chatting!
	for {
		select {
		case interlocutorMessage, ok := <-client.readChannel:
			if !ok {
				log.Println(ErrReadChannelClosed)
				return
			}
			log.Printf("[%s] received message from [%s]: %s", client.name, client.interlocutor, interlocutorMessage)
			if err := communication.SendMessage(conn, interlocutorMessage); err != nil {
				log.Println("Couldn't send the message:", err)
				continue
			}
		case clientMessage := <-ioReadChannel:
			log.Printf("[%s] sent message to [%s]: %s", client.name, client.interlocutor, clientMessage)
			client.writeChannel <- clientMessage
		case err := <-errorChannel:
			if err.Error() == io.EOF.Error() {
				log.Println("Connection closed by client")
				return
			}
		}
	}
}
