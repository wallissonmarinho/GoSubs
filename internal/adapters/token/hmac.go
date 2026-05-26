package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wallissonmarinho/GoSubs/internal/core/domain"
)

type HMACCodec struct {
	Secret []byte
}

func (c HMACCodec) Encode(payload domain.SubtitleTokenPayload) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sig := c.sign(body)
	return base64.RawURLEncoding.EncodeToString(body) + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func (c HMACCodec) Decode(token string) (domain.SubtitleTokenPayload, error) {
	var out domain.SubtitleTokenPayload
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return out, fmt.Errorf("invalid token")
	}
	body, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return out, err
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return out, err
	}
	if !hmac.Equal(sig, c.sign(body)) {
		return out, fmt.Errorf("invalid signature")
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return out, err
	}
	return out, nil
}

func (c HMACCodec) sign(body []byte) []byte {
	m := hmac.New(sha256.New, c.Secret)
	_, _ = m.Write(body)
	return m.Sum(nil)
}
