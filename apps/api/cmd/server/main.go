package main

import (
	"log"

	"github.com/JosePintos/tora/apps/api/internal/platform/app"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found")
	}
	app := app.New()
	app.Run()
}
