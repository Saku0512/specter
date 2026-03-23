package server

import (
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
	engine atomic.Pointer[gin.Engine]
}

func New(cfg *config.Config) *Server {
	s := &Server{}
	s.engine.Store(newEngine(cfg))
	return s
}

func (s *Server) Reload(cfg *config.Config) {
	s.engine.Store(newEngine(cfg))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.engine.Load().ServeHTTP(w, r)
}

func newEngine(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	if cfg.CORS {
		r.Use(corsMiddleware())
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
