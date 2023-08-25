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
		t            int64 = tool.MakeTimestamp()
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

// @Summary Cold down the brc20 transfer
// @Description Cold down the brc20 transfer
// @Produce  json
// @Param Request body request.ColdDownBrcTransfer true "Request"
// @Tags System
// @Success 200 {object} respond.Brc20TransferCommitResp ""
// @Router /brc20/transfer/colddown [post]
func ColdDownBrc20Transfer(c *gin.Context) {
	var (
		t            int64 = tool.MakeTimestamp()
		requestModel *request.ColdDownBrcTransfer
	)
	if c.ShouldBindJSON(&requestModel) == nil {
		responseModel, err := order_brc20_service.ColdDownBrc20Transfer(requestModel)
		if err != nil {
			c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
			return
		}
		c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
		return
	}
	c.JSONP(http.StatusInternalServerError, respond.RespErr(errors.New("error parameter"), tool.MakeTimestamp()-t, respond.HttpsCodeError))
}

// @Summary Cold down the brc20 transfer batch
// @Description Cold down the brc20 transfer batch
// @Produce  json
// @Param Request body request.ColdDownBrcTransferBatch true "Request"
// @Tags System
// @Success 200 {object} respond.Brc20TransferCommitBatchResp ""
// @Router /brc20/transfer/colddown/batch [post]
func ColdDownBrc20TransferBatch(c *gin.Context) {
	var (
		t            int64 = tool.MakeTimestamp()
		requestModel *request.ColdDownBrcTransferBatch
	)
	if c.ShouldBindJSON(&requestModel) == nil {
		responseModel, err := order_brc20_service.ColdDownBrc20TransferBatch(requestModel)
		if err != nil {
			c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
			return
		}
		c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
		return
	}
	c.JSONP(http.StatusInternalServerError, respond.RespErr(errors.New("error parameter"), tool.MakeTimestamp()-t, respond.HttpsCodeError))
}

// @Summary Cold down the brc20 transfer batch
// @Description Cold down the brc20 transfer batch
// @Produce  json
// @Param Request body request.ColdDownBrcTransferBatch true "Request"
// @Tags System
// @Success 200 {object} respond.Brc20TransferCommitBatchResp ""
// @Router /brc20/transfer/colddown/batch/ask [post]
func ColdDownBatchBrc20TransferAndMakeAsk(c *gin.Context) {
	var (
		t            int64 = tool.MakeTimestamp()
		requestModel *request.ColdDownBrcTransferBatch
	)
	if c.ShouldBindJSON(&requestModel) == nil {
		responseModel, err := order_brc20_service.ColdDownBatchBrc20TransferAndMakeAsk(requestModel)
		if err != nil {
			c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
			return
		}
		c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
		return
	}
	c.JSONP(http.StatusInternalServerError, respond.RespErr(errors.New("error parameter"), tool.MakeTimestamp()-t, respond.HttpsCodeError))
}

// @Summary Cold down the brc20 transfer batch
// @Description Cold down the brc20 transfer batch
// @Produce  json
// @Param Request body request.ColdDownBrcTransferBatch true "Request"
// @Tags System
// @Success 200 {object} respond.Brc20TransferCommitBatchResp ""
// @Router /brc20/transfer/colddown/batch/pool [post]
func ColdDownBatchBrc20TransferAndMakePool(c *gin.Context) {
	var (
		t            int64 = tool.MakeTimestamp()
		requestModel *request.ColdDownBrcTransferBatch
	)
	if c.ShouldBindJSON(&requestModel) == nil {
		responseModel, err := order_brc20_service.ColdDownBatchBrc20TransferAndMakePool(requestModel)
		if err != nil {
			c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
			return
		}
		c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
		return
	}
	c.JSONP(http.StatusInternalServerError, respond.RespErr(errors.New("error parameter"), tool.MakeTimestamp()-t, respond.HttpsCodeError))
}
