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

// @Summary Fetch pool key
// @Description Fetch pool key
// @Produce  json
// @Tags brc20
// @Param net query string true "net"
// @Success 200 {object} respond.PoolKeyInfoResp ""
// @Router /brc20/pool/pair/key [get]
func FetchPoolPlatformPublicKey(c *gin.Context) {
	var (
		t   int64                     = tool.MakeTimestamp()
		req *request.PoolBrc20PushReq = &request.PoolBrc20PushReq{
			Net: c.DefaultQuery("net", ""),
		}
	)
	responseModel, err := order_brc20_service.FetchPoolPlatformPublicKey(req)
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
	return
}

// @Summary Push pool order
// @Description Push pool order
// @Produce  json
// @Param Request body request.PoolBrc20PushReq true "Request"
// @Tags brc20
// @Success 200 {object} respond.Message ""
// @Router /brc20/pool/order/push [post]
func PushPoolOrder(c *gin.Context) {
	var (
		t            int64 = tool.MakeTimestamp()
		requestModel *request.PoolBrc20PushReq
		publicKey    string = ""
	)
	if c.ShouldBindJSON(&requestModel) == nil {
		publicKey = getAuthParams(c)
		responseModel, err := order_brc20_service.PushPoolOrder(requestModel, publicKey)
		if err != nil {
			c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
			return
		}
		c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
		return
	}
	c.JSONP(http.StatusInternalServerError, respond.RespErr(errors.New("error parameter"), tool.MakeTimestamp()-t, respond.HttpsCodeError))
}

// @Summary Fetch pool orders
// @Description Fetch pool orders
// @Produce  json
// @Tags brc20
// @Param net query string false "net:mainnet/signet/testnet"
// @Param tick query string false "tick"
// @Param address query string false "address"
// @Param poolState query int false "poolState: 1-add,2-remove,3-used,4-claim"
// @Param poolType query int false "poolType: 1-tick,2-btc"
// @Param limit query int false "limit: Max-50"
// @Param page query int false "page"
// @Param flag query int false "flag"
// @Param sortKey query string false "sortKey: timestamp/coinRatePrice, default:timestamp"
// @Param sortType query int false "sortType: 1/-1"
// @Success 200 {object} respond.PoolResponse ""
// @Router /brc20/pool/orders [get]
func FetchPoolOrders(c *gin.Context) {
	var (
		t            int64                      = tool.MakeTimestamp()
		poolStateStr                            = c.DefaultQuery("poolState", "0")
		poolTypeStr                             = c.DefaultQuery("poolType", "0")
		limitStr                                = c.DefaultQuery("limit", "50")
		pageStr                                 = c.DefaultQuery("page", "0")
		flagStr                                 = c.DefaultQuery("flag", "0")
		sortTypeStr                             = c.DefaultQuery("sortType", "0")
		req          *request.PoolBrc20FetchReq = &request.PoolBrc20FetchReq{
			Net:       c.DefaultQuery("net", ""),
			Tick:      c.DefaultQuery("tick", ""),
			PoolState: 0,
			PoolType:  0,
			Limit:     0,
			Flag:      0,
			Page:      0,
			Address:   c.DefaultQuery("address", ""),
			SortKey:   c.DefaultQuery("sortKey", ""),
			SortType:  0,
		}
	)
	poolState, _ := strconv.ParseInt(poolStateStr, 10, 64)
	poolType, _ := strconv.ParseInt(poolTypeStr, 10, 64)
	req.PoolState = model.PoolState(poolState)
	req.PoolType = model.PoolType(poolType)
	req.Limit, _ = strconv.ParseInt(limitStr, 10, 64)
	req.Flag, _ = strconv.ParseInt(flagStr, 10, 64)
	req.Page, _ = strconv.ParseInt(pageStr, 10, 64)
	req.SortType, _ = strconv.ParseInt(sortTypeStr, 10, 64)
	responseModel, err := order_brc20_service.FetchPoolOrders(req)
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
	return
}

// @Summary Fetch one pool order
// @Description Fetch one pool order
// @Produce  json
// @Tags brc20
// @Param orderId path string true "orderId"
// @Param address query string true "address"
// @Success 200 {object} respond.PoolBrc20Item ""
// @Router /brc20/pool/order/{orderId} [get]
func FetchOnePoolOrder(c *gin.Context) {
	var (
		t   int64                         = tool.MakeTimestamp()
		req *request.PoolBrc20FetchOneReq = &request.PoolBrc20FetchOneReq{
			Net:     c.DefaultQuery("net", ""),
			Tick:    c.DefaultQuery("tick", ""),
			OrderId: c.Param("orderId"),
			Address: c.DefaultQuery("address", ""),
		}
		publicKey = getAuthParams(c)
	)
	responseModel, err := order_brc20_service.FetchOnePoolOrder(req, publicKey, c.ClientIP())
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
	return
}

// @Summary Fetch pool pair info
// @Description Fetch pool pair info
// @Produce  json
// @Tags brc20
// @Param net query string false "net"
// @Param tick query string false "tick"
// @Param pair query string false "pair"
// @Success 200 {object} respond.PoolInfoResponse ""
// @Router /brc20/pool/pair/info [get]
func FetchPoolPairInfo(c *gin.Context) {
	var (
		t   int64                        = tool.MakeTimestamp()
		req *request.PoolPairFetchOneReq = &request.PoolPairFetchOneReq{
			Net:  c.DefaultQuery("net", ""),
			Tick: c.DefaultQuery("tick", ""),
			Pair: c.DefaultQuery("pair", ""),
		}
	)
	responseModel, err := order_brc20_service.FetchPoolPairInfo(req)
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
	return
}
