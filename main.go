package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/zeroyukiy/the-throne-api/database"
	"github.com/zeroyukiy/the-throne-api/database/model"
	"github.com/zeroyukiy/the-throne-api/database/repository"
	"github.com/zeroyukiy/the-throne-api/internal"
	"github.com/zeroyukiy/the-throne-api/internal/handler"
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

	// DATABASE
	conn := database.Init()
	if err := conn.Ping(); err != nil {
		log.Fatal("database error connection: ", err)
	}
	defer conn.Close()

	router := gin.Default()
	// loggerConfig := gin.LoggerConfig{SkipPaths: []string{"/chat/example"}}
	// router := gin.New()
	// router.Use(gin.LoggerWithConfig(loggerConfig))

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

	// store, err := postgres.NewStore(conn.DB, []byte("secret"))
	// if err != nil {
	// 	// handle err
	// }

	store.Options(sessions.Options{
		HttpOnly: true,
		// Secure:   true,
		MaxAge: 3600 * 2,
	})

	router.Use(sessions.Sessions("mysession", store))

	// e.Use(middleware.Logger())
	// e.Use(middleware.Gzip())

	router.Static("/assets", "./public")

	{
		chatRouter := router.Group("/chat")
		chathandler := handler.NewChatHandler(conn)
		chatRouter.GET("/all", chathandler.Index)
		chatRouter.GET("/:slug", chathandler.Show)
	}

	{
		userRouter := router.Group("/users")
		userRouter.GET("/:id", func(c *gin.Context) {
			id := c.Param("id")
			userRepo := repository.NewUserRepository(conn)
			user, err := userRepo.FindOne(id)
			if err != nil {
				fmt.Println(err)
			}
			c.JSON(http.StatusOK, gin.H{
				"user_id":  user.Id,
				"username": user.Username,
			})
		})
	}

	router.GET("/ws/chat/list", func(c *gin.Context) {
		type Room struct {
			Id         string   `json:"room_id"`
			Status     string   `json:"status"`
			Users      []string `json:"clients"`
			TotalUsers int      `json:"total_clients"`
			Distinct   int      `json:"distinct"`
		}
		type Result struct {
			Rooms []Room `json:"rooms"`
		}
		res := Result{
			Rooms: []Room{},
		}
		list := hub.GetRoomRepository().ListRooms()
		for _, r := range list {
			room := Room{
				Id:     r.GetRoomId(),
				Status: r.GetStatus(),
			}
			for _, client := range r.GetClients() {
				room.Users = append(room.Users, client.GetUserId())
			}
			room.TotalUsers = len(room.Users)
			room.Distinct = r.GetUniqueUsers()
			res.Rooms = append(res.Rooms, room)
		}
		sort.Slice(res.Rooms, func(i, j int) bool {
			if res.Rooms[i].Id < res.Rooms[j].Id {
				return true
			} else {
				return false
			}
		})
		c.JSON(http.StatusOK, res)
	})

	// router.GET("/chat/example", func(c *gin.Context) {
	// 	type Result struct {
	// 		RoomId  string   `json:"room_id"`
	// 		Clients []string `json:"clients"`
	// 	}
	// 	res := Result{
	// 		RoomId:  "example",
	// 		Clients: nil,
	// 	}
	// 	c.JSON(http.StatusOK, res)
	// })

	router.POST("/login", func(c *gin.Context) {
		var login LoginForm
		if err := c.ShouldBindJSON(&login); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userRepo := repository.NewUserRepository(conn)
		user, err := userRepo.Get(login.Username, login.Password)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		// uid := uuid.New()
		// user.Id = uid.String()
		// user.Avatar = fmt.Sprintf("http://localhost:8000/assets/avatars/%s", user.Avatar)

		// user := model.User{
		// 	Id:       uid.String(),
		// 	Username: login.Username,
		// 	Avatar:   "http://localhost:8000/assets/avatars/pippo.jpg",
		// }

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
		var user *model.User
		session := sessions.Default(c)
		//

		//
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
				// Secure:   true,
				MaxAge: 3600 * 2,
			})
			session.Save()
			c.JSON(http.StatusOK, gin.H{"user": user})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
	})

	router.GET("/users/:id/cards", func(c *gin.Context) {
		user_id := c.Param("id")
		id, err := strconv.ParseInt(user_id, 10, 64)
		if err != nil {
			fmt.Println(err)
		}
		userRepo := repository.NewUserRepository(conn)
		cards, err := userRepo.GetCards(int(id))
		if err != nil {
			fmt.Println(err)
			return
		}
		c.JSON(http.StatusOK, cards)
	})

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "welcome"})
	})

	router.GET("/ws", func(c *gin.Context) {
		var user *model.User

		// user = &model.User{
		// 	Username: "pippo",
		// 	Id:       "123",
		// 	Avatar:   "http://localhost:8000/assets/avatars/pippo.jpg",
		// }

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

	router.Run("127.0.0.1:8000")
}
