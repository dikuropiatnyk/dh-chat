package types

import (
	"errors"
	"log"
	"net"
	"time"

	"github.com/dikuropiatnyk/dh-chat/internal/constants"
	"github.com/dikuropiatnyk/dh-chat/pkg/communication"
	"github.com/dikuropiatnyk/dh-chat/pkg/diffiehellman"
)

// New error type for closed read channel
var ErrReadChannelClosed = errors.New("read channel is closed")
var ErrWaitingTimeoutExceeded = errors.New("waiting timeout exceeded")

type DHClient struct {
	clientAddress net.Addr
	name          string
	interlocutor  string
	readChannel   chan string
	writeChannel  chan string
}

func NewDHClient(clientAddress net.Addr, name string, interlocutorName string) *DHClient {
	log.Printf("New client %s connected. Address: %s Interlocutor: %s\n", name, clientAddress.String(), interlocutorName)
	return &DHClient{clientAddress: clientAddress, name: name, interlocutor: interlocutorName}
}

func (c *DHClient) Close() {
	if c.writeChannel != nil {
		close(c.writeChannel)
	}
}

func (c *DHClient) SyncWithInterlocutor(clientConnection net.Conn, buffer []byte) error {
	// Collect the public salt from the current client via the clientConnection
	clientPublicSalt, err := communication.ReadMessage(clientConnection, buffer)
	if err != nil {
		return err
	}
	log.Printf("Received a public salt from %s!\n%s", c.name, clientPublicSalt)
	// Send the client confirmation to the interlocutor
	c.writeChannel <- clientPublicSalt

	// Wait for the interlocutor to provide the public salt
	interlocutorPublicSalt, ok := <-c.readChannel
	if !ok {
		return ErrReadChannelClosed
	}

	sharedMessage := constants.CHAT_CONFIRMED + constants.DATA_SEPARATOR + interlocutorPublicSalt
	if err = communication.SendMessage(clientConnection, sharedMessage); err != nil {
		return err
	}

	log.Printf("Successful chat synchronization for %s!", c.name)

	return nil
}

func (c *DHClient) HandleFirstClient(conn net.Conn, buffer []byte) error {
	if err := communication.SendMessage(conn, constants.NO_INTERLOCUTOR); err != nil {
		return err
	}
	// Set up a blocking waiter until the interlocutor is found, which is unblocked
	// by the interlocutor goroutine
	select {
	case chatSecrets, ok := <-c.readChannel:
		if !ok {
			return ErrReadChannelClosed
		}
		if err := communication.SendMessage(conn, chatSecrets); err != nil {
			return err
		}
		if err := c.SyncWithInterlocutor(conn, buffer); err != nil {
			return err
		}
	// If the interlocutor doesn't show up in time, remove the client from the waiting pool
	case <-time.After(constants.INTERLOCUTOR_WAIT_TIME * time.Second):
		if err := communication.SendMessage(conn, constants.INTERLOCUTOR_WAIT_TIMEOUT); err != nil {
			return err
		}
		return ErrWaitingTimeoutExceeded
	}

	return nil
}

// By second client it's implied the client that is connected after its
// interlocutor, and so it's the moment to exchange the base secrets
// and start the chat
func (c *DHClient) HandleSecondClient(conn net.Conn, buffer []byte) error {
	p, g, err := diffiehellman.GenerateBaseSecrets()
	if err != nil {
		return err
	}

	log.Printf("Generated base secrets for chat %s <=> %s!\np=%s, g=%s\n", c.name, c.interlocutor, p.String(), g.String())

	// Prepare the message with base secrets to send to both clients
	sharedMessage := constants.INTERLOCUTOR_FOUND + constants.DATA_SEPARATOR + p.String() + constants.DATA_SEPARATOR + g.String()
	// Send the message to the current client
	if err = communication.SendMessage(conn, sharedMessage); err != nil {
		return err
	}
	// Send the message to the interlocutor via the write channel
	c.writeChannel <- sharedMessage

	// Synchronize the chat between the current client and the interlocutor
	if err = c.SyncWithInterlocutor(conn, buffer); err != nil {
		log.Println("Chat synchronization error:", err)
		return err
	}
	return nil
}
