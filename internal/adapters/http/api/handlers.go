package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/encoding/charmap"

	"github.com/wallissonmarinho/GoSubs/internal/core/domain"
	"github.com/wallissonmarinho/GoSubs/internal/core/services"
)

type Handler struct {
	Subs   *services.SubtitleService
	APIKey string
	HTTP   *http.Client
}

func (h *Handler) getHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) getManifest(c *gin.Context) {
	c.JSON(http.StatusOK, h.Subs.Manifest())
}

func (h *Handler) getSubtitles(c *gin.Context) {
	tail := strings.TrimPrefix(c.Param("tail"), "/")
	if !strings.HasSuffix(tail, ".json") {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	tail = strings.TrimSuffix(tail, ".json")
	parts := strings.SplitN(tail, "/", 2)
	rawID := parts[0]
	extra := ""
	if len(parts) == 2 {
		extra = parts[1]
	}
	out, err := h.Subs.List(c.Request.Context(), c.Param("type"), rawID, extra, func(path string) string {
		return absoluteURL(c, path)
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) getSource(c *gin.Context) {
	data, err := h.Subs.ServeSource(c.Request.Context(), strings.TrimSuffix(c.Param("token"), ".srt"), h.fetchSubtitle)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/x-subrip", data)
}

func (h *Handler) getGenerate(c *gin.Context) {
	data, err := h.Subs.Generate(c.Request.Context(), strings.TrimSuffix(c.Param("token"), ".srt"), h.fetchSubtitle)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/x-subrip", data)
}

func (h *Handler) getCache(c *gin.Context) {
	data, err := h.Subs.Generate(c.Request.Context(), c.Param("key"), func(ctx context.Context, payload domain.SubtitleTokenPayload) ([]byte, error) {
		return nil, fmt.Errorf("cache miss")
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cache_miss"})
		return
	}
	c.Data(http.StatusOK, "application/x-subrip", data)
}

func (h *Handler) postCleanup(c *gin.Context) {
	out, err := h.Subs.Cleanup()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) fetchSubtitle(ctx context.Context, payload domain.SubtitleTokenPayload) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, payload.SourceURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "GoSubs/1.0")
	resp, err := h.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("subtitle source status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return normalizeSubtitleEncoding(data, payload.Encoding), nil
}

func normalizeSubtitleEncoding(data []byte, encoding string) []byte {
	if utf8.Valid(data) {
		return data
	}
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "latin-1", "latin1", "iso-8859-1", "iso8859-1":
		out, err := charmap.ISO8859_1.NewDecoder().Bytes(data)
		if err == nil {
			return out
		}
	}
	out, err := charmap.ISO8859_1.NewDecoder().Bytes(data)
	if err == nil {
		return out
	}
	return data
}

func absoluteURL(c *gin.Context, path string) string {
	scheme := "https"
	if c.Request.TLS == nil {
		scheme = "http"
	}
	if xf := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")); xf != "" {
		scheme = xf
	}
	host := c.Request.Host
	if h := strings.TrimSpace(c.GetHeader("X-Forwarded-Host")); h != "" {
		host = h
	}
	u := url.URL{Scheme: scheme, Host: host, Path: path}
	return u.String()
}
