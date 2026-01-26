package handler

import "github.com/gin-gonic/gin"

func ctxGetInt64(c *gin.Context, key string) int64 {
	val, exists := c.Get(key)
	if !exists {
		return 0
	}
	if id, ok := val.(int64); ok {
		return id
	}
	return 0
}

func ctxGetInt64Ptr(c *gin.Context, key string) *int64 {
	val, exists := c.Get(key)
	if !exists {
		return nil
	}
	if id, ok := val.(int64); ok {
		return &id
	}
	return nil
}
