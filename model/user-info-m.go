package model

type UserInfo struct {
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	Phone     string   `json:"phone"`
	Orders    []string `json:"orders"`
	TriggerWA bool     `json:"trigger_wa"`
}
