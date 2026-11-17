package utils

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(public []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		for _, p := range public {
			if path == p || strings.HasPrefix(path, p+"/") || matchRoutePattern(path, p) {
				c.Next()
				return
			}
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			authHeader = c.Query("token")
		}

		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "กรุณาเข้าสู่ระบบก่อนใช้งาน",
			})
			return
		}

		tokenString := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else if parts := strings.Split(authHeader, " "); len(parts) == 2 {
			tokenString = parts[1]
		}

		jwtSecret := []byte(os.Getenv("JWT_SECRET"))
		if len(jwtSecret) == 0 {
			jwtSecret = []byte("binance_journal_super_secret_access_key_2026")
		}

		token, err := ValidateToken(tokenString, jwtSecret)
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "โทเคนไม่ถูกต้องหรือหมดอายุแล้ว",
			})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_id", claims["user_id"])
		}

		c.Next()
	}
}

func GenerateTokenPair(userID string) (string, string, error) {
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("binance_journal_super_secret_access_key_2026")
	}

	jwtRefreshSecret := []byte(os.Getenv("JWT_REFRESH_SECRET"))
	if len(jwtRefreshSecret) == 0 {
		jwtRefreshSecret = []byte("binance_journal_super_secret_refresh_key_2026")
	}

	jwtExpHrs, err := strconv.Atoi(os.Getenv("JWT_EXPIRATION_HOURS"))
	if err != nil || jwtExpHrs <= 0 {
		jwtExpHrs = 24
	}

	atClaims := jwt.MapClaims{
		"user_id": userID,
		"type":    "access",
		"exp":     time.Now().Add(time.Duration(jwtExpHrs) * time.Hour).Unix(),
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	accessToken, err := at.SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}

	rtClaims := jwt.MapClaims{
		"user_id": userID,
		"type":    "refresh",
		"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(),
	}
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	refreshToken, err := rt.SignedString(jwtRefreshSecret)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func ValidateToken(tokenString string, secret []byte) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
}

func matchRoutePattern(reqPath, pattern string) bool {
	reqParts := strings.Split(strings.Trim(reqPath, "/"), "/")
	patParts := strings.Split(strings.Trim(pattern, "/"), "/")

	if len(reqParts) != len(patParts) {
		return false
	}

	for i := 0; i < len(patParts); i++ {
		if strings.HasPrefix(patParts[i], ":") {
			continue
		}
		if reqParts[i] != patParts[i] {
			return false
		}
	}
	return true
}
