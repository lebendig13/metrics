package server

import (
	"net/http"

	"github.com/lebendig13/metrics/internal/hash"
)

type hashWriter struct {
	http.ResponseWriter
	hashKey string
}

func (w *hashWriter) Write(b []byte) (int, error) {
	hashValue, err := hash.GetHashSha256(w.hashKey, b)
	if err != nil {
		return 0, err
	}

	w.Header().Set("HashSHA256", hashValue)
	return w.ResponseWriter.Write(b)
}
