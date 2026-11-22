package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
)

func getEncryptionKey() []byte {
	seed := strings.TrimSpace(os.Getenv("BINANCE_ENCRYPTION_KEY"))
	if seed == "" {
		seed = strings.TrimSpace(os.Getenv("JWT_SECRET"))
	}
	if seed == "" {
		seed = "binance_journal_default_secret_key_2026"
	}
	hash := sha256.Sum256([]byte(seed))
	return hash[:]
}

func EncryptSecret(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	block, err := aes.NewCipher(getEncryptionKey())
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	cipherText := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func DecryptSecret(cipherB64 string) (string, error) {
	if cipherB64 == "" {
		return "", nil
	}
	raw, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(getEncryptionKey())
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(raw) < nonceSize {
		return "", fmt.Errorf("ข้อมูลเข้ารหัสไม่สมบูรณ์")
	}
	nonce, cipherText := raw[:nonceSize], raw[nonceSize:]
	plain, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func MaskAPIKey(key string) string {
	k := strings.TrimSpace(key)
	if len(k) <= 8 {
		return "****"
	}
	return k[:4] + "••••••••" + k[len(k)-4:]
}
