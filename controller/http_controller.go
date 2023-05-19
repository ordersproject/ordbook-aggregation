package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"net/http"
	"ordbook-aggregation/config"
	"ordbook-aggregation/controller/middleware"
	"time"
)


func Run() {
	router := gin.Default()
	router.Use(Cors())
	router.Use(Logger())
	router.Use(middleware.ResponseTime())


	limiter := middleware.NewIPRateLimiter(1 * time.Second, 120)
	router.Use(middleware.IPRateLimitMiddleware(limiter))


	// meta
	brc20 := router.Group("/brc20")
	{
		brc20.POST("/order/push", PushOrder)
		brc20.GET("/orders", FetchOrders)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	_ = router.Run(fmt.Sprintf("0.0.0.0:%s", config.Port))
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		//origin := c.Request.Header.Get("Origin")
		//if origin != "" {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization,X-API-KEY")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Set("content-type", "application/json")
		//}
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

func Logger() gin.HandlerFunc {
	return func(context *gin.Context) {
		//context.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		//context.Abort()
		context.Next()
	}
}

func Handle(r *gin.Engine, httpMethods []string, relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	var routes gin.IRoutes
	for _, httpMethod := range httpMethods {
		routes = r.Handle(httpMethod, relativePath, handlers...)
	}
	return routes
}

