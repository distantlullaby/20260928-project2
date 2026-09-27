package middleware

import (
	"net/http"
	"strings"

	"taste-server/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// CurrentUserID 注入到 gin.Context 中的当前登录用户 ID
const CurrentUserID = "currentUserID"

// Auth 解析 Bearer Token 并写入当前用户 ID
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		if tokenStr == "" || tokenStr == header {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(config.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登录已失效，请重新登录"})
			return
		}
		uidFloat, ok := claims["uid"].(float64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "无效的凭证"})
			return
		}
		c.Set(CurrentUserID, uint(uidFloat))
		c.Next()
	}
}

// OptionalAuth 可选登录：带了有效 Token 就写入用户 ID，方便 Feed 流标记 mine
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		if tokenStr != "" && tokenStr != header {
			claims := jwt.MapClaims{}
			if token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
				return []byte(config.JWTSecret), nil
			}); err == nil && token.Valid {
				if uidFloat, ok := claims["uid"].(float64); ok {
					c.Set(CurrentUserID, uint(uidFloat))
				}
			}
		}
		c.Next()
	}
}

// GetUserID 从上下文取出当前用户 ID
func GetUserID(c *gin.Context) uint {
	if v, ok := c.Get(CurrentUserID); ok {
		return v.(uint)
	}
	return 0
}
