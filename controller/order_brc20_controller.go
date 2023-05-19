package controller

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"ordbook-aggregation/controller/request"
	"ordbook-aggregation/controller/respond"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/order_brc20_service"
	"ordbook-aggregation/tool"
	"strconv"
)

// @Summary Push order
// @Description Push order
// @Produce  json
// @Param Request body request.OrderBrc20PushReq true "Request"
// @Tags brc20
// @Success 200 {object} respond.Message ""
// @Router /brc20/order/push [post]
func PushOrder(c *gin.Context) {
	var (
		t   int64            = tool.MakeTimestamp()
		requestModel *request.OrderBrc20PushReq
	)
	if c.ShouldBindJSON(&requestModel) == nil {
		responseModel, err := order_brc20_service.PushOrder(requestModel)
		if err != nil {
			c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
			return
		}
		c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
		return
	}
	c.JSONP(http.StatusInternalServerError, respond.RespErr(errors.New("error parameter"), tool.MakeTimestamp()-t, respond.HttpsCodeError))
}

// @Summary Fetch orders
// @Description Fetch orders
// @Produce  json
// @Tags brc20
// @Param tick query string false "tick"
// @Param sellerAddress query string false "sellerAddress"
// @Param buyerAddress query string false "buyerAddress"
// @Param orderState query int false "orderState"
// @Param orderType query int false "orderType"
// @Param limit query int false "limit"
// @Param flag query int false "flag"
// @Param sortKey query string false "sortKey"
// @Param sortType query int false "sortType"
// @Success 200 {object} respond.OrderResponse ""
// @Router /brc20/orders [get]
func FetchOrders(c *gin.Context) {
	var (
		t   int64            = tool.MakeTimestamp()
		orderStateStr = c.DefaultQuery("orderState", "0")
		orderTypeStr = c.DefaultQuery("orderType", "0")
		limitStr = c.DefaultQuery("limit", "50")
		flagStr = c.DefaultQuery("flag", "0")
		sortTypeStr = c.DefaultQuery("sortType", "0")
		req *request.OrderBrc20FetchReq = &request.OrderBrc20FetchReq{
			Tick:          c.DefaultQuery("tick", ""),
			OrderState:    0,
			OrderType:     0,
			Limit:         0,
			Flag:          0,
			SellerAddress: c.DefaultQuery("sellerAddress", ""),
			BuyerAddress:  c.DefaultQuery("buyerAddress", ""),
			SortKey:       c.DefaultQuery("sortKey", ""),
			SortType:      0,
		}
	)
	orderState, _ := strconv.ParseInt(orderStateStr, 10, 64)
	orderType, _ := strconv.ParseInt(orderTypeStr, 10, 64)
	req.OrderState = model.OrderState(orderState)
	req.OrderType = model.OrderType(orderType)
	req.Limit, _ = strconv.ParseInt(limitStr, 10, 64)
	req.Flag, _ = strconv.ParseInt(flagStr, 10, 64)
	req.SortType, _ = strconv.ParseInt(sortTypeStr, 10, 64)
	responseModel, err := order_brc20_service.FetchOrders(req)
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
	return
}
