package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)


func RateLimiter(rdb *redis.Client) gin.HandlerFunc{
	return func(c *gin.Context){
		//use ip address to identifying the clinet 
		ip := c.ClientIP()
		key := "rate:" + ip 

		limit := 5 
		window := 1 * time.Minute

		ctx := context.Background()

		//increment and expiry in same network trip 
		pipe := rdb.Pipeline()
		incr := pipe.Incr(ctx,key)

		pipe.Expire(ctx, key, window)

		_, err := pipe.Exec(ctx)
        
		if err != nil {
			c.Next()//log the error but move and don't stop the business logic
			return
		}
		count := incr.Val()
        if count > int64(limit){
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":"Too many requests, slow down",
				"retry after":window.String(),
			})
			return
		}
		c.Next()
	}
	
}