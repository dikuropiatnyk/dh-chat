package actions

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/dikuropiatnyk/dh-chat/pkg/communication"
)

func HandleServerResponse(conn net.Conn, buffer []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		serverMessage, err := communication.ReadMessage(conn, buffer)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("\n>> %s\n", serverMessage)
	}
}

func HandleUserResponse(conn net.Conn, reader *bufio.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		userMessage, err := communication.GetInput("", reader)
		if err != nil {
			log.Fatalln(err)
		}
		if err = communication.SendMessage(conn, userMessage); err != nil {
			log.Fatalln(err)
		}
	}
}
