package actions

import (
	"io"
	"log"
	"net"
	"os"

	"github.com/dikuropiatnyk/dh-chat/internal/client/gui"
	"github.com/dikuropiatnyk/dh-chat/pkg/communication"
	"github.com/jroimartin/gocui"
)

func HandleServerResponse(conn net.Conn, buffer []byte, renderedGUI *gocui.Gui, interlocutorName string, clientKey []byte) {
	for {
		serverMessage, err := communication.ReadEncryptedMessage(conn, buffer, clientKey)
		if err != nil {
			if err.Error() == io.EOF.Error() {
				renderedGUI.Close()
				log.Println("Connection closed by the server, see ya!")
				os.Exit(0)
			} else {
				log.Fatalln("Couldn't read the message: ", err)
			}
		}
		if err = gui.UpdateChatView(renderedGUI, serverMessage, interlocutorName); err != nil {
			log.Fatalln(err)
		}
	}
}
