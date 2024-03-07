package types

import (
	"bufio"
	"log"
	"net"
	"os"
	"sync"

	"github.com/dikuropiatnyk/dh-chat/internal/client/actions"
	"github.com/dikuropiatnyk/dh-chat/internal/constants"
	"github.com/dikuropiatnyk/dh-chat/pkg/communication"
)

type DHClient struct {
	clientAddress net.Addr
	serverAddress net.Addr
}

func (c *DHClient) Connect() (net.Conn, error) {
	conn, err := net.Dial(constants.SERVER_CONNECTION_TYPE, constants.SERVER_ADDRESS)
	if err != nil {
		return nil, err
	}
	c.clientAddress = conn.LocalAddr()
	c.serverAddress = conn.RemoteAddr()
	log.Println("Connected to", c.serverAddress)
	return conn, nil
}

// Main function, where client makes all interactions with the server via an established connection
func (c *DHClient) Interact(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	// Read user input and send it to the server
	userName, err := communication.GetInput("Enter your name: ", reader)
	if err != nil {
		log.Fatalln("Couldn't read the name:", err)
	}
	interlocutorName, err := communication.GetInput("Enter interlocutor's name: ", reader)
	if err != nil {
		log.Fatalln("Couldn't read the interlocutor's name:", err)
	}
	// Concatenate the user name and the interlocutor's name
	if err = communication.SendMessage(conn, userName+constants.DATA_SEPARATOR+interlocutorName); err != nil {
		return
	}

	buffer := make([]byte, constants.BUFFER_SIZE)
	// First reading from the connection to get the user name and the interlocutor
	serverResponse, err := communication.ReadMessage(conn, buffer)
	if err != nil {
		log.Fatalln("Couldn't get a user info:", err)
	}
	if serverResponse == constants.NO_INTERLOCUTOR {
		log.Println("No interlocutor found! Wait, please...")
		serverUpdate, err := communication.ReadMessage(conn, buffer)
		if err != nil {
			log.Fatalln("Couldn't get a user info:", err)
		}
		if serverUpdate == constants.INTERLOCUTOR_FOUND {
			log.Println("Interlocutor found! Start chatting...")
			actions.ConfirmChat(conn, buffer, reader)
		} else if serverUpdate == constants.INTERLOCUTOR_WAIT_TIMEOUT {
			log.Println("No interlocutor found! Try again later...")
			return
		}
	} else if serverResponse == constants.INTERLOCUTOR_FOUND {
		log.Println("Interlocutor found! Start chatting...")
		actions.ConfirmChat(conn, buffer, reader)
	}
	log.Println("Let the chat begin!")

	var wg sync.WaitGroup
	wg.Add(2)
	go actions.HandleServerResponse(conn, buffer, &wg)
	go actions.HandleUserResponse(conn, reader, &wg)
	wg.Wait()
}
