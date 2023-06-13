package controller

import "github.com/gin-gonic/gin"

func getAuthParams(c *gin.Context) string {
	publicKey := ""
	if publicKeyInterface, exists := c.Get("publicKey"); exists {
		publicKey = publicKeyInterface.(string)
	}
	return publicKey
}