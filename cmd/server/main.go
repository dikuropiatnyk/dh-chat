package main

import "github.com/dikuropiatnyk/dh-chat/internal/server/types"

func main() {
	server := types.NewDHServer()
	server.Start()
}
