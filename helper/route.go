package helper

import (
	"context"
	"encoding/json"
	"log"
	"main/model"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func InitRoute(app *fiber.App) {
	// Route initialization logic goes here
	log.Println("Initializing Routes ...")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Ready to GO!!!!!!!!")
	})

	app.Post("/debug", func(c *fiber.Ctx) error {
		log.Println("POST request received at /debug")
		UserInfo := new(model.UserInfo)
		return c.Status(200).JSON(UserInfo)
	})

	app.Post("/email/webhook", func(c *fiber.Ctx) error {
		log.Println("POST request received at /email/webhook")
		Email := new(model.EmailWebhook)
		UserInfo := new(model.UserInfo)
		var orders []string

		if err := c.BodyParser(Email); err != nil {
			log.Println("POST request received at /email/webhook : Error parsing body -", err.Error())
			return c.Status(400).SendString(err.Error())
		}

		// Extract header and body information
		Email.Header.APIKey = c.Get("X-API-KEY")
		Email.Header.Source = c.Get("X-SOURCE")

		if Email.Header.APIKey != GetEnv("API_KEY") || Email.Header.Source != GetEnv("X_SOURCE") {
			log.Println("POST request received at /email/webhook : Unauthorized access attempt")
			return c.Status(401).SendString("Unauthorized")
		}

		s := strings.Split(strings.ToLower(Email.BodyPlain), "order from ")[1]

		UserInfo.Name = strings.Split(s, ",")[0]
		UserInfo.Email = strings.Split(s, " ")[1]
		if !strings.ContainsAny(strings.Split(s, " ")[2], "items") {
			UserInfo.Phone = strings.Split(s, " ")[2]
		} else {
			UserInfo.Phone = "N/A"
		}

		listOrder := strings.Split(Email.BodyHTML, "</li>")
		orders = append(orders, strings.Split(listOrder[0], `<li style="padding: 0; margin: 0; box-sizing: border-box;">`)[1])

		if len(listOrder) > 2 {
			for n, order := range listOrder {
				if n > 0 && n < len(listOrder)-1 {
					order = strings.Split(order, ">")[1]
					orders = append(orders, order)
				}
			}
		}

		UserInfo.Orders = orders
		println(Email.MessageID, UserInfo.Name, UserInfo.Email, UserInfo.Phone, len(UserInfo.Orders))

		// DB logic
		tx, err := DB.Begin(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		defer tx.Rollback(context.Background()) // Rollback if not committed

		_, err = tx.Exec(context.Background(), "INSERT INTO email (sender, cc, subject, body_plain, body_html, message_id) VALUES ($1, $2, $3, $4, $5, $6)", Email.From, Email.Cc, Email.Subject, Email.BodyPlain, Email.BodyHTML, Email.MessageID)
		if err != nil {
			log.Fatal(err)
		}

		jsonData, err := json.Marshal(orders)
		if err != nil {
			log.Fatal(err)
		}

		_, err = tx.Exec(context.Background(), "INSERT INTO purchase_order (message_id, name, email, phone, orders, trigger_wa) VALUES ($1, $2, $3, $4, $5, $6)", Email.MessageID, UserInfo.Name, UserInfo.Email, UserInfo.Phone, string(jsonData), false)
		if err != nil {
			log.Fatal(err)
		}

		err = tx.Commit(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		return c.Status(200).JSON(UserInfo)
	})
}
