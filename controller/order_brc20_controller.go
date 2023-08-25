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
// @Router /brc20/order/ask/push [post]
func PushOrder(c *gin.Context) {
	var (
		t            int64 = tool.MakeTimestamp()
		requestModel *request.OrderBrc20PushReq
		publicKey    string = ""
	)
	if c.ShouldBindJSON(&requestModel) == nil {
		publicKey = getAuthParams(c)
		responseModel, err := order_brc20_service.PushOrder(requestModel, publicKey)
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
// @Param net query string false "net:mainnet/signet/testnet"
// @Param tick query string false "tick"
// @Param sellerAddress query string false "sellerAddress"
// @Param buyerAddress query string false "buyerAddress"
// @Param orderState query int false "orderState: 1-create,2-finish,3-cancel,5-timeout,6-err,100-all"
// @Param orderType query int false "orderType: 1-sell,2-buy"
// @Param limit query int false "limit: Max-50"
// @Param page query int false "page"
// @Param flag query int false "flag"
// @Param sortKey query string false "sortKey: timestamp/coinRatePrice, default:timestamp"
// @Param sortType query int false "sortType: 1/-1"
// @Success 200 {object} respond.OrderResponse ""
// @Router /brc20/orders [get]
func FetchOrders(c *gin.Context) {
	var (
		t             int64                       = tool.MakeTimestamp()
		orderStateStr                             = c.DefaultQuery("orderState", "0")
		orderTypeStr                              = c.DefaultQuery("orderType", "0")
		limitStr                                  = c.DefaultQuery("limit", "50")
		pageStr                                   = c.DefaultQuery("page", "0")
		flagStr                                   = c.DefaultQuery("flag", "0")
		sortTypeStr                               = c.DefaultQuery("sortType", "0")
		req           *request.OrderBrc20FetchReq = &request.OrderBrc20FetchReq{
			Net:           c.DefaultQuery("net", ""),
			Tick:          c.DefaultQuery("tick", ""),
			OrderState:    0,
			OrderType:     0,
			Limit:         0,
			Flag:          0,
			Page:          0,
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
	req.Page, _ = strconv.ParseInt(pageStr, 10, 64)
	req.SortType, _ = strconv.ParseInt(sortTypeStr, 10, 64)
	responseModel, err := order_brc20_service.FetchOrders(req)
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
	return
}

// @Summary Fetch one order
// @Description Fetch one order
// @Produce  json
// @Tags brc20
// @Param orderId path string true "orderId"
// @Param buyerAddress query string true "buyerAddress"
// @Success 200 {object} respond.Brc20Item ""
// @Router /brc20/order/{orderId} [get]
func FetchOneOrder(c *gin.Context) {
	var (
		t   int64                          = tool.MakeTimestamp()
		req *request.OrderBrc20FetchOneReq = &request.OrderBrc20FetchOneReq{
			Net:          c.DefaultQuery("net", ""),
			Tick:         c.DefaultQuery("tick", ""),
			OrderId:      c.Param("orderId"),
			BuyerAddress: c.DefaultQuery("buyerAddress", ""),
		}
		publicKey = getAuthParams(c)
	)
	responseModel, err := order_brc20_service.FetchOneOrders(req, publicKey, c.ClientIP())
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
	return
}

// @Summary Fetch user orders
// @Description Fetch user orders
// @Produce  json
// @Tags brc20
// @Param net query string false "net:mainnet/signet/testnet"
// @Param tick query string false "tick"
// @Param address path string true "address"
// @Param orderState query int false "orderState: 1-create,2-finish,3-cancel,5-timeout,6-err,100-all"
// @Param orderType query int false "orderType: 1-sell,2-buy"
// @Param limit query int false "limit: Max-50"
// @Param flag query int false "flag"
// @Param page query int false "page"
// @Param sortKey query string false "sortKey: timestamp/coinRatePrice, default:timestamp"
// @Param sortType query int false "sortType: 1/-1"
// @Success 200 {object} respond.OrderResponse ""
// @Router /brc20/orders/user/{address} [get]
func FetchUserOrders(c *gin.Context) {
	var (
		t             int64                         = tool.MakeTimestamp()
		orderStateStr                               = c.DefaultQuery("orderState", "0")
		orderTypeStr                                = c.DefaultQuery("orderType", "0")
		limitStr                                    = c.DefaultQuery("limit", "50")
		flagStr                                     = c.DefaultQuery("flag", "0")
		pageStr                                     = c.DefaultQuery("page", "0")
		sortTypeStr                                 = c.DefaultQuery("sortType", "0")
		req           *request.Brc20OrderAddressReq = &request.Brc20OrderAddressReq{
			Net:        c.DefaultQuery("net", ""),
			Tick:       c.DefaultQuery("tick", ""),
			OrderState: 0,
			OrderType:  0,
			Limit:      0,
			Flag:       0,
			Page:       0,
			Address:    c.Param("address"),
			SortKey:    c.DefaultQuery("sortKey", ""),
			SortType:   0,
		}
	)
	orderState, _ := strconv.ParseInt(orderStateStr, 10, 64)
	orderType, _ := strconv.ParseInt(orderTypeStr, 10, 64)
	req.OrderState = model.OrderState(orderState)
	req.OrderType = model.OrderType(orderType)
	req.Limit, _ = strconv.ParseInt(limitStr, 10, 64)
	req.Flag, _ = strconv.ParseInt(flagStr, 10, 64)
	req.Page, _ = strconv.ParseInt(pageStr, 10, 64)
	req.SortType, _ = strconv.ParseInt(sortTypeStr, 10, 64)
	responseModel, err := order_brc20_service.FetchUserOrders(req)
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
	return
}

// @Summary Fetch tick info
// @Description Fetch tick info
// @Produce  json
// @Tags brc20
// @Param net query string false "net"
// @Param tick query string false "tick"
// @Success 200 {object} respond.Brc20TickInfoResponse ""
// @Router /brc20/tickers [get]
func FetchTicker(c *gin.Context) {
	var (
		t   int64                      = tool.MakeTimestamp()
		req *request.TickBrc20FetchReq = &request.TickBrc20FetchReq{
			Net:      c.DefaultQuery("net", ""),
			Tick:     c.DefaultQuery("tick", ""),
			Limit:    0,
			Flag:     0,
			SortKey:  c.DefaultQuery("sortKey", ""),
			SortType: 0,
		}
	)

	resp, err := order_brc20_service.FetchTickers(req)
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(resp, tool.MakeTimestamp()-t))
	return
}

// @Summary Fetch tick kline data
// @Description Fetch tick kline data
// @Produce  json
// @Tags brc20
// @Param net query string true "net"
// @Param tick query string true "tick"
// @Param interval query string false "interval：15m/1h/4h/1d/1w/"
// @Param limit query int false "limit"
// @Param flag query int false "flag"
// @Success 200 {object} respond.KlineItem ""
// @Router /brc20/kline [get]
func FetchKline(c *gin.Context) {
	var (
		t   int64                      = tool.MakeTimestamp()
		req *request.TickKlineFetchReq = &request.TickKlineFetchReq{
			Net:      c.DefaultQuery("net", "livenet"),
			Tick:     c.DefaultQuery("tick", ""),
			Limit:    0,
			Interval: c.DefaultQuery("interval", "15m"),
		}
	)
	resp, err := order_brc20_service.FetchTickKline(req)
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(resp, tool.MakeTimestamp()-t))
	return
}

// @Summary Update order
// @Description Update order
// @Produce  json
// @Param Request body request.OrderBrc20UpdateReq true "Request"
// @Tags brc20
// @Success 200 {object} respond.Message ""
// @Router /brc20/order/update [post]
func UpdateOrder(c *gin.Context) {
	var (
		t            int64 = tool.MakeTimestamp()
		requestModel *request.OrderBrc20UpdateReq
		publicKey    string = ""
	)
	if c.ShouldBindJSON(&requestModel) == nil {
		publicKey = getAuthParams(c)
		responseModel, err := order_brc20_service.UpdateOrder(requestModel, publicKey, c.ClientIP())
		if err != nil {
			c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
			return
		}
		c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
		return
	}
	c.JSONP(http.StatusInternalServerError, respond.RespErr(errors.New("error parameter"), tool.MakeTimestamp()-t, respond.HttpsCodeError))
}

// @Summary Fetch pre bid
// @Description Fetch pre bid
// @Produce  json
// @Tags brc20
// @Param net query string false "net:mainnet/signet/testnet"
// @Param tick query string false "tick"
// @Param limit query int false "limit: Max-50"
// @Param flag query int false "flag"
// @Param isPool query bool false "isPool"
// @Success 200 {object} respond.OrderResponse ""
// @Router /brc20/order/bid/pre [get]
func FetchPreBid(c *gin.Context) {
	var (
		t         int64                        = tool.MakeTimestamp()
		limitStr                               = c.DefaultQuery("limit", "50")
		pageStr                                = c.DefaultQuery("page", "0")
		isPoolStr                              = c.DefaultQuery("isPool", "false")
		req       *request.OrderBrc20GetBidReq = &request.OrderBrc20GetBidReq{
			Net:  c.DefaultQuery("net", ""),
			Tick: c.DefaultQuery("tick", ""),
		}
	)
	req.Limit, _ = strconv.ParseInt(limitStr, 10, 64)
	req.Page, _ = strconv.ParseInt(pageStr, 10, 64)
	req.IsPool, _ = strconv.ParseBool(isPoolStr)
	responseModel, err := order_brc20_service.FetchPreBid(req)
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
	return
}

// @Summary Fetch bid
// @Description Fetch bid
// @Produce  json
// @Tags brc20
// @Param net query string false "net:mainnet/signet/testnet"
// @Param tick query string false "tick"
// @Param inscriptionId query string false "inscriptionId"
// @Param inscriptionNumber query string false "inscriptionNumber"
// @Param coinAmount query string false "coinAmount"
// @Param address query string false "address"
// @Param amount query int false "amount"
// @Param isPool query bool false "isPool for pool"
// @Param orderId query string false "orderId of pool"
// @Success 200 {object} respond.BidPsbt ""
// @Router /brc20/order/bid [get]
func FetchBidPsbt(c *gin.Context) {
	var (
		t         int64                        = tool.MakeTimestamp()
		amountStr string                       = c.DefaultQuery("amount", "0")
		isPoolStr                              = c.DefaultQuery("isPool", "false")
		req       *request.OrderBrc20GetBidReq = &request.OrderBrc20GetBidReq{
			Net:               c.DefaultQuery("net", ""),
			Tick:              c.DefaultQuery("tick", ""),
			InscriptionId:     c.DefaultQuery("inscriptionId", ""),
			InscriptionNumber: c.DefaultQuery("inscriptionNumber", ""),
			CoinAmount:        c.DefaultQuery("coinAmount", "0"),
			Address:           c.DefaultQuery("address", ""),
			PoolOrderId:       c.DefaultQuery("poolOrderId", ""),
		}
	)
	req.Amount, _ = strconv.ParseUint(amountStr, 10, 64)
	req.IsPool, _ = strconv.ParseBool(isPoolStr)
	responseModel, err := order_brc20_service.FetchBidPsbt(req)
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
	return
}

// @Summary Push bid order
// @Description Push bid order
// @Produce  json
// @Param Request body request.OrderBrc20UpdateBidReq true "Request"
// @Tags brc20
// @Success 200 {object} respond.Message ""
// @Router /brc20/order/bid/push [post]
func UpdateBidPsbt(c *gin.Context) {
	var (
		t            int64 = tool.MakeTimestamp()
		requestModel *request.OrderBrc20UpdateBidReq
	)
	if c.ShouldBindJSON(&requestModel) == nil {
		responseModel, err := order_brc20_service.UpdateBidPsbt(requestModel)
		if err != nil {
			c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
			return
		}
		c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
		return
	}
	c.JSONP(http.StatusInternalServerError, respond.RespErr(errors.New("error parameter"), tool.MakeTimestamp()-t, respond.HttpsCodeError))
}

// @Summary Do bid order
// @Description Do bid order
// @Produce  json
// @Param Request body request.OrderBrc20DoBidReq true "Request"
// @Tags brc20
// @Success 200 {object} respond.Message ""
// @Router /brc20/order/bid/do [post]
func DoBid(c *gin.Context) {
	var (
		t            int64 = tool.MakeTimestamp()
		requestModel *request.OrderBrc20DoBidReq
	)
	if c.ShouldBindJSON(&requestModel) == nil {
		responseModel, err := order_brc20_service.DoBid(requestModel)
		if err != nil {
			c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
			return
		}
		c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
		return
	}
	c.JSONP(http.StatusInternalServerError, respond.RespErr(errors.New("error parameter"), tool.MakeTimestamp()-t, respond.HttpsCodeError))
}

// @Summary Fetch uuid
// @Description Fetch uuid
// @Produce  json
// @Tags brc20
// @Success 200 {object} respond.WsUuidResp ""
// @Router /brc20/ws/uuid [get]
func GetWsUuid(c *gin.Context) {
	var (
		t int64 = tool.MakeTimestamp()
	)

	resp, err := order_brc20_service.GetWsUuid(c.ClientIP())
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(resp, tool.MakeTimestamp()-t))
	return
}

// @Summary Check inscription brc20 valid
// @Description Check inscription brc20 valid
// @Produce  json
// @Param inscriptionId query string false "inscriptionId"
// @Param inscriptionNumber query string false "inscriptionNumber"
// @Tags brc20
// @Success 200 {object} respond.CheckBrc20InscriptionReq ""
// @Router /brc20/check/info [get]
func CheckBrc20(c *gin.Context) {
	var (
		t   int64                             = tool.MakeTimestamp()
		req *request.CheckBrc20InscriptionReq = &request.CheckBrc20InscriptionReq{
			InscriptionId:     c.DefaultQuery("inscriptionId", ""),
			InscriptionNumber: c.DefaultQuery("inscriptionNumber", ""),
		}
	)
	resp, err := order_brc20_service.CheckBrc20(req)
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(resp, tool.MakeTimestamp()-t))
	return
}

// @Summary Check inscription brc20 valid
// @Description Check inscription brc20 valid
// @Produce  json
// @Param tick path string true "tick"
// @Param address path string true "address"
// @Param net query string false "net"
// @Param page query int false "page"
// @Param limit query int false "limit"
// @Tags brc20
// @Success 200 {object} respond.BalanceDetails ""
// @Router /brc20/address/{address}/{tick} [get]
func GetBrc20BalanceDetail(c *gin.Context) {
	var (
		t        int64                    = tool.MakeTimestamp()
		pageStr  string                   = c.DefaultQuery("page", "1")
		limitStr string                   = c.DefaultQuery("limit", "60")
		req      *request.Brc20AddressReq = &request.Brc20AddressReq{
			Net:     c.DefaultQuery("net", ""),
			Tick:    c.Param("tick"),
			Address: c.Param("address"),
		}
	)
	req.Page, _ = strconv.ParseInt(pageStr, 10, 64)
	req.Limit, _ = strconv.ParseInt(limitStr, 10, 64)
	resp, err := order_brc20_service.GetBrc20BalanceDetail(req)
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(resp, tool.MakeTimestamp()-t))
	return
}

// @Summary Get brc20 balance
// @Description Get brc20 balance
// @Produce  json
// @Param tick query string false "tick"
// @Param address path string true "address"
// @Param net query string false "net"
// @Param page query int false "page"
// @Param limit query int false "limit"
// @Tags brc20
// @Success 200 {object} respond.Brc20BalanceList ""
// @Router /brc20/address/{address}/balance/info [get]
func GetBrc20BalanceList(c *gin.Context) {
	var (
		t        int64                    = tool.MakeTimestamp()
		pageStr  string                   = c.DefaultQuery("page", "1")
		limitStr string                   = c.DefaultQuery("limit", "60")
		req      *request.Brc20AddressReq = &request.Brc20AddressReq{
			Net:     c.DefaultQuery("net", ""),
			Tick:    c.DefaultQuery("tick", ""),
			Address: c.Param("address"),
		}
	)
	req.Page, _ = strconv.ParseInt(pageStr, 10, 64)
	req.Limit, _ = strconv.ParseInt(limitStr, 10, 64)
	resp, err := order_brc20_service.GetBrc20BalanceList(req)
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(resp, tool.MakeTimestamp()-t))
	return
}

// @Summary Get bid dummy list
// @Description Get bid dummy list
// @Produce  json
// @Param address path string true "address"
// @Param net query string false "net"
// @Param skip query int false "skip"
// @Param limit query int false "limit"
// @Tags brc20
// @Success 200 {object} respond.Brc20BidDummyResponse ""
// @Router /brc20/order/bid/dummy/{address} [get]
func GetBidDummyList(c *gin.Context) {
	var (
		t        int64                            = tool.MakeTimestamp()
		skipStr  string                           = c.DefaultQuery("skip", "1")
		limitStr string                           = c.DefaultQuery("limit", "60")
		req      *request.Brc20BidAddressDummyReq = &request.Brc20BidAddressDummyReq{
			Net:     c.DefaultQuery("net", ""),
			Tick:    c.Param("tick"),
			Address: c.Param("address"),
		}
	)
	req.Skip, _ = strconv.ParseInt(skipStr, 10, 64)
	req.Limit, _ = strconv.ParseInt(limitStr, 10, 64)
	resp, err := order_brc20_service.GetBidDummyList(req)
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(resp, tool.MakeTimestamp()-t))
	return
}
