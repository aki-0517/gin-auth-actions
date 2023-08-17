package main

import (
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/aki-0517/go-user-management/handlers"
	"github.com/aki-0517/go-user-management/middleware"
	"github.com/aki-0517/go-user-management/util"
)

type App struct {
	DB     *gorm.DB
	JWTKey []byte
	RDB    *redis.Client
}

func main() {
	var app App

	app.JWTKey = []byte(os.Getenv("JWT_KEY"))

	app.DB = util.DBConnect()
	sqlDB, err := app.DB.DB()
	if err != nil {
		panic("Failed to retrieve the database connection")
	}
	defer sqlDB.Close()

	app.RDB = util.RedisClient()

	uh := handlers.UserHandler(app.DB, app.JWTKey)
	ah := handlers.AuthHandlerInit(app.DB, app.JWTKey, app.RDB)
	m := middleware.NewMiddleware(app.JWTKey)

	r := gin.Default()

	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	rl := util.NewRateLimiter(5)
	r.Use(func(c *gin.Context) {
		rl.MiddleWare()
		c.Next()
	})

	authorized := r.Group("/me")
	authorized.Use(m.AuthenticateMiddleware())
	{
		authorized.PUT("/:id", uh.UpdateUserHandler())
		authorized.PUT("/:id/password", ah.ChangePasswordHandler())
		authorized.DELETE("/:id", uh.DeleteUserHandler())
		authorized.POST("/refresh-token", ah.RefreshTokenHandler())
		authorized.GET("/:id", uh.GetUserHandler())
		authorized.POST("/logout", ah.LogOutHandler())
	}

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World!")
	})
	r.GET("/users", uh.ListUsersHandler())
	r.GET("/user/:id", uh.GetUserHandler())
	r.POST("/user", uh.CreateUserHandler())
	r.POST("/login", ah.LoginHandler())

	r.Run(":8080")
}
