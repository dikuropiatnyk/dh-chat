package actions

import (
	"bufio"
	"log"
	"net"

	"github.com/dikuropiatnyk/dh-chat/internal/constants"
	"github.com/dikuropiatnyk/dh-chat/pkg/communication"
)

// A shakedown of the chat between the user and the interlocutor
func ConfirmChat(userConnection net.Conn, buffer []byte, reader *bufio.Reader) error {
	userConfirmation, err := communication.GetInput("Type the confirmation password: ", reader)
	if err != nil {
		log.Fatalln("Couldn't get the confirmation:", err)
	}

	// Send the confirmation to the user
	if err = communication.SendMessage(userConnection, userConfirmation); err != nil {
		return err
	}
	// Read the confirmation from the interlocutor
	chatConfirmation, err := communication.ReadMessage(userConnection, buffer)
	if err != nil {
		return err
	}
	if chatConfirmation == constants.CHAT_CONFIRMED {
		log.Println("Chat confirmed!")
	} else {
		log.Fatalln("Chat is not confirmed!")
	}
	return nil
}
