package app

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/wallissonmarinho/GoSubs/internal/adapters/goai"
	"github.com/wallissonmarinho/GoSubs/internal/adapters/http/api"
	"github.com/wallissonmarinho/GoSubs/internal/adapters/localcache"
	"github.com/wallissonmarinho/GoSubs/internal/adapters/token"
	"github.com/wallissonmarinho/GoSubs/internal/adapters/wyzie"
	appcfg "github.com/wallissonmarinho/GoSubs/internal/app/config"
	"github.com/wallissonmarinho/GoSubs/internal/core/services"
)

func Wire(cfg appcfg.Config) *gin.Engine {
	httpClient := &http.Client{Timeout: 45 * time.Second}
	provider := &wyzie.Client{
		BaseURL:   cfg.WyzieBaseURL,
		APIKey:    cfg.WyzieAPIKey,
		Source:    cfg.WyzieSource,
		UserAgent: cfg.UserAgent,
		HTTP:      httpClient,
	}
	translator := &goai.Client{
		BaseURL:   cfg.GoAIBaseURL,
		APIKey:    cfg.GoAIAPIKey,
		UserAgent: cfg.UserAgent,
		HTTP:      &http.Client{Timeout: 90 * time.Second},
	}
	cache := &localcache.Cache{Dir: cfg.CacheDir, TTL: cfg.CacheTTL}
	tokenCodec := token.HMACCodec{Secret: []byte(cfg.TokenSecret)}
	subs := services.NewSubtitleService(provider, translator, cache, tokenCodec, cfg.MaxBatchChars)
	h := &api.Handler{Subs: subs, APIKey: cfg.AdminAPIKey, HTTP: httpClient}
	return api.NewRouter(cfg.AdminAPIKey, h)
}
