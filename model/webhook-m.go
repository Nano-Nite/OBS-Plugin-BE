package model

type WebhookPayload struct {
	Event string `json:"event"`
	Data  Data   `json:"data"`
}

type Data struct {
	MessageAction string      `json:"message_action"`
	MessageCode   string      `json:"message_code"`
	MessageData   MessageData `json:"message_data"`
	MessageDesc   string      `json:"message_desc"`
	MessageID     string      `json:"message_id"`
	MessageTitle  string      `json:"message_title"`
}

type MessageData struct {
	CreatedAt       string   `json:"createdAt"`
	Customer        Customer `json:"customer"`
	Items           []Item   `json:"items"`
	RefID           string   `json:"refId"`
	ShippingAddress string   `json:"shippingAddress"`
	ShippingInfo    string   `json:"shippingInfo"`
	Totals          Totals   `json:"totals"`
}

type Customer struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type Item struct {
	Addons                 []Addon                `json:"addons"`
	AppointmentData        map[string]interface{} `json:"appointment_data"`
	PafID                  string                 `json:"pafId"`
	Price                  int                    `json:"price"`
	PublicAffiliateContent map[string]interface{} `json:"public_affiliate_content"`
	Qty                    int                    `json:"qty"`
	Stock                  int                    `json:"stock"`
	Title                  string                 `json:"title"`
	UUID                   string                 `json:"uuid"`
}

type Addon struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price string `json:"price"` // "50,000" → keep as string
}

type Totals struct {
	Affiliate      int `json:"affiliate"`
	ConvenienceFee int `json:"convenienceFee"`
	Discount       int `json:"discount"`
	GrandTotal     int `json:"grandTotal"`
	TotalAddon     int `json:"totalAddon"`
	TotalItem      int `json:"totalItem"`
	TotalPrice     int `json:"totalPrice"`
	TotalShipping  int `json:"totalShipping"`
}
