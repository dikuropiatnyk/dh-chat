package types

import (
	"bufio"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/dikuropiatnyk/dh-chat/internal/client/actions"
	"github.com/dikuropiatnyk/dh-chat/internal/client/gui"
	"github.com/dikuropiatnyk/dh-chat/internal/constants"
	"github.com/dikuropiatnyk/dh-chat/pkg/communication"
	"github.com/jroimartin/gocui"
)

type DHClient struct {
	clientAddress net.Addr
	serverAddress net.Addr
	key           []byte
}

func (c *DHClient) Connect() (net.Conn, error) {
	conn, err := net.Dial(constants.SERVER_CONNECTION_TYPE, constants.SERVER_ADDRESS)
	if err != nil {
		return nil, err
	}
	c.clientAddress = conn.LocalAddr()
	c.serverAddress = conn.RemoteAddr()
	log.Println("Connected to", c.serverAddress)
	return conn, nil
}

// Main function, where client makes all interactions with the server via an established connection
func (c *DHClient) Interact(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	// Read user input and send it to the server
	clientName, err := communication.GetInput("Enter your name: ", reader)
	if err != nil {
		log.Fatalln("Couldn't read the name:", err)
	}
	interlocutorName, err := communication.GetInput("Enter interlocutor's name: ", reader)
	if err != nil {
		log.Fatalln("Couldn't read the interlocutor's name:", err)
	}
	// Concatenate the user name and the interlocutor's name
	if err = communication.SendMessage(conn, clientName+constants.DATA_SEPARATOR+interlocutorName); err != nil {
		log.Fatalln("Couldn't send the user info:", err)
	}

	buffer := make([]byte, constants.BUFFER_SIZE)
	// First reading from the connection to get the user name and the interlocutor
	serverResponse, err := communication.ReadMessage(conn, buffer)
	if err != nil {
		log.Fatalln("Couldn't get a user info:", err)
	}

	switch {
	case strings.HasPrefix(serverResponse, constants.CLIENT_EXISTS):
		log.Fatalln("Client already exists! Exiting...")

	case strings.HasPrefix(serverResponse, constants.INTERLOCUTOR_FOUND):
		log.Println("Interlocutor found! Start chatting...")
		derivedKey, err := actions.Shakedown(conn, buffer, reader, serverResponse)
		if err != nil {
			log.Fatalln("Couldn't shake hands with the interlocutor:", err)
		}
		c.key = derivedKey

	case strings.HasPrefix(serverResponse, constants.NO_INTERLOCUTOR):
		log.Println("No interlocutor found! Wait, please...")
		serverUpdate, err := communication.ReadMessage(conn, buffer)
		if err != nil {
			log.Fatalln("Couldn't get a user info:", err)
		}

		switch {
		case strings.HasPrefix(serverUpdate, constants.INTERLOCUTOR_FOUND):
			log.Println("Interlocutor found! Start chatting...")
			derivedKey, err := actions.Shakedown(conn, buffer, reader, serverUpdate)
			if err != nil {
				log.Fatalln("Couldn't shake hands with the interlocutor:", err)
			}
			c.key = derivedKey
		case strings.HasPrefix(serverUpdate, constants.INTERLOCUTOR_WAIT_TIMEOUT):
			log.Println("Interlocutor didn't show up! Exiting...")
			return

		default:
			log.Fatalln("Unknown server response! Exiting...")
		}
	}

	log.Println("Let the chat begin!")

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Fatalln(err)
	}
	defer g.Close()
	g.Cursor = true

	g.SetManagerFunc(gui.InitLayout)

	var wg sync.WaitGroup
	wg.Add(1)
	// Set the keybindings
	if err = gui.SetKeyBindings(g, conn, &wg, clientName, c.key); err != nil {
		log.Fatalln(err)
	}

	go actions.HandleServerResponse(conn, buffer, g, interlocutorName, c.key)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Fatalln(err)
	}
	wg.Wait()
}
