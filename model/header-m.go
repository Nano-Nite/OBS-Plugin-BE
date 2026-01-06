package model

type HeaderEmailWebhook struct {
	APIKey string
	Source string
}

type HeaderLogin struct {
	XSignature string
	XDeviceID  string
}
