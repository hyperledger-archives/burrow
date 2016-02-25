package cors

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type requestOptions struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    io.Reader
}

func request(server *gin.Engine, options requestOptions) *httptest.ResponseRecorder {
	if options.Method == "" {
		options.Method = "GET"
	}

	w := httptest.NewRecorder()
	req, err := http.NewRequest(options.Method, options.URL, options.Body)

	if options.Headers != nil {
		for key, value := range options.Headers {
			req.Header.Set(key, value)
		}
	}

	server.ServeHTTP(w, req)

	if err != nil {
		panic(err)
	}

	return w
}

func newServer() *gin.Engine {
	g := gin.New()
	g.Use(Middleware(Options{}))

	return g
}

func TestDefault(t *testing.T) {
	g := newServer()
	assert := assert.New(t)

	g.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	r := request(g, requestOptions{
		URL: "/test",
		Headers: map[string]string{
			"Origin": "http://maji.moe",
		},
	})

	assert.Equal("http://maji.moe", r.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal("OK", r.Body.String())
}

func TestAllowOrigins(t *testing.T) {
	g := gin.New()
	g.Use(Middleware(Options{
		AllowOrigins: []string{"http://maji.moe", "http://example.com"},
	}))
	assert := assert.New(t)

	g.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	r := request(g, requestOptions{
		URL: "/test",
		Headers: map[string]string{
			"Origin": "http://maji.moe",
		},
	})

	assert.Equal("http://maji.moe http://example.com", r.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal("OK", r.Body.String())
}

func TestAllowCredentials(t *testing.T) {
	g := gin.New()
	g.Use(Middleware(Options{
		AllowCredentials: true,
	}))
	assert := assert.New(t)

	g.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	r := request(g, requestOptions{
		URL: "/test",
		Headers: map[string]string{
			"Origin": "http://maji.moe",
		},
	})

	assert.Equal("true", r.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal("OK", r.Body.String())
}

func TestExposeHeaders(t *testing.T) {
	g := gin.New()
	g.Use(Middleware(Options{
		ExposeHeaders: []string{"Foo", "Bar"},
	}))
	assert := assert.New(t)

	g.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	r := request(g, requestOptions{
		URL: "/test",
		Headers: map[string]string{
			"Origin": "http://maji.moe",
		},
	})

	assert.Equal("Foo,Bar", r.Header().Get("Access-Control-Expose-Headers"))
	assert.Equal("OK", r.Body.String())
}

func TestOptionsRequest(t *testing.T) {
	g := newServer()
	assert := assert.New(t)

	r := request(g, requestOptions{
		Method: "OPTIONS",
		URL:    "/",
		Headers: map[string]string{
			"Origin": "http://maji.moe",
		},
	})

	assert.Equal("http://maji.moe", r.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal("GET,POST,PUT,DELETE,PATCH,HEAD", r.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal("Origin,Accept,Content-Type,Authorization", r.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal("", r.Body.String())
	assert.Equal(200, r.Code)
}

func TestAllowMethods(t *testing.T) {
	g := gin.New()
	g.Use(Middleware(Options{
		AllowMethods: []string{"GET", "POST", "PUT"},
	}))
	assert := assert.New(t)

	r := request(g, requestOptions{
		Method: "OPTIONS",
		URL:    "/",
		Headers: map[string]string{
			"Origin": "http://maji.moe",
		},
	})

	assert.Equal("GET,POST,PUT", r.Header().Get("Access-Control-Allow-Methods"))
}

func TestRequestMethod(t *testing.T) {
	g := gin.New()
	g.Use(Middleware(Options{
		AllowMethods: []string{},
	}))
	assert := assert.New(t)

	r := request(g, requestOptions{
		Method: "OPTIONS",
		URL:    "/",
		Headers: map[string]string{
			"Origin":                        "http://maji.moe",
			"Access-Control-Request-Method": "PUT",
		},
	})

	assert.Equal("PUT", r.Header().Get("Access-Control-Allow-Methods"))
}

func TestAllowHeaders(t *testing.T) {
	g := gin.New()
	g.Use(Middleware(Options{
		AllowHeaders: []string{"X-Custom-Header", "X-Auth-Token"},
	}))
	assert := assert.New(t)

	r := request(g, requestOptions{
		Method: "OPTIONS",
		URL:    "/",
		Headers: map[string]string{
			"Origin": "http://maji.moe",
		},
	})

	assert.Equal("X-Custom-Header,X-Auth-Token", r.Header().Get("Access-Control-Allow-Headers"))
}

func TestRequestHeaders(t *testing.T) {
	g := gin.New()
	g.Use(Middleware(Options{
		AllowHeaders: []string{},
	}))
	assert := assert.New(t)

	r := request(g, requestOptions{
		Method: "OPTIONS",
		URL:    "/",
		Headers: map[string]string{
			"Origin": "http://maji.moe",
			"Access-Control-Request-Headers": "Foo,Bar",
		},
	})

	assert.Equal("Foo,Bar", r.Header().Get("Access-Control-Allow-Headers"))
}

func TestMaxAge(t *testing.T) {
	g := gin.New()
	g.Use(Middleware(Options{
		MaxAge: time.Hour,
	}))
	assert := assert.New(t)

	r := request(g, requestOptions{
		Method: "OPTIONS",
		URL:    "/",
		Headers: map[string]string{
			"Origin": "http://maji.moe",
		},
	})

	assert.Equal("3600", r.Header().Get("Access-Control-Max-Age"))
}
