package helper

import "github.com/gofiber/fiber/v2"

func InitRoute(app *fiber.App) {
	// Route initialization logic goes here

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Ready to GO!!!!!!!!")
	})

	app.Post("/email/webhook", func(c *fiber.Ctx) error {
		body := c.Body()
		println(string(body))
		return c.SendStatus(200)
	})
}
