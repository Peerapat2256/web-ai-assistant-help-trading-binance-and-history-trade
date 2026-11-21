package system

import (
	"backend/controller"
	"backend/utils"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func StartServer() {
	if err := godotenv.Load(); err != nil {
		_ = godotenv.Load("../.env")
	}

	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "debug"
	}
	gin.SetMode(ginMode)

	router := gin.Default()
	_ = router.SetTrustedProxies(nil)

	// ตั้งค่า CORS รองรับทั้ง Localhost และโดเมน Vercel
	router.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return strings.HasPrefix(origin, "http://localhost") ||
				strings.HasPrefix(origin, "http://127.0.0.1") ||
				strings.HasSuffix(origin, ".vercel.app") ||
				strings.Contains(origin, "onrender.com")
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.Static("/uploads", "./uploads")
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	publicPaths := []string{
		"/ping",
		"/uploads",
		"/auth/login",
		"/auth/google-login",
		"/auth/register",
		"/auth/refresh-token",
		"/auth/forgot-password",
		"/auth/verify-reset-token",
		"/auth/reset-password",
	}
	router.Use(utils.AuthMiddleware(publicPaths))

	controller.AuthenController(router)
	controller.PortfolioController(router)
	controller.TradeController(router)
	controller.DashboardController(router)
	controller.AiController(router)
	controller.NewsController(router)
	controller.SheetController(router)
	controller.WithdrawalController(router)

	// รองรับทั้ง Localhost (SERVER_PORT) และ Cloud / Render (PORT)
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}

	log.Printf("เซิร์ฟเวอร์ Backend พร้อมทำงานที่พอร์ต %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("ไม่สามารถเปิดเซิร์ฟเวอร์ได้: %v", err)
	}
}
