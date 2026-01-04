package main

import (
	"log"
	"main/helper"
	"os"

	"github.com/gofiber/fiber/v2"
)

func main() {
	helper.InitDB()

	app := fiber.New()
	helper.InitRoute(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	log.Println("Listening on port", port)
	app.Listen(":" + port)
}
