package server

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/Saku0512/specter/config"
)

func New(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	for _, route := range cfg.Route {
		r := route

		r.Handle(route.Method, route.Path, func(c *gin.Context) {
			status := r.Status
			if status == 0 {
				status = http.StatusOK
			}
			c.JSON(status, r.Response)
		})
	}

	return r
}
