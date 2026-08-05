package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/zeroyukiy/the-throne-api/internal"
)

type LoginForm struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Message struct {
	Username  string `json:"username"`
	Avatar    string `json:"avatar"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

var messages []*Message = []*Message{
	{
		Username:  "John Snow Prime",
		Avatar:    "",
		Message:   "Sed ligula tellus, dignissim non urna sed, commodo ullamcorper lorem. Cras ac scelerisque mauris. Aliquam metus neque, fringilla a ligula id, auctor tempus nulla. Donec euismod libero quis felis eleifend blandit pharetra in libero. Aenean interdum ultrices lorem, eu interdum velit convallis eu.",
		CreatedAt: time.Now().Format("15:04"),
	},
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// func cors() echo.MiddlewareFunc {
// 	return func(next echo.HandlerFunc) echo.HandlerFunc {
// 		return func(c *echo.Context) error {
// 			c.Response().Header().Set("Access-Control-Allow-Origin", "*")
// 			c.Response().Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
// 			c.Response().Header().Set("Access-Control-Allow-Headers", "origin, content-type, accept, authorization")
// 			c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
// 			c.Response().Header().Set("Access-Control-Allow-Credentials", "true")

// 			if c.Request().Method == "OPTIONS" {
// 				return c.JSON(http.StatusOK, nil)
// 			}

// 			return next(c)
// 		}
// 	}
// }

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	hub := internal.NewHub()
	go hub.Run()

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Create cookie-based session store with a secret key
	store := cookie.NewStore([]byte("your-secret-key"))
	store.Options(sessions.Options{
		HttpOnly: true,
		Secure:   true,
		MaxAge:   3600 * 2,
	})
	router.Use(sessions.Sessions("mysession", store))

	// e.Use(middleware.Logger())
	// e.Use(middleware.Gzip())

	router.Static("/assets", "./public")

	// DATABASE
	// conn := database.Init()
	// if err := conn.Ping(); err != nil {
	// 	log.Fatal("database error connection: ", err)
	// }
	// var conn *sqlx.DB = &sqlx.DB{}

	// handler := handler.NewAuthHandler(conn)

	router.POST("/login", func(c *gin.Context) {
		var login LoginForm
		if err := c.ShouldBindJSON(&login); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user := internal.User{
			Username:  login.Username,
			Age:       99,
			CreatedAt: time.Now().UnixMilli(),
		}
		b, err := json.Marshal(user)
		if err != nil {
			fmt.Println(err)
		}
		session := sessions.Default(c)
		session.Set("user", b)
		session.Save()
		c.JSON(http.StatusOK, gin.H{"user": user})
	})

	router.POST("/logout", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		session.Save()
		c.JSON(http.StatusOK, gin.H{"message": "logged out"})
	})

	router.GET("/@me", func(c *gin.Context) {
		var user *internal.User
		session := sessions.Default(c)
		u := session.Get("user")
		if u != nil {
			err := json.Unmarshal(u.([]byte), &user)
			if err != nil {
				fmt.Println(err)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "not logged in"})
				return
			}
			session.Options(sessions.Options{
				HttpOnly: true,
				Secure:   true,
				MaxAge:   3600 * 2,
			})
			session.Save()
			c.JSON(http.StatusOK, gin.H{"user": user})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
	})

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "welcome"})
	})

	router.GET("/ws", func(c *gin.Context) {
		var user *internal.User
		session := sessions.Default(c)
		u := session.Get("user")
		if u != nil {
			err := json.Unmarshal(u.([]byte), &user)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "not logged in"})
				return
			}
			internal.ServeWs(hub, c.Writer, c.Request, user)
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
	})

	router.Run("localhost:8000")
}
