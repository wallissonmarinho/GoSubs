package wyzie

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/wallissonmarinho/GoSubs/internal/core/domain"
)

type Client struct {
	BaseURL   string
	APIKey    string
	Source    string
	UserAgent string
	HTTP      *http.Client
}

type searchItem struct {
	ID                string   `json:"id"`
	URL               string   `json:"url"`
	Format            string   `json:"format"`
	Encoding          string   `json:"encoding"`
	Display           string   `json:"display"`
	Language          string   `json:"language"`
	Source            string   `json:"source"`
	Release           string   `json:"release"`
	Releases          []string `json:"releases"`
	FileName          string   `json:"fileName"`
	MatchedRelease    string   `json:"matchedRelease"`
	MatchedFilter     string   `json:"matchedFilter"`
	AI                bool     `json:"ai"`
	IsHearingImpaired bool     `json:"isHearingImpaired"`
}

func (c *Client) Search(ctx context.Context, req domain.SubtitleSearchRequest) ([]domain.SubtitleCandidate, error) {
	hc := c.HTTP
	if hc == nil {
		hc = &http.Client{Timeout: 30 * time.Second}
	}
	q := url.Values{}
	if req.Media.IMDbID != "" {
		q.Set("id", req.Media.IMDbID)
	} else {
		q.Set("id", strconv.Itoa(req.Media.TMDBID))
	}
	if req.Media.Season > 0 && req.Media.Episode > 0 {
		q.Set("season", strconv.Itoa(req.Media.Season))
		q.Set("episode", strconv.Itoa(req.Media.Episode))
	}
	q.Set("format", "srt")
	q.Set("language", "pt,en,es")
	q.Set("source", defaultString(c.Source, "all"))
	q.Set("key", c.APIKey)
	if strings.TrimSpace(req.ReleaseHint) != "" {
		q.Set("fileName", req.ReleaseHint)
	}
	endpoint := strings.TrimRight(c.BaseURL, "/") + "/search?" + q.Encode()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", defaultString(c.UserAgent, "GoSubs/1.0"))
	resp, err := hc.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wyzie search status %d", resp.StatusCode)
	}
	var items []searchItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}
	out := make([]domain.SubtitleCandidate, 0, len(items))
	for _, item := range items {
		out = append(out, domain.SubtitleCandidate{
			ID:                item.ID,
			URL:               item.URL,
			Format:            item.Format,
			Encoding:          item.Encoding,
			Display:           item.Display,
			Language:          item.Language,
			Source:            item.Source,
			Release:           item.Release,
			FileName:          item.FileName,
			MatchedRelease:    item.MatchedRelease,
			MatchedFilter:     item.MatchedFilter,
			AI:                item.AI,
			IsHearingImpaired: item.IsHearingImpaired,
		})
	}
	return out, nil
}

func defaultString(v, fallback string) string {
	if v != "" {
		return v
	}
	return fallback
}
