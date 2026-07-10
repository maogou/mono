package main

import (
	"context"
	"log"
	"os"

	"go_template/internal/command"
)

func main() {
	app := command.NewApp()

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
