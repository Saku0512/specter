package server

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/Saku0512/specter/config"
)

func New(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	for _, route := range cfg.Route {
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
