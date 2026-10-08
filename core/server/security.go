package server

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS 白名单：通过 FAFA_CORS_ORIGINS 环境变量配置（逗号分隔）。
// 未配置时保持兼容（允许任意 Origin）。
var corsOrigins = func() map[string]struct{} {
	m := make(map[string]struct{})
	for _, o := range strings.Split(os.Getenv("FAFA_CORS_ORIGINS"), ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			m[o] = struct{}{}
		}
	}
	return m
}()

func corsAllowOrigin(origin string) bool {
	if len(corsOrigins) == 0 {
		return true
	}
	_, ok := corsOrigins[origin]
	return ok
}

// SecurityHeaders 基础安全响应头（HTTPS 就绪后可在 nginx 补 HSTS/CSP）。
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}
