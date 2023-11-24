package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"ordbook-aggregation/controller/respond"
	"ordbook-aggregation/service/rate_service"
	"ordbook-aggregation/tool"
)

// @Summary Fetch rate
// @Description Fetch rate
// @Produce  json
// @Tags brc20
// @Success 200 {object} respond.RateResp ""
// @Router /brc20/common/rate/btc [get]
func FetchRate(c *gin.Context) {
	var (
		t int64 = tool.MakeTimestamp()
	)
	responseModel, err := rate_service.FetchRate()
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
	return
}
