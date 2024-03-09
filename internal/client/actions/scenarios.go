package actions

import (
	"bufio"
	"errors"
	"log"
	"math/big"
	"net"
	"strings"

	"github.com/dikuropiatnyk/dh-chat/internal/constants"
	"github.com/dikuropiatnyk/dh-chat/pkg/communication"
	"github.com/dikuropiatnyk/dh-chat/pkg/crypt"
	"github.com/dikuropiatnyk/dh-chat/pkg/diffiehellman"
)

var ErrStringToBigInt = errors.New("couldn't convert the string to a big integer")

// A shakedown of the chat between the user and the interlocutor
func Shakedown(userConnection net.Conn, buffer []byte, reader *bufio.Reader, sharedMessage string) ([]byte, error) {
	// Split the client data into the client name and the interlocutor
	// The client data is in the format "clientName;interlocutor"
	sharedMessageSlice := strings.Split(sharedMessage, constants.DATA_SEPARATOR)
	if len(sharedMessageSlice) != 3 {
		return nil, errors.New("invalid shared message")
	}
	// The shared message is in the format "signal:p:g"
	pStr, gStr := sharedMessageSlice[1], sharedMessageSlice[2]

	// Convert the public secrets to big integers
	p, success := new(big.Int).SetString(pStr, 10)
	if !success {
		return nil, ErrStringToBigInt
	}
	g, success := new(big.Int).SetString(gStr, 10)
	if !success {
		return nil, ErrStringToBigInt
	}

	// Generate a private salt
	privateSalt, err := diffiehellman.GeneratePrivateSalt(p)
	if err != nil {
		return nil, err
	}
	log.Println("Generated private salt. Don't show it to no one!")
	log.Println(privateSalt)

	// Generate a public salt
	publicSalt := diffiehellman.GeneratePublicSalt(p, g, privateSalt)

	// Send the public salt to the user
	if err = communication.SendMessage(userConnection, publicSalt.String()); err != nil {
		return nil, err
	}
	// Read the public salt from the interlocutor
	chatConfirmation, err := communication.ReadMessage(userConnection, buffer)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(chatConfirmation, constants.CHAT_CONFIRMED) {
		return nil, errors.New("chat confirmation failed")
	}
	chatConfirmationSlice := strings.Split(chatConfirmation, constants.DATA_SEPARATOR)
	if len(chatConfirmationSlice) != 2 {
		return nil, errors.New("invalid chat confirmation")
	}
	interlocutorPublicSaltStr := chatConfirmationSlice[1]
	interlocutorPublicSalt, success := new(big.Int).SetString(interlocutorPublicSaltStr, 10)
	if !success {
		return nil, ErrStringToBigInt
	}

	// Generate the symmetric key
	symmetricKey := diffiehellman.GenerateSymmetricKey(p, interlocutorPublicSalt, privateSalt)
	// Derive the key
	derivedKey, err := crypt.DeriveKey(symmetricKey)
	if err != nil {
		return nil, err
	}
	log.Println("Derived key:", derivedKey)

	return derivedKey, nil
}
