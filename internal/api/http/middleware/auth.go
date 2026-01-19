// Package middleware provides HTTP middleware for the AI Gateway.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"ai-gateway/config"
)

// Auth creates an authentication middleware with the given configuration.
// If auth is disabled, it only extracts the API key without validation.
// If auth is enabled, it validates the API key against the configured list.
func Auth(authCfg config.AuthConfig) gin.HandlerFunc {
	// Build a set of valid keys for O(1) lookup
	validKeys := make(map[string]bool)
	for _, key := range authCfg.APIKeys {
		if key != "" {
			validKeys[key] = true
		}
	}

	return func(c *gin.Context) {
		var apiKey string

		// Try OpenAI format first (Authorization: Bearer xxx)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				apiKey = parts[1]
			}
		}

		// Try Anthropic format (x-api-key: xxx)
		if apiKey == "" {
			apiKey = c.GetHeader("x-api-key")
		}

		// Check if API key is provided
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": "Missing API key. Please provide via Authorization header (Bearer) or x-api-key header.",
					"type":    "authentication_error",
				},
			})
			return
		}

		// Validate API key if auth is enabled
		if authCfg.Enabled {
			if !validKeys[apiKey] {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": gin.H{
						"message": "Invalid API key.",
						"type":    "authentication_error",
					},
				})
				return
			}
		}

		// Store API key in context for potential usage (logging, rate limiting, etc.)
		c.Set("api_key", apiKey)
		c.Next()
	}
}
