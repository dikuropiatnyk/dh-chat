package actions

import (
	"log"
	"net"

	"github.com/dikuropiatnyk/dh-chat/pkg/communication"
)

func ReadFromConnection(conn net.Conn, buffer []byte, output chan<- string, quit chan<- error) {
	for {
		message, err := communication.ReadMessage(conn, buffer)
		if err != nil {
			quit <- err
			return
		}
		output <- message
	}
}

func CloseConnection(conn net.Conn) {
	conn.Close()
	log.Printf("Closed connection with %s\n", conn.RemoteAddr())
}
