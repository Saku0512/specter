package server

import (
	"net/http"
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
		r.Handle(rt.Method, rt.Path, func(c *gin.Context) {
			if rt.Delay > 0 {
				time.Sleep(time.Duration(rt.Delay) * time.Millisecond)
			}
			status := rt.Status
			if status == 0 {
				status = http.StatusOK
			}
			c.JSON(status, rt.Response)
		})
	}

	return r
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
