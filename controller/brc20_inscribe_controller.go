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

// @Summary Pre inscribe
// @Description Pre inscribe
// @Produce  json
// @Param Request body request.Brc20PreReq true "Request"
// @Tags brc20
// @Success 200 {object} respond.Message ""
// @Router /brc20/inscribe/pre [post]
func PreInscribe(c *gin.Context) {
	var (
		t   int64            = tool.MakeTimestamp()
		requestModel *request.Brc20PreReq
	)
	if c.ShouldBindJSON(&requestModel) == nil {
		responseModel, err := order_brc20_service.PreInscribe(requestModel)
		if err != nil {
			c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
			return
		}
		c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
		return
	}
	c.JSONP(http.StatusInternalServerError, respond.RespErr(errors.New("error parameter"), tool.MakeTimestamp()-t, respond.HttpsCodeError))
}

// @Summary Commit inscribe
// @Description Commit inscribe
// @Produce  json
// @Param Request body request.Brc20CommitReq true "Request"
// @Tags brc20
// @Success 200 {object} respond.Message ""
// @Router /brc20/inscribe/commit [post]
func CommitInscribe(c *gin.Context) {
	var (
		t            int64 = tool.MakeTimestamp()
		requestModel *request.Brc20CommitReq
	)
	if c.ShouldBindJSON(&requestModel) == nil {
		responseModel, err := order_brc20_service.CommitInscribe(requestModel)
		if err != nil {
			c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
			return
		}
		c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
		return
	}
	c.JSONP(http.StatusInternalServerError, respond.RespErr(errors.New("error parameter"), tool.MakeTimestamp()-t, respond.HttpsCodeError))
}