package helper

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"main/model"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
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

		headers := make(map[string]string)
		c.Context().Request.Header.VisitAll(func(key, value []byte) {
			headers[string(key)] = string(value)
		})

		for k, v := range c.GetReqHeaders() {
			log.Printf("%s: %s", k, v)
		}
		log.Println(string(c.Body()))

		UserInfo := new(model.UserInfo)
		return c.Status(200).JSON(UserInfo)
	})

	app.Post("/login", func(c *fiber.Ctx) error {
		log.Println("POST request received at /login")

		if os.Getenv("ENV") != "development" && c.Get("postman-token") != "" {
			log.Println("POST request received at /login : Unauthorized access attempt")
			return c.Status(401).SendString("Unauthorized")
		}

		if c.Get("x-device-id") == "" || c.Get("x-signature") == "" {
			log.Println("POST request received at /login : missing device id or signature")
			log.Println("Device ID : " + c.Get("x-device-id"))
			log.Println("Signature : " + c.Get("x-signature"))

			return c.Status(401).SendString("Unauthorized")
		}
		HeaderLogin := new(model.HeaderLogin)
		HeaderLogin.XDeviceID = c.Get("x-device-id")
		HeaderLogin.XSignature = c.Get("x-signature")

		LoginPayload := new(model.LoginPayload)
		LoginPayload.Header = *HeaderLogin

		log.Println("Body : " + string(c.Body()))
		if err := c.BodyParser(LoginPayload); err != nil {
			log.Println("POST request received at /login : Error parsing body -", err.Error())
			return c.Status(400).SendString(err.Error())
		}

		sDecode, err := base64.StdEncoding.DecodeString(HeaderLogin.XSignature)
		if err != nil {
			log.Println("POST request received at /login : Error decoding signature -", err.Error())
			return c.Status(400).SendString(err.Error())
		}

		splitSignature := strings.Split(string(sDecode), "|")
		if len(splitSignature) != 3 || splitSignature[1] != LoginPayload.Email {
			log.Println("POST request received at /login : Invalid signature")
			return c.Status(401).SendString("Unauthorized")
		}

		// prevent user not buying the product to login
		var purchaseOrder []*model.PurchaseOrder
		log.Printf("Searching for purchase order with email: %s", LoginPayload.Email)
		err = pgxscan.Select(c.Context(), DB, &purchaseOrder, "SELECT * FROM purchase_order WHERE email=$1", LoginPayload.Email)
		if err != nil {
			log.Println("POST request received at /login : Error fetching purchase order -", err.Error())
			return c.Status(500).SendString("Internal Server Error")
		}
		if len(purchaseOrder) == 0 {
			log.Println("POST request received at /login : No purchase order found for email:", LoginPayload.Email)
			return c.Status(401).SendString("Unauthorized")
		}

		// get login log
		var loginLog []*model.LoginLog
		log.Printf("Searching for login log with email: %s", LoginPayload.Email)
		err = pgxscan.Select(c.Context(), DB, &loginLog, "SELECT * FROM login_log WHERE email=$1", LoginPayload.Email)
		if err != nil {
			log.Println("POST request received at /login : Error fetching login log -", err.Error())
			return c.Status(500).SendString("Internal Server Error")
		}
		if len(loginLog) == 0 {
			// First time login, insert new record
			_, err := DB.Exec(c.Context(), "INSERT INTO login_log (email, signature, device_id, failed_attempt, last_login, created_at) VALUES ($1, $2, $3, $4, $5, $6)", LoginPayload.Email, HeaderLogin.XSignature, HeaderLogin.XDeviceID, 0, getCurrentTime(), getCurrentTime())
			if err != nil {
				log.Println("POST request received at /login : Error inserting login log -", err.Error())
				return c.Status(500).SendString("Internal Server Error")
			}
		}

		if len(loginLog) >= 1 {
			if loginLog[len(loginLog)-1].Signature != HeaderLogin.XSignature || loginLog[len(loginLog)-1].DeviceID != HeaderLogin.XDeviceID {
				// Invalid login attempt
				_, err := DB.Exec(c.Context(), "UPDATE login_log SET failed_attempt=failed_attempt+1 WHERE id=$1", loginLog[len(loginLog)-1].ID)
				if err != nil {
					log.Println("POST request received at /login : Error updating failed attempt -", err.Error())
					return c.Status(500).SendString("Internal Server Error")
				}

				_, err = DB.Exec(c.Context(), "INSERT INTO login_log (email, signature, device_id, failed_attempt, last_login, created_at) VALUES ($1, $2, $3, $4, $5, $6)", LoginPayload.Email, HeaderLogin.XSignature, HeaderLogin.XDeviceID, 0, getCurrentTime(), getCurrentTime())
				if err != nil {
					log.Println("POST request received at /login : Error inserting login log -", err.Error())
					return c.Status(500).SendString("Internal Server Error")
				}
			} else {
				// Second time login, update new record
				_, err := DB.Exec(c.Context(), "UPDATE login_log SET last_login=$1 WHERE id=$2", getCurrentTime(), loginLog[len(loginLog)-1].ID)
				if err != nil {
					log.Println("POST request received at /login : Error updating login log -", err.Error())
					return c.Status(500).SendString("Internal Server Error")
				}
			}
		}

		// fetch product
		var products []*model.Product
		log.Printf("Searching for product with email: %s", LoginPayload.Email)
		q := `SELECT distinct on (item) 
				split_part(item.value, ' - ', 1) AS item,
				pr.url
				FROM purchase_order p
				JOIN LATERAL json_array_elements_text(p.orders) AS item(value)
				ON TRUE
				left join product pr on lower(pr.code) = lower(split_part(item.value, ' - ', 1))
				where p.email = $1`

		err = pgxscan.Select(c.Context(), DB, &products, q, LoginPayload.Email)
		if err != nil {
			log.Println("POST request received at /login : Error fetching products -", err.Error())
			return c.Status(500).SendString("Internal Server Error")
		}

		result := make(map[string]interface{})
		result["email"] = LoginPayload.Email
		result["products"] = products

		return c.Status(200).JSON(result)
	})

	// will be deprecated
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

		// Insert email record
		_, err = tx.Exec(context.Background(), "INSERT INTO email (sender, cc, subject, body_plain, body_html, message_id) VALUES ($1, $2, $3, $4, $5, $6)", Email.From, Email.Cc, Email.Subject, Email.BodyPlain, Email.BodyHTML, Email.MessageID)
		if err != nil {
			log.Fatal(err)
		}

		// Insert purchase order record
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

	app.Post("/webhook", func(c *fiber.Ctx) error {
		log.Println("POST request received at /webhook")

		headers := make(map[string]string)
		c.Context().Request.Header.VisitAll(func(key, value []byte) {
			headers[string(key)] = string(value)
		})

		for k, v := range c.GetReqHeaders() {
			log.Printf("%s: %s", k, v)
		}
		log.Println(string(c.Body()))

		log.Println("POST request received at /webhook")
		WebhookPayload := new(model.WebhookPayload)
		if err := c.BodyParser(WebhookPayload); err != nil {
			log.Println("POST request received at /webhook : Error parsing body -", err.Error())
			return c.Status(400).SendString(err.Error())
		}

		// validation
		if c.Get("X-Lynk-Signature") == "" {
			log.Println("POST request received at /webhook : Unauthorized access attempt")
			return c.Status(401).SendString("Unauthorized")
		}
		if !ValidateLynkSignature(WebhookPayload.Data.MessageData.RefID, strconv.Itoa(WebhookPayload.Data.MessageData.Totals.GrandTotal), WebhookPayload.Data.MessageID, c.Get("X-Lynk-Signature"), os.Getenv("LYNK_SIG")) {
			log.Println("POST request received at /webhook : Unauthorized access attempt")
			return c.Status(401).SendString("Unauthorized")
		}

		// DB logic
		tx, err := DB.Begin(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err != nil {
				_ = tx.Rollback(ctx)
			}
		}()

		User := new(model.User)
		User.Name = WebhookPayload.Data.MessageData.Customer.Name
		User.Email = WebhookPayload.Data.MessageData.Customer.Email
		User.Phone = WebhookPayload.Data.MessageData.Customer.Phone
		User.IsTrial = false
		User.LastPurchase = getCurrentTime()
		User.CreatedAt = getCurrentTime()

		err = tx.QueryRow(
			context.Background(),
			`
			INSERT INTO "users" (name, email, phone, is_trial, trial_until, last_purchase, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (email) 
			DO UPDATE SET
				last_purchase = EXCLUDED.last_purchase
			RETURNING id
			`,
			User.Name,
			User.Email,
			User.Phone,
			User.IsTrial,
			nil,
			User.LastPurchase,
			User.CreatedAt,
		).Scan(&User.ID)

		if err != nil {
			_ = tx.Rollback(ctx)
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		}

		// insert purchase order based on product data
		for _, v := range WebhookPayload.Data.MessageData.Items {
			log.Println(v)

			var Product []*model.Product
			log.Printf("Searching for product: %s", v.Title)
			err = pgxscan.Select(c.Context(), DB, &Product, "SELECT * FROM product WHERE name=$1 limit 1", v.Title)
			if err != nil {
				log.Println("POST request received at /webhook : Error get data Product - ", err.Error())
				_ = tx.Rollback(ctx)
				return c.Status(500).JSON(fiber.Map{
					"success": false,
					"error":   err.Error(),
				})
			}
			if len(Product) == 0 {
				log.Println("POST request received at /webhook : Product not found - ", err.Error())
				_ = tx.Rollback(ctx)
				return c.Status(500).JSON(fiber.Map{
					"success": false,
					"error":   err.Error(),
				})
			}

			PurchaseOrder := new(model.PurchaseOrder)
			PurchaseOrder.UserID = User.ID
			PurchaseOrder.ProductID = Product[0].ID
			PurchaseOrder.MessageID = WebhookPayload.Data.MessageID
			PurchaseOrder.TriggerWA = false
			PurchaseOrder.CreatedAt = getCurrentTime()

			_, err = tx.Exec(context.Background(), "INSERT INTO purchase_order (user_id, product_id, message_id, trigger_wa, created_at) VALUES ($1, $2, $3, $4, $5)", PurchaseOrder.UserID, PurchaseOrder.ProductID, PurchaseOrder.MessageID, PurchaseOrder.TriggerWA, PurchaseOrder.CreatedAt)
			if err != nil {
				log.Println("POST request received at /webhook : Fail insert Purchase Order - ", err.Error())
				_ = tx.Rollback(ctx)
				return c.Status(500).JSON(fiber.Map{
					"success": false,
					"error":   err.Error(),
				})
			}

			fmt.Printf("%s", AddDaysFromNextMidnight(time.Now(), Product[0].AddedDuration))
			_, err = tx.Exec(context.Background(), "UPDATE users SET subs_until=$1 WHERE id=$2", AddDaysFromNextMidnight(time.Now(), Product[0].AddedDuration), User.ID)
			if err != nil {
				log.Println("POST request received at /webhook : Failed to Update subs_until - ", err.Error())
				_ = tx.Rollback(ctx)
				return c.Status(500).JSON(fiber.Map{
					"success": false,
					"error":   err.Error(),
				})
			}
		}

		err = tx.Commit(context.Background())
		if err != nil {
			return fmt.Errorf("commit tx: %w", err)
		}

		return c.Status(200).JSON(User)
	})
}
