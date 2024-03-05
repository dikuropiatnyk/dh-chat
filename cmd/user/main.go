package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/dikuropiatnyk/dh-chat/internal/constants"
	"github.com/dikuropiatnyk/dh-chat/pkg/communication"
)

type DHUser struct {
	userAddress   net.Addr
	serverAddress net.Addr
}

func (u *DHUser) Connect() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatalln("Connection error:", err)
		return
	}
	defer conn.Close()
	u.userAddress = conn.LocalAddr()
	u.serverAddress = conn.RemoteAddr()
	log.Println("Connected to", u.serverAddress)
	u.SendData(conn)
}

func confirmChat(userConnection net.Conn, buffer []byte, reader *bufio.Reader) error {
	userConfirmation, err := communication.GetInput("Type the confirmation password: ", reader)
	if err != nil {
		log.Fatalln("Couldn't get the confirmation:", err)
	}

	// Send the confirmation to the user
	err = communication.SendMessage(userConnection, userConfirmation)
	if err != nil {
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

func handleServerResponse(conn net.Conn, buffer []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		serverMessage, err := communication.ReadMessage(conn, buffer)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("[CHAT] => %s\n", serverMessage)
	}
}

func handleUserResponse(conn net.Conn, reader *bufio.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		userMessage, err := communication.GetInput("[CHAT]: ", reader)
		if err != nil {
			log.Fatalln(err)
		}
		communication.SendMessage(conn, userMessage)
	}
}

func (u *DHUser) SendData(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)

	// Read user input and send it to the server
	userName, err := communication.GetInput("Enter your name: ", reader)
	if err != nil {
		log.Fatalln("Couldn't read the name:", err)
	}
	interlocutorName, err := communication.GetInput("Enter interlocutor's name: ", reader)
	if err != nil {
		log.Fatalln("Couldn't read the interlocutor's name:", err)
	}
	// Concatenate the user name and the interlocutor's name
	communication.SendMessage(conn, userName+constants.DATA_SEPARATOR+interlocutorName)

	buffer := make([]byte, constants.BUFFER_SIZE)
	// First reading from the connection to get the user name and the interlocutor
	serverResponse, err := communication.ReadMessage(conn, buffer)
	if err != nil {
		log.Fatalln("Couldn't get a user info:", err)
	}
	if serverResponse == constants.NO_INTERLOCUTOR {
		log.Println("No interlocutor found! Wait, please...")
		serverUpdate, err := communication.ReadMessage(conn, buffer)
		if err != nil {
			log.Fatalln("Couldn't get a user info:", err)
		}
		if serverUpdate == constants.INTERLOCUTOR_FOUND {
			log.Println("Interlocutor found! Start chatting...")
			confirmChat(conn, buffer, reader)
		}
	} else if serverResponse == constants.INTERLOCUTOR_FOUND {
		log.Println("Interlocutor found! Start chatting...")
		confirmChat(conn, buffer, reader)
	}
	log.Println("Let the chat begin!")

	var wg sync.WaitGroup
	wg.Add(2)
	go handleServerResponse(conn, buffer, &wg)
	go handleUserResponse(conn, reader, &wg)
	wg.Wait()
}

func main() {
	user := DHUser{}
	user.Connect()
}
