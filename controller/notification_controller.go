package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"ordbook-aggregation/controller/request"
	"ordbook-aggregation/controller/respond"
	"ordbook-aggregation/service/order_brc20_service"
	"ordbook-aggregation/tool"
	"strconv"
)

// @Summary Fetch address notification
// @Description Fetch address notification
// @Produce  json
// @Tags brc20
// @Param address query string true "address"
// @Success 200 {object} respond.NotificationResponse ""
// @Router /brc20/common/notification/address [get]
func FetchAddressNotification(c *gin.Context) {
	var (
		t   int64                              = tool.MakeTimestamp()
		req *request.OrderNotificationFetchReq = &request.OrderNotificationFetchReq{
			Address: c.DefaultQuery("address", ""),
		}
	)
	responseModel, err := order_brc20_service.FetchAddressNotification(req)
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
	return
}

// @Summary Clear address notification
// @Description Clear address notification
// @Produce  json
// @Tags brc20
// @Param address query string true "address"
// @Param notificationType query int false "notificationType 0-all, 1-poolUsed, 2-bidInvalid, 3-orderFinish"
// @Success 200 {object} respond.Message ""
// @Router /brc20/common/notification/clear [get]
func ClearAllNotification(c *gin.Context) {
	var (
		t                   int64                              = tool.MakeTimestamp()
		notificationTypeStr string                             = c.DefaultQuery("notificationType", "0")
		req                 *request.OrderNotificationFetchReq = &request.OrderNotificationFetchReq{
			Address:          c.DefaultQuery("address", ""),
			NotificationType: 0,
		}
	)
	req.NotificationType, _ = strconv.ParseInt(notificationTypeStr, 10, 64)
	responseModel, err := order_brc20_service.ClearAllNotification(req)
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
	return
}
