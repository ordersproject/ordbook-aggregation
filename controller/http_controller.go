package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
	"ordbook-aggregation/config"
	"ordbook-aggregation/controller/auth"
	"ordbook-aggregation/controller/middleware"
	_ "ordbook-aggregation/docs"
)

func Run() {
	router := gin.Default()
	router.Use(Cors())
	router.Use(Logger())
	router.Use(middleware.ResponseTime())

	//limiter := middleware.NewIPRateLimiter(1 * time.Second, 120)
	//router.Use(middleware.IPRateLimitMiddleware(limiter))

	// brc20
	brc20 := router.Group("/brc20")
	{
		brc20.POST("/order/push", auth.AuthSignMiddleware(), PushOrder)
		brc20.POST("/order/ask/push", auth.AuthSignMiddleware(), PushOrder)
		brc20.GET("/orders", FetchOrders)
		brc20.GET("/order/:orderId", auth.AuthSignMiddleware(), FetchOneOrder)
		brc20.GET("/orders/user/:address", FetchUserOrders)
		brc20.GET("/tickers", FetchTicker)
		brc20.GET("/kline", FetchKline)

		brc20.GET("/order/bid/pre", FetchPreBid)
		brc20.GET("/order/bid", FetchBidPsbt)
		brc20.POST("/order/bid/push", auth.AuthSignMiddleware(), UpdateBidPsbt)
		brc20.POST("/order/bid/do", DoBid)
		brc20.POST("/order/update", auth.AuthSignMiddleware(), UpdateOrder)

		brc20.POST("/inscribe/pre", PreInscribe)
		brc20.POST("/inscribe/commit", CommitInscribe)

		brc20.POST("/transfer/colddown", ColdDownBrc20Transfer)
		brc20.POST("/transfer/colddown/batch", ColdDownBrc20TransferBatch)
		brc20.POST("/transfer/colddown/batch/ask", ColdDownBatchBrc20TransferAndMakeAsk)
		brc20.POST("/utxo/colddown", ColdDownUtxo)

		brc20.GET("/ws/uuid", GetWsUuid)
		brc20.GET("/check/info", CheckBrc20)
		brc20.GET("/address/:address/:tick", GetBrc20BalanceDetail)
		brc20.GET("/address/:address/balance/info", GetBrc20BalanceList)

		brc20.GET("/order/bid/dummy/:address", GetBidDummyList)

		brc20.GET("/claim/order", FetchClaimOrder)
		brc20.POST("/claim/order/update", UpdateClaimOrder)
	}

	url := ginSwagger.URL("/swagger/doc.json")
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	_ = router.Run(fmt.Sprintf("0.0.0.0:%s", config.Port))
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		//origin := c.Request.Header.Get("Origin")
		//if origin != "" {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization,X-API-KEY,X-Signature,X-Public-Key")
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
