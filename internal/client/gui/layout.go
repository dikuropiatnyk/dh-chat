package gui

import (
	"fmt"
	"net"
	"sync"

	"github.com/dikuropiatnyk/dh-chat/internal/constants"
	"github.com/dikuropiatnyk/dh-chat/pkg/communication"
	"github.com/jroimartin/gocui"
)

func InitLayout(g *gocui.Gui) error {
	// Render two views: one for the chat and one for the input
	maxX, maxY := g.Size()
	chatView, err := g.SetView(constants.CHAT_VIEWNAME, 0, 0, maxX-1, maxY-3)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	chatView.Title = "Chat"
	chatView.Autoscroll = true
	inputView, err := g.SetView(constants.INPUT_VIEWNAME, 0, maxY-3, maxX-1, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	inputView.Title = "Enter your message"
	inputView.Editable = true
	inputView.Wrap = true
	if _, err := g.SetCurrentView(constants.INPUT_VIEWNAME); err != nil {
		return err
	}
	return nil
}

func exit(_ *gocui.Gui, _ *gocui.View, wg *sync.WaitGroup) error {
	defer wg.Done()
	return gocui.ErrQuit
}

func sendMessage(g *gocui.Gui, v *gocui.View, conn net.Conn, clientName string) error {
	// Get the message from the input view
	message := v.Buffer()
	v.Clear()
	if err := v.SetCursor(0, 0); err != nil {
		return err
	}
	// Display the message to the chat view
	chatView, err := g.View(constants.CHAT_VIEWNAME)
	if err != nil {
		return err
	}
	// Display client's name and the message with the specific color
	fmt.Fprintf(chatView, "%s[%s] %s", constants.GREEN_COLOR, clientName, message)

	// Send the message to the server
	if err = communication.SendMessage(conn, message); err != nil {
		return err
	}

	return nil
}

func SetKeyBindings(g *gocui.Gui, conn net.Conn, wg *sync.WaitGroup, clientName string) error {
	// Default keybingding to exit the application

	if err := g.SetKeybinding(
		"",
		gocui.KeyCtrlC,
		gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error { return exit(g, v, wg) }); err != nil {
		return err
	}

	// Keybinding to send the message
	if err := g.SetKeybinding(
		constants.INPUT_VIEWNAME,
		gocui.KeyEnter,
		gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error { return sendMessage(g, v, conn, clientName) }); err != nil {
		return err
	}
	return nil
}

func UpdateChatView(g *gocui.Gui, message string, interlocutorName string) error {
	g.Update(func(g *gocui.Gui) error {
		chatView, err := g.View(constants.CHAT_VIEWNAME)
		if err != nil {
			return err
		}
		fmt.Fprintf(chatView, "%s[%s] %s", constants.RED_COLOR, interlocutorName, message)
		return nil
	})
	return nil
}
