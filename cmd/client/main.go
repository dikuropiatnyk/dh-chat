package main

import (
	"log"

	"github.com/dikuropiatnyk/dh-chat/internal/client/types"
)

func main() {
	user := types.DHClient{}
	connection, err := user.Connect()
	if err != nil {
		log.Fatalln("Couldn't connect to the server:", err)
	}
	user.Interact(connection)
}
