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

// @Summary Update pool order
// @Description Update pool order
// @Produce  json
// @Param Request body request.OrderPoolBrc20UpdateReq true "Request"
// @Tags brc20
// @Success 200 {object} respond.Message ""
// @Router /brc20/pool/order/update [post]
func UpdatePoolOrder(c *gin.Context) {
	var (
		t            int64 = tool.MakeTimestamp()
		requestModel *request.OrderPoolBrc20UpdateReq
		publicKey    string = ""
	)
	if c.ShouldBindJSON(&requestModel) == nil {
		publicKey = getAuthParams(c)
		responseModel, err := order_brc20_service.UpdatePoolOrder(requestModel, publicKey, c.ClientIP())
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

// @Summary Fetch one pool pair info
// @Description Fetch one pool pair info
// @Produce  json
// @Tags brc20
// @Param net query string false "net"
// @Param tick query string false "tick"
// @Param pair query string false "pair"
// @Param address query string false "address"
// @Success 200 {object} respond.PoolInfoItem ""
// @Router /brc20/pool/pair/info/one [get]
func FetchOnePoolPairInfo(c *gin.Context) {
	var (
		t   int64                        = tool.MakeTimestamp()
		req *request.PoolPairFetchOneReq = &request.PoolPairFetchOneReq{
			Net:     c.DefaultQuery("net", ""),
			Tick:    c.DefaultQuery("tick", ""),
			Pair:    c.DefaultQuery("pair", ""),
			Address: c.DefaultQuery("address", ""),
		}
	)
	responseModel, err := order_brc20_service.FetchOnePoolPairInfo(req)
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
	return
}

// @Summary Fetch pool inscription
// @Description Fetch pool inscription
// @Produce  json
// @Tags brc20
// @Param net query string false "net"
// @Param tick query string false "tick"
// @Param address query string true "address"
// @Success 200 {object} respond.PoolInscriptionResp ""
// @Router /brc20/pool/inscription [get]
func FetchPoolInscription(c *gin.Context) {
	var (
		t   int64                                 = tool.MakeTimestamp()
		req *request.PoolBrc20FetchInscriptionReq = &request.PoolBrc20FetchInscriptionReq{
			Net:     c.DefaultQuery("net", ""),
			Tick:    c.DefaultQuery("tick", ""),
			Address: c.DefaultQuery("address", ""),
		}
		publicKey = getAuthParams(c)
	)
	responseModel, err := order_brc20_service.FetchPoolInscription(req, publicKey, c.ClientIP())
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
	return
}

// @Summary get claim pool order
// @Description get claim pool order
// @Produce  json
// @Param Request body request.PoolBrc20ClaimReq true "Request"
// @Tags brc20
// @Success 200 {object} respond.PoolBrc20ClaimResp ""
// @Router /brc20/pool/order/claim [post]
func ClaimPool(c *gin.Context) {
	var (
		t            int64 = tool.MakeTimestamp()
		requestModel *request.PoolBrc20ClaimReq
		publicKey    string = ""
	)
	if c.ShouldBindJSON(&requestModel) == nil {
		publicKey = getAuthParams(c)
		responseModel, err := order_brc20_service.ClaimPool(requestModel, publicKey, c.ClientIP())
		if err != nil {
			c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
			return
		}
		c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
		return
	}
	c.JSONP(http.StatusInternalServerError, respond.RespErr(errors.New("error parameter"), tool.MakeTimestamp()-t, respond.HttpsCodeError))
}

// @Summary claim pool order
// @Description claim pool order
// @Produce  json
// @Param Request body request.PoolBrc20ClaimUpdateReq true "Request"
// @Tags brc20
// @Success 200 {object} respond.Message ""
// @Router /brc20/pool/order/claim/commit [post]
func UpdateClaim(c *gin.Context) {
	var (
		t            int64 = tool.MakeTimestamp()
		requestModel *request.PoolBrc20ClaimUpdateReq
		publicKey    string = ""
	)
	if c.ShouldBindJSON(&requestModel) == nil {
		publicKey = getAuthParams(c)
		responseModel, err := order_brc20_service.UpdateClaim(requestModel, publicKey, c.ClientIP())
		if err != nil {
			c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
			return
		}
		c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
		return
	}
	c.JSONP(http.StatusInternalServerError, respond.RespErr(errors.New("error parameter"), tool.MakeTimestamp()-t, respond.HttpsCodeError))
}

// @Summary Fetch pool reward info
// @Description Fetch pool reward info
// @Produce  json
// @Tags brc20
// @Param net query string true "net"
// @Param tick query string false "tick"
// @Param address query string true "address"
// @Param rewardType query int false "rewardType"
// @Success 200 {object} respond.PoolBrc20RewardResp ""
// @Router /brc20/pool/reward/info [get]
func FetchOwnerReward(c *gin.Context) {
	var (
		t             int64                       = tool.MakeTimestamp()
		rewardTypeStr                             = c.DefaultQuery("rewardType", "0")
		req           *request.PoolBrc20RewardReq = &request.PoolBrc20RewardReq{
			Net:     c.DefaultQuery("net", ""),
			Tick:    c.DefaultQuery("tick", ""),
			Address: c.DefaultQuery("address", ""),
		}
	)
	rewardType, _ := strconv.ParseInt(rewardTypeStr, 10, 64)
	req.RewardType = model.RewardType(rewardType)
	responseModel, err := order_brc20_service.FetchOwnerReward(req)
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
	return
}

// @Summary create one claim reward order
// @Description create one claim reward order
// @Produce  json
// @Param Request body request.PoolBrc20ClaimRewardReq true "Request"
// @Tags brc20
// @Success 200 {object} respond.Message ""
// @Router /brc20/pool/reward/claim [post]
func ClaimReward(c *gin.Context) {
	var (
		t            int64 = tool.MakeTimestamp()
		requestModel *request.PoolBrc20ClaimRewardReq
		publicKey    string = ""
	)
	if c.ShouldBindJSON(&requestModel) == nil {
		publicKey = getAuthParams(c)
		responseModel, err := order_brc20_service.ClaimReward(requestModel, publicKey, c.ClientIP())
		if err != nil {
			c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
			return
		}
		c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
		return
	}
	c.JSONP(http.StatusInternalServerError, respond.RespErr(errors.New("error parameter"), tool.MakeTimestamp()-t, respond.HttpsCodeError))
}

// @Summary Fetch pool reward orders
// @Description Fetch pool reward orders
// @Produce  json
// @Tags brc20
// @Param net query string false "net:mainnet/signet/testnet"
// @Param tick query string false "tick"
// @Param address query string false "address"
// @Param rewardState query int false "rewardState: 1-create,2-inscription,3-send,100-all"
// @Param limit query int false "limit: Max-50"
// @Param page query int false "page"
// @Param flag query int false "flag"
// @Param sortKey query string false "sortKey: timestamp, default:timestamp"
// @Param sortType query int false "sortType: 1/-1"
// @Success 200 {object} respond.PoolRewardOrderResponse ""
// @Router /brc20/pool/reward/orders [get]
func FetchPoolRewardOrders(c *gin.Context) {
	var (
		t              int64                            = tool.MakeTimestamp()
		rewardStateStr                                  = c.DefaultQuery("rewardState", "0")
		limitStr                                        = c.DefaultQuery("limit", "50")
		pageStr                                         = c.DefaultQuery("page", "0")
		flagStr                                         = c.DefaultQuery("flag", "0")
		sortTypeStr                                     = c.DefaultQuery("sortType", "0")
		rewardTypeStr                                   = c.DefaultQuery("rewardType", "0")
		req            *request.PoolRewardOrderFetchReq = &request.PoolRewardOrderFetchReq{
			Net:         c.DefaultQuery("net", ""),
			Tick:        c.DefaultQuery("tick", ""),
			RewardState: 0,
			Limit:       0,
			Flag:        0,
			Page:        0,
			Address:     c.DefaultQuery("address", ""),
			SortKey:     c.DefaultQuery("sortKey", ""),
			SortType:    0,
		}
	)
	rewardType, _ := strconv.ParseInt(rewardTypeStr, 10, 64)
	req.RewardType = model.RewardType(rewardType)
	rewardState, _ := strconv.ParseInt(rewardStateStr, 10, 64)
	req.RewardState = model.RewardState(rewardState)
	req.Limit, _ = strconv.ParseInt(limitStr, 10, 64)
	req.Flag, _ = strconv.ParseInt(flagStr, 10, 64)
	req.Page, _ = strconv.ParseInt(pageStr, 10, 64)
	req.SortType, _ = strconv.ParseInt(sortTypeStr, 10, 64)
	responseModel, err := order_brc20_service.FetchPoolRewardOrders(req)
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
	return
}

// @Summary Fetch event reward info
// @Description Fetch pool reward info
// @Produce  json
// @Tags brc20
// @Param net query string true "net"
// @Param tick query string false "tick"
// @Param address query string true "address"
// @Param rewardType query int false "rewardType"
// @Success 200 {object} respond.PoolBrc20RewardResp ""
// @Router /brc20/event/reward/info [get]
func FetchEventOwnerReward(c *gin.Context) {
	var (
		t             int64                       = tool.MakeTimestamp()
		rewardTypeStr                             = c.DefaultQuery("rewardType", "0")
		req           *request.PoolBrc20RewardReq = &request.PoolBrc20RewardReq{
			Net:     c.DefaultQuery("net", ""),
			Tick:    c.DefaultQuery("tick", ""),
			Address: c.DefaultQuery("address", ""),
		}
	)
	rewardType, _ := strconv.ParseInt(rewardTypeStr, 10, 64)
	req.RewardType = model.RewardType(rewardType)
	responseModel, err := order_brc20_service.FetchEventOwnerReward(req)
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
	return
}

// @Summary Cal event claim fee
// @Description Cal event claim fee
// @Produce  json
// @Tags brc20
// @Param version query int false "version"
// @Param networkFeeRate query int false "networkFeeRate"
// @Success 200 {object} respond.CalFeeResp ""
// @Router /brc20/event/reward/cal/fee [get]
func CalEventClaimFee(c *gin.Context) {
	var (
		t                 int64                        = tool.MakeTimestamp()
		versionStr        string                       = c.DefaultQuery("version", "0")
		networkFeeRateStr                              = c.DefaultQuery("networkFeeRate", "0")
		req               *request.OrderBrc20CalFeeReq = &request.OrderBrc20CalFeeReq{
			Version:        0,
			NetworkFeeRate: 0,
			Net:            c.DefaultQuery("net", "livenet"),
		}
	)
	req.Version, _ = strconv.ParseInt(versionStr, 10, 64)
	req.NetworkFeeRate, _ = strconv.ParseInt(networkFeeRateStr, 10, 64)
	responseModel, err := order_brc20_service.CalEventClaimFee(req)
	if err != nil {
		c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
		return
	}
	c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
	return
}

// @Summary create one event claim reward order
// @Description create one event claim reward order
// @Produce  json
// @Param Request body request.PoolBrc20ClaimRewardReq true "Request"
// @Tags brc20
// @Success 200 {object} respond.Message ""
// @Router /brc20/event/reward/claim [post]
func ClaimEventReward(c *gin.Context) {
	var (
		t            int64 = tool.MakeTimestamp()
		requestModel *request.PoolBrc20ClaimRewardReq
		publicKey    string = ""
	)
	if c.ShouldBindJSON(&requestModel) == nil {
		publicKey = getAuthParams(c)
		responseModel, err := order_brc20_service.ClaimEventReward(requestModel, publicKey, c.ClientIP())
		if err != nil {
			c.JSONP(http.StatusOK, respond.RespErr(err, tool.MakeTimestamp()-t, respond.HttpsCodeError))
			return
		}
		c.JSONP(http.StatusOK, respond.RespSuccess(responseModel, tool.MakeTimestamp()-t))
		return
	}
	c.JSONP(http.StatusInternalServerError, respond.RespErr(errors.New("error parameter"), tool.MakeTimestamp()-t, respond.HttpsCodeError))
}
