package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// VerifyXenditSignature verifies Xendit webhook signature
// https://developers.xendit.co/api-reference/#webhooks
func VerifyXenditSignature(payload []byte, signature, token string) bool {
	// TODO: Implement actual Xendit signature verification
	// For POC, we'll use a simple HMAC-SHA256 verification

	mac := hmac.New(sha256.New, []byte(token))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// VerifyCustodySignature verifies Custody provider webhook signature
func VerifyCustodySignature(payload []byte, signature, secret string) bool {
	// TODO: Implement actual Custody signature verification
	// For POC, we'll use a simple HMAC-SHA256 verification

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// GenerateWebhookID generates a unique webhook ID
func GenerateWebhookID(provider, externalID string) string {
	return fmt.Sprintf("%s:%s", provider, externalID)
}
