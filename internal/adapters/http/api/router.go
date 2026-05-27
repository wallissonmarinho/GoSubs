package api

import "github.com/gin-gonic/gin"

func NewRouter(adminKey string, h *Handler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	e := gin.New()
	e.Use(gin.Recovery())
	e.Use(corsMiddleware())

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

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, HEAD")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
