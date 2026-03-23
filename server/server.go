package server

import (
	"net/http"
	"sync/atomic"

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
	for _, route := range cfg.Routes {
		rt := route
		r.Handle(rt.Method, rt.Path, func(c *gin.Context) {
			status := rt.Status
			if status == 0 {
				status = http.StatusOK
			}
			c.JSON(status, rt.Response)
		})
	}
	return r
}
