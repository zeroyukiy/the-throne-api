package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	echojwt "github.com/labstack/echo-jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/zeroyukiy/the-throne-api/handler"
	"github.com/zeroyukiy/the-throne-api/internal"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
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

func cors(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		c.Response().Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Response().Header().Set("Access-Control-Allow-Headers", "origin, content-type, accept, authorization")
		c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Response().Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request().Method == "OPTIONS" {
			return c.JSON(http.StatusOK, nil)
		}

		return next(c)
	}
}

// func spa(next echo.HandlerFunc) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		if c.Response().Status == http.StatusNotFound {
// 			fmt.Println("status: ", 404)
// 			t, err := template.ParseFiles("public/index.html")
// 			if err != nil {
// 				fmt.Println(err)
// 			}
// 			return t.ExecuteTemplate(c.Response(), "index.html", nil)
// 		}
// 		return next(c)
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

	// chatHandler := handler.NewChatHandler(hub)

	e := echo.New()
	e.Use(cors)

	// e.Use(middleware.Logger())
	// e.Use(middleware.Gzip())

	// e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
	// 	Root:   "public",     // This is the path to your SPA build folder, the folder that is created from running "npm build"
	// 	Index:  "index.html", // This is the default html page for your SPA
	// 	Browse: false,
	// 	HTML5:  true,
	// }))

	e.Static("/assets", "./public")

	// DATABASE
	// conn := database.Init()
	// if err := conn.Ping(); err != nil {
	// 	log.Fatal("database error connection: ", err)
	// }
	var conn *sqlx.DB = &sqlx.DB{}

	handler := handler.NewAuthHandler(conn)

	e.GET("/", func(c echo.Context) error {
		t, err := template.ParseFiles("index.html")
		if err != nil {
			fmt.Println(err)
		}
		return t.ExecuteTemplate(c.Response(), "index.html", nil)
	})

	// these endpoints need limit middleware
	e.POST("/api/auth/login", handler.Login)
	e.POST("/api/auth/logout", handler.Logout)

	csrf_config := middleware.CSRFConfig{
		TokenLookup:    "cookie:_csrf",
		CookiePath:     "/",
		CookieDomain:   "localhost",
		CookieSecure:   true,
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteStrictMode,
	}
	refresh := e.Group("/api/auth", middleware.CSRFWithConfig(csrf_config))
	{
		refresh.GET("/refresh", handler.Refresh)
	}

	config := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(internal.JwtCustomClaims)
		},
		SigningKey: []byte("my_secret_string"),
	}

	r := e.Group("/api")
	{
		// r.GET("/csrf-token", func(c echo.Context) error {
		// 	// fmt.Println(c.Get("csrf").(string))
		// 	csrf := c.Get("csrf").(string)

		// 	return c.JSON(http.StatusOK, echo.Map{
		// 		"_csrf": csrf,
		// 	})
		// })

		r.Use(echojwt.WithConfig(config))

		r.GET("/me", func(c echo.Context) error {
			claims := c.Get("user").(*jwt.Token).Claims.(*internal.JwtCustomClaims)

			return c.JSON(http.StatusOK, echo.Map{
				"user": map[string]interface{}{
					"name":   claims.Name,
					"avatar": claims.Avatar,
				},
			})
		})

		type MessageBinding struct {
			Message string `json:"message"`
		}

		r.POST("/message", func(c echo.Context) error {
			claims := c.Get("user").(*jwt.Token).Claims.(*internal.JwtCustomClaims)

			message := &MessageBinding{}
			err := c.Bind(message)
			if err != nil {
				fmt.Println(err)
				return c.JSON(http.StatusInternalServerError, err)
			}

			msg := &Message{
				Username:  claims.Name,
				Avatar:    claims.Avatar,
				Message:   message.Message,
				CreatedAt: time.Now().Format("15:04"),
			}

			messages = append(messages, msg)
			return c.JSON(http.StatusOK, nil)
		})

		type ChatRoomBinding struct {
			Id string `param:"chat_id"`
		}

		r.GET("/messages/:chat_id", func(c echo.Context) error {
			chat_room := &ChatRoomBinding{}
			err := c.Bind(chat_room)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, err)
			}
			return c.JSON(http.StatusOK, messages)
		})
	}

	// e.GET("/avatars", cors(func(c echo.Context) error {
	// 	avatars := make(map[string][]string)
	// 	// entries, err := os.ReadDir("./sticker")
	// 	entries, err := os.ReadDir("./faceset")
	// 	if err != nil {
	// 		log.Println(err)
	// 		return c.JSON(http.StatusInternalServerError, err)
	// 	}
	// 	for _, e := range entries {
	// 		avatars["avatars"] = append(avatars["avatars"], e.Name())
	// 	}
	// 	return c.JSON(http.StatusOK, avatars)
	// }))

	// e.GET("/api/chat/:id", cors(chatHandler.GetRoom))

	e.GET("/ws", func(c echo.Context) error {
		authBinding := &struct {
			Authorization string `query:"auth"`
		}{}

		err := c.Bind(authBinding)
		if err != nil {
			fmt.Println(err)
			return c.JSON(http.StatusBadRequest, "no auth")
		}

		claims := &internal.JwtCustomClaims{}
		_, err = jwt.ParseWithClaims(authBinding.Authorization, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte("my_secret_string"), nil
		})
		if err != nil {
			fmt.Println(err)
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": "error with the token",
			})
		}

		fmt.Println(claims)

		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			log.Fatal(err)
		}

		// create a client
		client := internal.NewClient(conn, hub, claims.Name, claims.Avatar)
		hub.RegisterClient(client)
		client.Run()

		return nil
	})

	type AccessTokenResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		Expires      int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}

	type UserDiscordResponse struct {
		Id       string `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		//   "username": "Nelly",
		//   "global_name": null,
		//   "discriminator": "1337",
		//   "avatar": "8342729096ea3675442027381ff50dfe",
		//   "verified": true,
		//   "email": "nelly@discord.com",
		//   "flags": 64,
		//   "banner": "06c16474723fe537c283b8efa61a30c8",
		//   "accent_color": 16711680,
		//   "premium_type": 1,
		//   "public_flags": 64,
		//   "avatar_decoration_data": {
		//     "sku_id": "1144058844004233369",
		//     "asset": "a_fed43ab12698df65902ba06727e20c0e"
		//   },
		//   "collectibles": {
		//     "nameplate": {
		//       "sku_id": "2247558840304243311",
		//       "asset": "nameplates/nameplates/twilight/",
		//       "label": "",
		//       "palette": "cobalt"
		//     }
		//   },
		//   "primary_guild": {
		//     "identity_guild_id": "1234647491267808778",
		//     "identity_enabled": true,
		//     "tag": "DISC",
		//     "badge": "7d1734ae5a615e82bc7a4033b98fade8"
		//   }
	}

	type CodeBinding struct {
		Code string `query:"code"`
	}
	e.GET("/prova", func(c echo.Context) error {
		endpoint := "https://discord.com/api/v10"
		client_id := os.Getenv("DISCORD_CLIENT_ID")
		client_secret := os.Getenv("DISCORD_CLIENT_SECRET")
		redirect_uri := "http://localhost:8000/prova"

		code := &CodeBinding{}
		err := c.Bind(code)
		if err != nil {
			fmt.Println(err)
			return c.JSON(http.StatusInternalServerError, nil)
		}

		data := url.Values{}
		data.Set("grant_type", "authorization_code")
		data.Set("code", code.Code)
		data.Set("redirect_uri", redirect_uri)

		client := &http.Client{}
		req, err := http.NewRequest("POST", fmt.Sprintf("%s/oauth2/token", endpoint), strings.NewReader(data.Encode()))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(client_id, client_secret)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()

		// client.Post(fmt.Sprintf("%s/oauth2/token", endpoint), "", nil)

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}

		accessTokenResponse := &AccessTokenResponse{}
		err = json.Unmarshal(b, accessTokenResponse)
		if err != nil {
			fmt.Println(err)
		}

		user := &UserDiscordResponse{}

		req, err = http.NewRequest("GET", "https://discord.com/api/users/@me", nil)
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Add("Authorization", "Bearer "+accessTokenResponse.AccessToken)
		resp, err = client.Do(req)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()

		// fmt.Println(resp)

		b, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}

		err = json.Unmarshal(b, user)
		if err != nil {
			fmt.Println(err)
		}

		return c.JSON(http.StatusOK, echo.Map{
			"user": user,
		})
	})

	e.Logger.Fatal(e.Start("localhost:8000"))
}
