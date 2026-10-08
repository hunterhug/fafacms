package server

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

func Server() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	gin.ForceConsoleColor()

	r := gin.New()

	// LoggerWithFormatter middleware will write the logs to gin.DefaultWriter
	// By default gin.DefaultWriter = os.Stdout
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {

		// your custom format
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s\" - %s - %s\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	// 安全响应头
	r.Use(SecurityHeaders())

	// 反爬限流
	r.Use(RateLimit())

	// 请求签名（防重放，默认宽松，FAFA_SIGN_STRICT=1 强制）
	r.Use(RequestSign())

	// Solve Cross
	r.Use(cors.New(cors.Config{
		AllowOriginFunc:  corsAllowOrigin,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Auth", "X-Ts", "X-Nonce", "X-Sign"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	return r
}
