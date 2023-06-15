package middleware

import (
	"github.com/gin-gonic/gin"
)

func ResponseTime() gin.HandlerFunc {
	return func(c *gin.Context) {
		//t := time.Now()
		c.Next()
		//latecny := time.Since(t)
		//status := c.Writer.Status()
		//logger.InfoS(fmt.Sprintf("%s\t[%v]\t%v", c.FullPath(), status, latecny))
	}
}

