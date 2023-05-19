package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/gin-gonic/gin"
)

/**
時間中間件，記錄處理每個請求所需要的時間
*/
func ResponseTime() gin.HandlerFunc {
	return func(c *gin.Context) {
		//t := time.Now()
		c.Next()
		//latecny := time.Since(t)
		//status := c.Writer.Status()
		//logger.InfoS(fmt.Sprintf("%s\t[%v]\t%v", c.FullPath(), status, latecny))
	}
}



func generateAPIKey() string {
	// 生成32字节的随机字节序列
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return ""
	}
	// 将字节序列转换为Base64字符串
	return base64.StdEncoding.EncodeToString(key)
}
