package cors

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	defaultAllowHeaders = []string{"Origin", "Accept", "Content-Type", "Authorization"}
	defaultAllowMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD"}
)

// Options stores configurations
type Options struct {
	AllowOrigins     []string
	AllowCredentials bool
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	MaxAge           time.Duration
}

// Middleware sets CORS headers for every request
func Middleware(options Options) gin.HandlerFunc {
	if options.AllowHeaders == nil {
		options.AllowHeaders = defaultAllowHeaders
	}

	if options.AllowMethods == nil {
		options.AllowMethods = defaultAllowMethods
	}

	return func(c *gin.Context) {
		req := c.Request
		res := c.Writer
		origin := req.Header.Get("Origin")
		requestMethod := req.Header.Get("Access-Control-Request-Method")
		requestHeaders := req.Header.Get("Access-Control-Request-Headers")

		if len(options.AllowOrigins) > 0 {
			res.Header().Set("Access-Control-Allow-Origin", strings.Join(options.AllowOrigins, " "))
		} else {
			res.Header().Set("Access-Control-Allow-Origin", origin)
		}

		if options.AllowCredentials {
			res.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if len(options.ExposeHeaders) > 0 {
			res.Header().Set("Access-Control-Expose-Headers", strings.Join(options.ExposeHeaders, ","))
		}

		if req.Method == "OPTIONS" {
			if len(options.AllowMethods) > 0 {
				res.Header().Set("Access-Control-Allow-Methods", strings.Join(options.AllowMethods, ","))
			} else if requestMethod != "" {
				res.Header().Set("Access-Control-Allow-Methods", requestMethod)
			}

			if len(options.AllowHeaders) > 0 {
				res.Header().Set("Access-Control-Allow-Headers", strings.Join(options.AllowHeaders, ","))
			} else if requestHeaders != "" {
				res.Header().Set("Access-Control-Allow-Headers", requestHeaders)
			}

			if options.MaxAge > time.Duration(0) {
				res.Header().Set("Access-Control-Max-Age", strconv.FormatInt(int64(options.MaxAge/time.Second), 10))
			}

			c.AbortWithStatus(http.StatusOK)
		} else {
			c.Next()
		}
	}
}
