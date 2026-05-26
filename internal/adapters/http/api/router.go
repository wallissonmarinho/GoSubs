package api

import "github.com/gin-gonic/gin"

func NewRouter(adminKey string, h *Handler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	e := gin.New()
	e.Use(gin.Recovery())

	e.GET("/healthz", h.getHealth)
	e.GET("/manifest.json", h.getManifest)
	e.GET("/subtitles/:type/*tail", h.getSubtitles)
	e.GET("/subtitles/source/:token", h.getSource)
	e.GET("/subtitles/generate/:token", h.getGenerate)
	e.GET("/subtitles/cache/:key", h.getCache)

	admin := e.Group("/admin")
	admin.Use(BearerAuth(adminKey))
	admin.POST("/subtitles/cache/cleanup", h.postCleanup)

	return e
}
