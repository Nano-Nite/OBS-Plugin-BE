package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type UserInfo struct {
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	Phone     string   `json:"phone"`
	Orders    []string `json:"orders"`
	TriggerWA bool     `json:"trigger_wa"`
}

type LoginPayload struct {
	Header HeaderLogin `json:"header"`
	Email  string      `json:"email"`
}

type LoginLog struct {
	ID            uuid.UUID `db:"id" json:"id"`
	Email         string    `db:"email" json:"email"`
	Signature     string    `db:"signature" json:"signature"`
	DeviceID      string    `db:"device_id" json:"device_id"`
	FailedAttempt int       `db:"failed_attempt" json:"failed_attempt"`
	LastLogin     time.Time `db:"last_login" json:"last_login"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type PurchaseOrder struct {
	ID        uuid.UUID       `db:"id" json:"id"`
	MessageID string          `db:"message_id" json:"message_id"`
	Name      string          `db:"name" json:"name"`
	Email     string          `db:"email" json:"email"`
	Phone     string          `db:"phone" json:"phone"`
	Orders    json.RawMessage `db:"orders" json:"orders"`
	TriggerWA bool            `db:"trigger_wa" json:"trigger_wa"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
}
