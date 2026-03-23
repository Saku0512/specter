package server

import (
	"bytes"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Saku0512/specter/config"
	"github.com/gin-gonic/gin"
)

type Server struct {
	engine  atomic.Pointer[gin.Engine]
	verbose bool
}

func New(cfg *config.Config, verbose bool) *Server {
	s := &Server{verbose: verbose}
	s.engine.Store(newEngine(cfg, verbose))
	return s
}

func (s *Server) Reload(cfg *config.Config) {
	s.engine.Store(newEngine(cfg, s.verbose))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.engine.Load().ServeHTTP(w, r)
}

func newEngine(cfg *config.Config, verbose bool) *gin.Engine {
	r := gin.Default()

	if cfg.CORS {
		r.Use(corsMiddleware())
	}

	if verbose {
		r.Use(verboseLogger())
	}

	for _, route := range cfg.Routes {
		rt := route

		var counter atomic.Uint64

		r.Handle(rt.Method, rt.Path, func(c *gin.Context) {
			if rt.Delay > 0 {
				time.Sleep(time.Duration(rt.Delay) * time.Millisecond)
			}
			for k, v := range rt.Headers {
				c.Header(k, v)
			}

			for _, m := range rt.Match {
				if matchesQuery(c, m.Query) {
					status := m.Status
					if status == 0 {
						status = http.StatusOK
					}
					c.JSON(status, applyParams(m.Response, c.Params))
					return
				}
			}

			if len(rt.Responses) > 0 {
				var picked config.RouteResponse
				switch rt.Mode {
				case "random":
					picked = rt.Responses[rand.IntN(len(rt.Responses))]
				default: // sequential
					idx := counter.Add(1) - 1
					picked = rt.Responses[idx%uint64(len(rt.Responses))]
				}
				status := picked.Status
				if status == 0 {
					status = http.StatusOK
				}
				c.JSON(status, applyParams(picked.Response, c.Params))
				return
			}

			status := rt.Status
			if status == 0 {
				status = http.StatusOK
			}
			c.JSON(status, applyParams(rt.Response, c.Params))
		})
	}

	return r
}

func verboseLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("→ %s %s", c.Request.Method, c.Request.URL.RequestURI())

		for k, v := range c.Request.Header {
			log.Printf("  %s: %s", k, strings.Join(v, ", "))
		}

		if c.Request.Body != nil && c.Request.ContentLength != 0 {
			body, err := io.ReadAll(c.Request.Body)
			if err == nil && len(body) > 0 {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
				log.Printf("  Body: %s", body)
			}
		}

		c.Next()
	}
}

// applyParams recursively replaces ":paramName" strings in the response with
// actual path parameter values. Numeric params are converted to int or float64.
func applyParams(v any, params gin.Params) any {
	switch val := v.(type) {
	case string:
		if strings.HasPrefix(val, ":") {
			if p := params.ByName(val[1:]); p != "" {
				if n, err := strconv.Atoi(p); err == nil {
					return n
				}
				if f, err := strconv.ParseFloat(p, 64); err == nil {
					return f
				}
				return p
			}
		}
		return val
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, v2 := range val {
			out[k] = applyParams(v2, params)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, v2 := range val {
			out[i] = applyParams(v2, params)
		}
		return out
	default:
		return v
	}
}

func matchesQuery(c *gin.Context, query map[string]string) bool {
	for k, v := range query {
		if c.Query(k) != v {
			return false
		}
	}
	return true
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
