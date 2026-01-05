package model

import "time"

type EmailWebhook struct {
	ID        uint      `json:"primaryKey"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Cc        string    `json:"cc"`
	Subject   string    `json:"subject"`
	BodyPlain string    `json:"body_plain"`
	BodyHTML  string    `json:"body_html"`
	MessageID string    `json:"message_id"`
	Date      time.Time `json:"date"`
	Header    HeaderEmailWebhook
}
