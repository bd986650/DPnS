package cityid

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func FromEN(cityEN, countryEN string) string {
	s := strings.ToLower(strings.TrimSpace(cityEN)) + "|" + strings.ToLower(strings.TrimSpace(countryEN))
	h := sha256.Sum256([]byte(s))
	return "c" + hex.EncodeToString(h[:8])
}
