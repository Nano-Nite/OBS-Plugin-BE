package main

import (
	"log"
	"main/db"
	"os"

	"github.com/gofiber/fiber/v2"
)

func main() {
	db.InitDB()

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Ready to GO!!!!!!!!")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	log.Println("Listening on port", port)
	app.Listen(":" + port)
}
