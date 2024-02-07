package router

import (
	"net/http"

	"realtimechat/controller"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	router := gin.Default()

	router.GET("", func(context *gin.Context) {
		context.JSON(http.StatusOK, "welcome To Real Time Chat Application")
	})

	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	router.GET("/wschat", controller.PrivateChat)
	router.GET("/wschat/broadcast", controller.Broadcast)

	return router
}
