package communication

import (
	"net"
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
