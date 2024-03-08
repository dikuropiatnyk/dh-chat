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

func HandleServerResponse(conn net.Conn, buffer []byte, renderedGUI *gocui.Gui, interlocutorName string) {
	for {
		serverMessage, err := communication.ReadMessage(conn, buffer)
		if err != nil {
			if err.Error() == io.EOF.Error() {
				renderedGUI.Close()
				log.Println("Connection closed by the server, see ya!")
				os.Exit(0)
			}
		}
		if err = gui.UpdateChatView(renderedGUI, serverMessage, interlocutorName); err != nil {
			log.Fatalln(err)
		}
	}
}
