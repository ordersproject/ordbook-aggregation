package auth

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"ordbook-aggregation/controller/respond"
	"ordbook-aggregation/tool"
)

const (
	verifyMessage string = "orders.exchange"
)

var (
	AuthErrParams1 error = errors.New("Auth params is empty(signature)")
	AuthErrParams2 error = errors.New("Auth params is empty(public-key)")
	AuthErrParamsVerifiedSignErr error = errors.New("Auth verified signature err")
	AuthErrParamsVerifiedSignWrong error = errors.New("Auth verified signature wrong")
)

func AuthSignMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := tool.MakeTimestamp()
		signatureStr := c.Request.Header.Get("X-signature")
		if signatureStr == "" {
			c.JSON(http.StatusUnauthorized, respond.RespErr(AuthErrParams1, tool.MakeTimestamp()-t, respond.HttpsCodeErrorAuth))
			c.Abort()
			return
		}
		publicKeyStr := c.Request.Header.Get("X-public-key")
		if publicKeyStr == "" {
			c.JSON(http.StatusUnauthorized, respond.RespErr(AuthErrParams2, tool.MakeTimestamp()-t, respond.HttpsCodeErrorAuth))
			c.Abort()
			return
		}

		verified, err := VerifySign(verifyMessage, signatureStr, publicKeyStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, respond.RespErr(AuthErrParamsVerifiedSignErr, tool.MakeTimestamp()-t, respond.HttpsCodeErrorAuth))
			c.Abort()
			return
		}

		if !verified {
			c.JSON(http.StatusUnauthorized, respond.RespErr(AuthErrParamsVerifiedSignWrong, tool.MakeTimestamp()-t, respond.HttpsCodeErrorAuth))
			c.Abort()
			return
		}

		c.Set("publicKey", publicKeyStr)
		c.Next()
	}
}