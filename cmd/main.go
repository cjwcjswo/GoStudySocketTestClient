package main

import (
	"chatClient/pkg/app"
)

func main() {
	cli := app.NewChatClientApp()
	cli.Start()
}
