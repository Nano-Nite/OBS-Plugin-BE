package main

import (
	"fmt"
	"log"
	"main/helper"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	InitENV()
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

func InitENV() {
	// Environment variable initialization logic goes here
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	for _, env := range os.Environ() {
		envPair := strings.SplitN(env, "=", 2)
		key := envPair[0]
		value := envPair[1]
		fmt.Printf("%s : %s\n", key, value)
	}
}
