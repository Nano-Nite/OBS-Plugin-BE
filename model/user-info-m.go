package model

import (
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

type User struct {
	ID           uuid.UUID `db:"id" json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	IsTrial      bool      `json:"is_trial"`
	TrialUntil   time.Time `json:"trial_until"`
	SubsUntil    time.Time `json:"subs_until"`
	LastPurchase time.Time `json:"last_purchase"`
	CreatedAt    time.Time `json:"created_at"`
}

type LoginPayload struct {
	Header HeaderLogin `json:"header"`
	Email  string      `json:"email"`
}

type LoginLog struct {
	ID            uuid.UUID `db:"id" json:"id"`
	UserID        uuid.UUID `db:"user_id" json:"user_id"`
	Signature     string    `db:"signature" json:"signature"`
	DeviceID      string    `db:"device_id" json:"device_id"`
	FailedAttempt int       `db:"failed_attempt" json:"failed_attempt"`
	LastLogin     time.Time `db:"last_login" json:"last_login"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type PurchaseOrder struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	ProductID uuid.UUID `db:"product_id" json:"product_id"`
	MessageID string    `db:"message_id" json:"message_id"`
	TriggerWA bool      `db:"trigger_wa" json:"trigger_wa"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Product struct {
	ID            uuid.UUID `db:"id" json:"id"`
	Name          string    `db:"name" json:"name"`
	Code          string    `db:"code" json:"code"`
	URL           *string   `db:"url" json:"url"`
	Price         float64   `db:"price" json:"price"`
	AddedDuration int       `db:"added_duration" json:"added_duration"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}
