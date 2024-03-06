package types

import (
	"errors"
	"log"
	"net"

	"github.com/dikuropiatnyk/dh-chat/internal/constants"
	"github.com/dikuropiatnyk/dh-chat/pkg/communication"
)

type DHClient struct {
	userAddress  net.Addr
	name         string
	interlocutor string
	readChannel  chan string
	writeChannel chan string
}

func (u *DHClient) syncWithInterlocutor(userConnection net.Conn, buffer []byte) error {
	// Collect the final confirmation from the current user via the userConnection
	userConfirmation, err := communication.ReadMessage(userConnection, buffer)
	if err != nil {
		return err
	}
	log.Printf("Received confirmation from %s!", u.name)
	// Send the user confirmation to the interlocutor
	u.writeChannel <- userConfirmation
	log.Printf("Wrote to channel %s!", u.name)

	// Wait for the interlocutor to confirm the chat
	interlocutorConfirmation := <-u.readChannel

	if userConfirmation != interlocutorConfirmation {
		return errors.New("failed chat confirmation")
	}

	err = communication.SendMessage(userConnection, constants.CHAT_CONFIRMED)
	if err != nil {
		return err
	}

	log.Printf("Successful chat synchronization for %s!", u.name)

	return nil
}
