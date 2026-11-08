package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

func GetHashSha256(hashKey string, rawBytes []byte) (string, error) {
	if hashKey == "" {
		return "", errors.New("hash key cannot be empty")
	}

	h := hmac.New(sha256.New, []byte(hashKey))
	h.Write(rawBytes)
	resultHash := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(resultHash), nil
}

func GetHashFromString(hashStr string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(hashStr)
}

func CheckHashSha256(hashKey string, expectedHash []byte, rawBytes []byte) bool {
	h := hmac.New(sha256.New, []byte(hashKey))
	h.Write(rawBytes)
	resultHash := h.Sum(nil)
	return hmac.Equal(expectedHash, resultHash)
}
