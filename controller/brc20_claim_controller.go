package controller

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"ordbook-aggregation/controller/request"
	"ordbook-aggregation/controller/respond"
	"ordbook-aggregation/service/order_brc20_service"
	"ordbook-aggregation/tool"
)

// @Summary Fetch one claim order
// @Description Fetch one claim order
// @Produce  json
// @Tags brc20
// @Param net query string true "net"
// @Param tick query string true "tick"
// @Param address query string true "address"
// @Success 200 {object} respond.Brc20ClaimItem ""
// @Router /brc20/claim/order [get]
func FetchClaimOrder(c *gin.Context) {
	var (
		t   int64                               = tool.MakeTimestamp()
		req *request.OrderBrc20ClaimFetchOneReq = &request.OrderBrc20ClaimFetchOneReq{
			Net:     c.DefaultQuery("net", ""),
			Tick:    c.DefaultQuery("tick", ""),
			Address: c.DefaultQuery("address", ""),
		}
		publicKey = getAuthParams(c)
	)
	responseModel, err := order_brc20_service.FetchClaimOrder(req, publicKey, c.ClientIP())
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
	return
}

// @Summary Update claim order
// @Description Update claim order
// @Produce  json
// @Param Request body request.OrderBrc20ClaimUpdateReq true "Request"
// @Tags brc20
// @Success 200 {object} respond.Message ""
// @Router /brc20/claim/order/update [post]
func UpdateClaimOrder(c *gin.Context) {
	var (
		t            int64 = tool.MakeTimestamp()
		requestModel *request.OrderBrc20ClaimUpdateReq
		publicKey    string = ""
	)
	if c.ShouldBindJSON(&requestModel) == nil {
		publicKey = getAuthParams(c)
		responseModel, err := order_brc20_service.UpdateClaimOrder(requestModel, publicKey, c.ClientIP())
		if err != nil {
			c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
			return
		}
		c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
		return
	}
	c.JSONP(http.StatusInternalServerError, respond.RespErr(errors.New("error parameter"), tool.MakeTimestamp()-t, respond.HttpsCodeError))
}
