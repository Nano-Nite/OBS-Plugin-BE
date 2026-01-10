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
	"github.com/google/uuid"
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

	app.Post("/trial", func(c *fiber.Ctx) error {
		log.Println("POST request received at /trial")
		result := make(map[string]interface{})
		result["status_code"] = "200"
		result["message"] = ""
		result["data"] = nil

		LoginPayload := new(model.LoginPayload)
		log.Println("Body : " + string(c.Body()))
		if err := c.BodyParser(LoginPayload); err != nil {
			log.Println("POST request received at /trial : Error parsing body -", err.Error())
			return ReturnResult(c, result, 500, "Internal server error", nil, false, nil, nil, nil)
		}

		var Users []*model.User
		log.Printf("Searching for User with email: %s", LoginPayload.Email)
		err := pgxscan.Select(c.Context(), DB, &Users, "SELECT * FROM users WHERE email=$1", LoginPayload.Email)
		if err != nil {
			log.Println("POST request received at /trial : Error fetching Users -", err.Error())
			return ReturnResult(c, result, 500, "Internal server error", nil, false, nil, nil, nil)
		}
		if len(Users) == 0 {
			log.Println("POST request received at /trial : No User for email:", LoginPayload.Email)
			return ReturnResult(c, result, 400, "Bad Request", nil, false, nil, nil, nil)
		}
		if !Users[0].IsTrial {
			log.Println("POST request received at /trial : Not Trial User:", LoginPayload.Email)
			return ReturnResult(c, result, 401, "Unauthorized", nil, false, &Users[0].ID, nil, nil)
		}
		if Users[0].IsTrial {
			if Users[0].TrialUntil == nil {
				log.Println("POST request received at /trial : Not Trial User:", LoginPayload.Email)
				return ReturnResult(c, result, 401, "Unauthorized", nil, false, &Users[0].ID, nil, nil)
			}
			if Users[0].TrialUntil.Before(getCurrentTime()) {
				log.Println("POST request received at /trial : Out of Trial Session:", LoginPayload.Email)
				return ReturnResult(c, result, 401, "Unauthorized", nil, false, &Users[0].ID, nil, nil)
			}
		}

		if os.Getenv("ENV") != "development" && c.Get("postman-token") != "" {
			log.Println("POST request received at /trial : Unauthorized access attempt")
			return ReturnResult(c, result, 401, "Unauthorized", nil, false, &Users[0].ID, nil, nil)
		}

		if c.Get("x-device-id") == "" || c.Get("x-signature") == "" {
			log.Println("POST request received at /trial : missing device id or signature")
			return ReturnResult(c, result, 401, "Unauthorized", nil, false, &Users[0].ID, nil, nil)
		}

		HeaderLogin := new(model.HeaderLogin)
		HeaderLogin.XDeviceID = c.Get("x-device-id")
		HeaderLogin.XSignature = c.Get("x-signature")
		LoginPayload.Header = *HeaderLogin
		sDecode, err := base64.StdEncoding.DecodeString(HeaderLogin.XSignature)
		if err != nil {
			log.Println("POST request received at /trial : Error decoding signature -", err.Error())
			return ReturnResult(c, result, 500, "Internal server error", nil, false, &Users[0].ID, &HeaderLogin.XSignature, &HeaderLogin.XDeviceID)
		}

		splitSignature := strings.Split(string(sDecode), "|")
		if len(splitSignature) != 3 || splitSignature[1] != LoginPayload.Email {
			log.Println("POST request received at /trial : Invalid signature")
			return ReturnResult(c, result, 401, "Unathorized", nil, false, &Users[0].ID, &HeaderLogin.XSignature, &HeaderLogin.XDeviceID)
		}

		// prevent user not buying the product to login
		if Users[0].SubsUntil.Before(getCurrentTime()) {
			log.Println("POST request received at /trial : User out of subscription: ", Users[0].SubsUntil.Format(time.RFC3339))
			return ReturnResult(c, result, 402, "Payment Required", nil, false, &Users[0].ID, &HeaderLogin.XSignature, &HeaderLogin.XDeviceID)
		}

		// fetch product
		var products []*model.Product
		log.Println("Searching for products")
		q := `SELECT * FROM product WHERE (owned_by is null and (url is not null or url != '')) or owned_by = $1 order by owned_by, code asc`

		err = pgxscan.Select(c.Context(), DB, &products, q)
		if err != nil {
			log.Println("POST request received at /trial : Error fetching products -", err.Error())
			return ReturnResult(c, result, 500, "Internal server error", nil, false, &Users[0].ID, &HeaderLogin.XSignature, &HeaderLogin.XDeviceID)
		}
		if len(products) == 0 {
			log.Println("POST request received at /trial : No products")
			return ReturnResult(c, result, 500, "Internal server error", nil, false, &Users[0].ID, &HeaderLogin.XSignature, &HeaderLogin.XDeviceID)
		}

		fProduct := make([]map[string]interface{}, 0)
		for _, v := range products {
			m := make(map[string]interface{})
			m["name"] = v.Code
			m["url"] = v.URL

			fProduct = append(fProduct, m)
		}

		return ReturnResult(c, result, 200, "Success", fProduct, true, &Users[0].ID, &HeaderLogin.XSignature, &HeaderLogin.XDeviceID)
	})

	app.Post("/login", func(c *fiber.Ctx) error {
		log.Println("POST request received at /login")
		result := make(map[string]interface{})
		result["status_code"] = "200"
		result["message"] = ""
		result["data"] = nil

		LoginPayload := new(model.LoginPayload)
		log.Println("Body : " + string(c.Body()))
		if err := c.BodyParser(LoginPayload); err != nil {
			log.Println("POST request received at /login : Error parsing body -", err.Error())
			return ReturnResult(c, result, 500, "Internal server error", nil, false, nil, nil, nil)
		}

		var Users []*model.User
		log.Printf("Searching for User with email: %s", LoginPayload.Email)
		err := pgxscan.Select(c.Context(), DB, &Users, "SELECT * FROM users WHERE email=$1", LoginPayload.Email)
		if err != nil {
			log.Println("POST request received at /login : Error fetching Users -", err.Error())
			return ReturnResult(c, result, 500, "Internal server error", nil, false, nil, nil, nil)
		}
		if len(Users) == 0 {
			log.Println("POST request received at /login : No User for email:", LoginPayload.Email)
			return ReturnResult(c, result, 400, "Bad Request", nil, false, nil, nil, nil)
		}
		if Users[0].IsTrial {
			if Users[0].TrialUntil == nil {
				log.Println("POST request received at /trial : Not Trial User:", LoginPayload.Email)
				return ReturnResult(c, result, 401, "Unauthorized", nil, false, &Users[0].ID, nil, nil)
			}
			if Users[0].TrialUntil.Before(getCurrentTime()) {
				log.Println("POST request received at /trial : Out of Trial Session:", LoginPayload.Email)
				return ReturnResult(c, result, 401, "Unauthorized", nil, false, &Users[0].ID, nil, nil)
			}
			log.Println("POST request received at /trial : Login not allowed on trial user:", LoginPayload.Email)
			return ReturnResult(c, result, 401, "Unauthorized", nil, false, &Users[0].ID, nil, nil)
		}

		if os.Getenv("ENV") != "development" && c.Get("postman-token") != "" {
			log.Println("POST request received at /login : Unauthorized access attempt")
			return ReturnResult(c, result, 401, "Unauthorized", nil, false, &Users[0].ID, nil, nil)
		}

		if c.Get("x-device-id") == "" || c.Get("x-signature") == "" {
			log.Println("POST request received at /login : missing device id or signature")
			return ReturnResult(c, result, 401, "Unauthorized", nil, false, &Users[0].ID, nil, nil)
		}

		HeaderLogin := new(model.HeaderLogin)
		HeaderLogin.XDeviceID = c.Get("x-device-id")
		HeaderLogin.XSignature = c.Get("x-signature")
		LoginPayload.Header = *HeaderLogin
		sDecode, err := base64.StdEncoding.DecodeString(HeaderLogin.XSignature)
		if err != nil {
			log.Println("POST request received at /login : Error decoding signature -", err.Error())
			return ReturnResult(c, result, 500, "Internal server error", nil, false, &Users[0].ID, &HeaderLogin.XSignature, &HeaderLogin.XDeviceID)
		}

		splitSignature := strings.Split(string(sDecode), "|")
		if len(splitSignature) != 3 || splitSignature[1] != LoginPayload.Email {
			log.Println("POST request received at /login : Invalid signature")
			return ReturnResult(c, result, 401, "Unathorized", nil, false, &Users[0].ID, &HeaderLogin.XSignature, &HeaderLogin.XDeviceID)
		}

		// prevent user not buying the product to login
		if Users[0].SubsUntil.Before(getCurrentTime()) {
			log.Println("POST request received at /login : User out of subscription: ", Users[0].SubsUntil.Format(time.RFC3339))
			return ReturnResult(c, result, 402, "Payment Required", nil, false, &Users[0].ID, &HeaderLogin.XSignature, &HeaderLogin.XDeviceID)
		}

		// fetch product
		var products []*model.Product
		log.Println("Searching for products")
		q := `SELECT * FROM product WHERE (owned_by is null and (url is not null or url != '')) or owned_by = $1 order by owned_by, code asc`

		err = pgxscan.Select(c.Context(), DB, &products, q, &Users[0].ID)
		if err != nil {
			log.Println("POST request received at /login : Error fetching products -", err.Error())
			return ReturnResult(c, result, 500, "Internal server error", nil, false, &Users[0].ID, &HeaderLogin.XSignature, &HeaderLogin.XDeviceID)
		}
		if len(products) == 0 {
			log.Println("POST request received at /login : No products")
			return ReturnResult(c, result, 500, "Internal server error", nil, false, &Users[0].ID, &HeaderLogin.XSignature, &HeaderLogin.XDeviceID)
		}

		fProduct := make([]map[string]interface{}, 0)
		for _, v := range products {
			m := make(map[string]interface{})
			m["name"] = v.Code
			m["url"] = v.URL

			fProduct = append(fProduct, m)
		}

		return ReturnResult(c, result, 200, "Success", fProduct, true, &Users[0].ID, &HeaderLogin.XSignature, &HeaderLogin.XDeviceID)
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
			qa := `UPDATE users SET subs_until = CASE
						WHEN subs_until IS NULL OR subs_until < $1::timestamptz
						THEN $1::timestamptz + concat($2::varchar, ' days')::interval
						ELSE subs_until + concat($2::varchar, ' days')::interval
					END
					WHERE id=$3`
			_, err = tx.Exec(context.Background(), qa, getCurrentTime(), Product[0].AddedDuration, User.ID)
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

func ReturnResult(c *fiber.Ctx, result map[string]interface{}, statusCode int, message string, data interface{}, isSuccess bool, userId *uuid.UUID, sig *string, deviceID *string) error {
	result["status_code"] = strconv.Itoa(statusCode)
	result["message"] = message
	result["data"] = data

	LoginLog := new(model.LoginLog)
	// LoginLog.UserID = ""
	if userId != nil {
		LoginLog.UserID = *userId
	}
	LoginLog.Signature = sig
	LoginLog.DeviceID = deviceID
	LoginLog.StatusCode = strconv.Itoa(statusCode)
	LoginLog.Message = message
	LoginLog.CreatedAt = getCurrentTime()

	// update user attempt and insert login log
	if !isSuccess {
		if userId != nil {
			_, err := DB.Exec(ctx, "UPDATE users SET failed_attempt = failed_attempt+1, login_attempt = login_attempt+1 where id=$1", userId)
			if err != nil {
				log.Println("POST request received at /login : Failed to update failed_attempt and login_attempt: ", userId)
				result["status_code"] = strconv.Itoa(statusCode)
				result["message"] = err.Error()
				result["data"] = nil
			}

			_, err1 := DB.Exec(ctx, "INSERT INTO login_log (user_id, signature, device_id, status_code, message, created_at) VALUES ($1, $2, $3, $4, $5, $6)", LoginLog.UserID, LoginLog.Signature, LoginLog.DeviceID, LoginLog.StatusCode, LoginLog.Message, LoginLog.CreatedAt)
			if err1 != nil {
				log.Println("POST request received at /login : Failed to insert login_log : ", userId)
				result["status_code"] = strconv.Itoa(statusCode)
				result["message"] = err1.Error()
				result["data"] = nil
			}
		}
	} else {
		_, err := DB.Exec(ctx, "UPDATE users SET login_attempt = login_attempt+1, last_login_at=$1 where id=$2", getCurrentTime(), userId)
		if err != nil {
			log.Println("POST request received at /login : Failed to update login_attempt: ", userId)
			result["status_code"] = strconv.Itoa(statusCode)
			result["message"] = err.Error()
			result["data"] = nil
		}
		_, err1 := DB.Exec(ctx, "INSERT INTO login_log (user_id, signature, device_id, status_code, message, created_at) VALUES ($1, $2, $3, $4, $5, $6)", LoginLog.UserID, LoginLog.Signature, LoginLog.DeviceID, LoginLog.StatusCode, LoginLog.Message, LoginLog.CreatedAt)
		if err1 != nil {
			log.Println("POST request received at /login : Failed to insert login_log : ", userId)
			result["status_code"] = strconv.Itoa(statusCode)
			result["message"] = err1.Error()
			result["data"] = nil
		}
		log.Println("Response code (" + strconv.Itoa(statusCode) + ") with data : " + fmt.Sprint(data))
	}

	return c.Status(statusCode).JSON(result)
}
