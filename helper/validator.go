package helper

import (
	"crypto/sha256"
	"encoding/hex"
)

func ValidateLynkSignature(refID, amount string, message_id string, receivedSignature, secretKey string) bool {
	signatureString := amount + refID + message_id + secretKey
	hash := sha256.New()
	hash.Write([]byte(signatureString))
	calculatedSignature := hex.EncodeToString(hash.Sum(nil))
	return calculatedSignature == receivedSignature
}
