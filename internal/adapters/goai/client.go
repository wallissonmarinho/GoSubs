package goai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/wallissonmarinho/GoSubs/internal/core/domain"
)

type Client struct {
	BaseURL   string
	APIKey    string
	UserAgent string
	HTTP      *http.Client
}

type translateRequest struct {
	SourceLanguage string                   `json:"source_language,omitempty"`
	TargetLanguage string                   `json:"target_language"`
	Items          []domain.SubtitleTextChunk `json:"items"`
}

type translateResponse struct {
	Items []domain.SubtitleTextChunk `json:"items"`
}

func (c *Client) TranslateBatch(ctx context.Context, sourceLang, targetLang string, items []domain.SubtitleTextChunk) ([]domain.SubtitleTextChunk, error) {
	hc := c.HTTP
	if hc == nil {
		hc = &http.Client{Timeout: 60 * time.Second}
	}
	body, _ := json.Marshal(translateRequest{
		SourceLanguage: sourceLang,
		TargetLanguage: targetLang,
		Items:          items,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+"/v1/translate/subtitles", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("User-Agent", defaultString(c.UserAgent, "GoSubs/1.0"))
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("goai translate status %d", resp.StatusCode)
	}
	var out translateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

func defaultString(v, fallback string) string {
	if v != "" {
		return v
	}
	return fallback
}
