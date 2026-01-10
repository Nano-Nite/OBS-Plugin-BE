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
	ID            uuid.UUID  `db:"id" json:"id"`
	Name          string     `db:"name" json:"name"`
	Email         string     `db:"email" json:"email"`
	Phone         string     `db:"phone" json:"phone"`
	IsTrial       bool       `db:"is_trial" json:"is_trial"`
	TrialUntil    *time.Time `db:"trial_until" json:"trial_until"`
	SubsUntil     time.Time  `db:"subs_until" json:"subs_until"`
	LastPurchase  time.Time  `db:"last_purchase" json:"last_purchase"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	LastLoginAt   *time.Time `db:"last_login_at" json:"last_login_at"`
	FailedAttempt int        `db:"failed_attempt" json:"failed_attempt"`
	LoginAttempt  int        `db:"login_attempt" json:"login_attempt"`
}

type LoginPayload struct {
	Header HeaderLogin `json:"header"`
	Email  string      `json:"email"`
}

type LoginLog struct {
	ID         uuid.UUID `db:"id" json:"id"`
	UserID     uuid.UUID `db:"user_id" json:"user_id"`
	Signature  *string   `db:"signature" json:"signature"`
	DeviceID   *string   `db:"device_id" json:"device_id"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	StatusCode string    `db:"status_code" json:"status_code"`
	Message    string    `db:"message" json:"message"`
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
	ID            uuid.UUID  `db:"id" json:"id"`
	OwnedBy       *uuid.UUID `db:"owned_by" json:"owned_by"`
	Name          string     `db:"name" json:"name"`
	Code          string     `db:"code" json:"code"`
	URL           *string    `db:"url" json:"url"`
	Price         float64    `db:"price" json:"price"`
	AddedDuration int        `db:"added_duration" json:"added_duration"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
}
