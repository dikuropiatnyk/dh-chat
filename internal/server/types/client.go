package types

import (
	"errors"
	"log"
	"net"

	"github.com/dikuropiatnyk/dh-chat/internal/constants"
	"github.com/dikuropiatnyk/dh-chat/pkg/communication"
)

// New error type for closed read channel
var ErrReadChannelClosed = errors.New("read channel is closed")

type DHClient struct {
	clientAddress net.Addr
	name          string
	interlocutor  string
	readChannel   chan string
	writeChannel  chan string
}

func (c *DHClient) Close() {
	if c.writeChannel != nil {
		close(c.writeChannel)
	}
}

func (c *DHClient) SyncWithInterlocutor(clientConnection net.Conn, buffer []byte) error {
	// Collect the final confirmation from the current client via the clientConnection
	clientConfirmation, err := communication.ReadMessage(clientConnection, buffer)
	if err != nil {
		return err
	}
	log.Printf("Received confirmation from %s!", c.name)
	// Send the client confirmation to the interlocutor
	c.writeChannel <- clientConfirmation

	// Wait for the interlocutor to confirm the chat
	interlocutorConfirmation, ok := <-c.readChannel
	if !ok {
		return ErrReadChannelClosed
	}

	if clientConfirmation != interlocutorConfirmation {
		return errors.New("failed chat confirmation")
	}

	if err = communication.SendMessage(clientConnection, constants.CHAT_CONFIRMED); err != nil {
		return err
	}

	log.Printf("Successful chat synchronization for %s!", c.name)

	return nil
}
