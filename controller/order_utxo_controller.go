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

// @Summary Cold down the utxo
// @Description Do bid order
// @Produce  json
// @Param Request body request.ColdDownUtxo true "Request"
// @Tags System
// @Success 200 {object} respond.Message ""
// @Router /brc20/utxo/colddown [post]
func ColdDownUtxo(c *gin.Context) {
	var (
		t   int64            = tool.MakeTimestamp()
		requestModel *request.ColdDownUtxo
	)
	if c.ShouldBindJSON(&requestModel) == nil {
		responseModel, err := order_brc20_service.ColdDownUtxo(requestModel)
		if err != nil {
			c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
			return
		}
		c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
		return
	}
	c.JSONP(http.StatusInternalServerError, respond.RespErr(errors.New("error parameter"), tool.MakeTimestamp()-t, respond.HttpsCodeError))
}