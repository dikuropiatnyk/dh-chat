package communication

import (
	"log"
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
	sent_bytes, err := conn.Write([]byte(message))
	if sent_bytes > 0 {
		log.Printf("Sent %d bytes to %s\n", sent_bytes, conn.RemoteAddr())
	} else if err != nil {
		log.Println("Couldn't send the message:", err)
	}
	return err
}
