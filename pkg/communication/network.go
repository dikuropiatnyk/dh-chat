package communication

import (
	"net"

	"github.com/dikuropiatnyk/dh-chat/pkg/crypt"
)

func ReadMessage(conn net.Conn, buffer []byte) (string, error) {
	n, err := conn.Read(buffer)
	if err != nil {
		return "", err
	}
	return string(buffer[:n]), nil
}

func SendMessage(conn net.Conn, message string) error {
	_, err := conn.Write([]byte(message))
	return err
}

func ReadEncryptedMessage(conn net.Conn, buffer []byte, key []byte) (string, error) {
	encryptedMessage, err := ReadMessage(conn, buffer)
	if err != nil {
		return "", err
	}

	return crypt.DecryptMessage(encryptedMessage, key)
}

func SendEncryptedMessage(conn net.Conn, message string, key []byte) error {
	encryptedMessage, err := crypt.EncryptMessage(message, key)
	if err != nil {
		return err
	}

	return SendMessage(conn, encryptedMessage)
}
