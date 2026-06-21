package utils

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

// GeneratePublicKey generates a public API key for project ingestion (ock_pub_...)
func GeneratePublicKey() string {
	return "ock_pub_" + randomString(24)
}

// GenerateSecretKey generates a secret API key for server side use (ock_sec_...)
func GenerateSecretKey() string {
	return "ock_sec_" + randomString(32)
}

func randomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	str := base64.URLEncoding.EncodeToString(b)
	str = strings.ReplaceAll(str, "=", "")
	str = strings.ReplaceAll(str, "-", "")
	str = strings.ReplaceAll(str, "_", "")
	if len(str) > length {
		return str[:length]
	}
	return str
}
