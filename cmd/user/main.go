package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

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

func getInput(prompt string, reader *bufio.Reader) (string, error) {
	fmt.Print(prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.Trim(input, "\n"), nil
}

func (u *DHUser) SendData(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)

	// Read user input and send it to the server
	userName, err := getInput("Enter your name: ", reader)
	if err != nil {
		log.Fatalln("Couldn't read your name:", err)
	}
	interlocutorName, err := getInput("Enter interlocutor's name: ", reader)
	if err != nil {
		log.Fatalln("Couldn't read the interlocutor's name:", err)
	}
	// Concatenate the user name and the interlocutor's name
	communication.SendMessage(conn, userName+":"+interlocutorName)

	buffer := make([]byte, 2048)
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
		}
	} else if serverResponse == constants.INTERLOCUTOR_FOUND {
		log.Println("Interlocutor found! Start chatting...")
	}
}

func main() {
	user := DHUser{}
	user.Connect()
}
