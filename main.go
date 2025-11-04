package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/zeroyukiy/the-throne-api/handler"
	"github.com/zeroyukiy/the-throne-api/internal"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func cors(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return next(c)
	}
}

func auth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("user", c)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, nil)
		}
		if sess.Values["username"] != "pippo" {
			return c.JSON(http.StatusUnauthorized, nil)
		}
		sess.Save(c.Request(), c.Response())
		return next(c)
	}
}

// type jwtCustomClaims struct {
// 	Name  string `json:"name"`
// 	Admin bool   `json:"admin"`
// 	jwt.RegisteredClaims
// }

type LoginBinding struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func login(c echo.Context) error {
	login := &LoginBinding{}
	err := c.Bind(login)
	if err != nil {
		return echo.ErrBadRequest
	}

	if login.Username != "pippo" || login.Password != "abcd1234" {
		return echo.ErrNotFound
	}

	sess, err := session.Get("user", c)
	if err != nil {
		fmt.Println(err)
	}

	sess.Options.HttpOnly = true
	sess.Options.SameSite = http.SameSiteLaxMode

	sess.Values["user_id"] = 1
	sess.Values["is_auth"] = true
	sess.Values["username"] = login.Username

	sess.Save(c.Request(), c.Response())

	// c.SetCookie(&http.Cookie{
	// 	Name:     "user",
	// 	Value:    "ok",
	// 	Path:     "/",
	// 	Domain:   "localhost",
	// 	Secure:   false,
	// 	Expires:  time.Now().Add(time.Hour * 24 * 30),
	// 	HttpOnly: true,
	// 	SameSite: http.SameSiteLaxMode,
	// })

	t := "ok"

	// claims := &jwtCustomClaims{
	// 	Name:  "Pippo",
	// 	Admin: true,
	// 	RegisteredClaims: jwt.RegisteredClaims{
	// 		Issuer:    "localhost",
	// 		Subject:   "user",
	// 		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 18)),
	// 	},
	// }

	// token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// t, err := token.SignedString([]byte("my_secret_string"))
	// if err != nil {
	// 	return err
	// }

	// claims_2 := &jwtCustomClaims{
	// 	Name:  "Pippo",
	// 	Admin: true,
	// 	RegisteredClaims: jwt.RegisteredClaims{
	// 		Issuer:    "localhost",
	// 		Subject:   "user",
	// 		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)),
	// 		NotBefore: jwt.NewNumericDate(time.Now()),
	// 	},
	// }

	// refresh_token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims_2)

	// r, err := refresh_token.SignedString([]byte("my_secret_string"))
	// if err != nil {
	// 	return err
	// }

	// // put the refresh_token in the cookie
	// c.SetCookie(&http.Cookie{
	// 	Name:     "refresh_token",
	// 	Value:    r,
	// 	Path:     "/",
	// 	Domain:   "localhost",
	// 	Secure:   false,
	// 	Expires:  time.Now().Add(time.Hour * 24 * 30),
	// 	HttpOnly: true,
	// 	SameSite: http.SameSiteLaxMode,
	// })

	return c.JSON(http.StatusOK, echo.Map{
		"token": t,
	})
}

// func refresh(c echo.Context) error {
// 	cookie, err := c.Cookie("refresh_token")
// 	if err != nil {
// 		fmt.Println(err)
// 		return c.JSON(http.StatusUnauthorized, echo.Map{
// 			"error": "no refresh_token",
// 		})
// 	}

// 	claims := &jwtCustomClaims{}
// 	token, err := jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (interface{}, error) {
// 		return []byte("my_secret_string"), nil
// 	})
// 	if err != nil {
// 		fmt.Println(err)
// 		return c.JSON(http.StatusInternalServerError, echo.Map{
// 			"error": "error",
// 		})
// 	}

// 	fmt.Println("claims token from cookie: ", claims)

// 	if claims.ExpiresAt.Unix() < time.Now().Unix() {
// 		return c.JSON(http.StatusUnauthorized, echo.Map{
// 			"error": "refresh_token has expired, you need to login.",
// 		})
// 	}

// 	// TODO clean this shit
// 	claims = &jwtCustomClaims{
// 		Name:  claims.Name,
// 		Admin: claims.Admin,
// 		RegisteredClaims: jwt.RegisteredClaims{
// 			Issuer:    "localhost",
// 			Subject:   "user",
// 			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 18)),
// 		},
// 	}

// 	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

// 	t, err := token.SignedString([]byte("my_secret_string"))
// 	if err != nil {
// 		return err
// 	}

// 	claims_2 := &jwtCustomClaims{
// 		Name:  claims.Name,
// 		Admin: claims.Admin,
// 		RegisteredClaims: jwt.RegisteredClaims{
// 			Issuer:    "localhost",
// 			Subject:   "user",
// 			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)),
// 			NotBefore: jwt.NewNumericDate(time.Now()),
// 		},
// 	}

// 	refresh_token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims_2)

// 	r, err := refresh_token.SignedString([]byte("my_secret_string"))
// 	if err != nil {
// 		return err
// 	}

// 	// put the refresh_token in the cookie
// 	c.SetCookie(&http.Cookie{
// 		Name:     "refresh_token",
// 		Value:    r,
// 		Path:     "/",
// 		Domain:   "localhost",
// 		Secure:   false,
// 		Expires:  time.Now().Add(time.Hour * 24 * 30),
// 		HttpOnly: true,
// 		SameSite: http.SameSiteLaxMode,
// 	})

// 	// the garbage end here

// 	return c.JSON(http.StatusOK, echo.Map{
// 		"message": "new access token and refresh token",
// 		"token":   t,
// 	})
// }

func main() {
	hub := internal.NewHub()
	go hub.Run()

	chatHandler := handler.NewChatHandler(hub)

	e := echo.New()
	e.Use(cors)
	sessionStore := sessions.NewFilesystemStore("./sessions", []byte("hello"))
	sessionStore.MaxAge(3600) // expires after 1 hr
	e.Use(session.Middleware(sessionStore))

	e.Static("/assets/avatars", "./faceset")
	e.GET("/", func(c echo.Context) error {
		sess, err := session.Get("user", c)
		if err != nil {
			fmt.Println(err)
		}
		sess.Save(c.Request(), c.Response())

		t, err := template.ParseFiles("index.html")
		if err != nil {
			fmt.Println(err)
		}
		return t.ExecuteTemplate(c.Response(), "index.html", sess.Values["username"])
	})

	e.POST("/login", login)

	r := e.Group("/api")
	{
		r.Use(auth)
		// r.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))
		// r.Use(session.Middleware(sessions.NewFilesystemStore("./sessions", []byte("hello"))))

		// config := echojwt.Config{
		// 	NewClaimsFunc: func(c echo.Context) jwt.Claims {
		// 		return new(jwtCustomClaims)
		// 	},
		// 	SigningKey: []byte("my_secret_string"),
		// }

		// r.Use(echojwt.WithConfig(config))

		r.GET("/restricted_page", func(c echo.Context) error {
			session, err := session.Get("user", c)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(session)
			name := session.Values["username"].(string)
			return c.JSON(http.StatusOK, echo.Map{
				"message": "Welcome " + name + "!",
			})
		})
	}

	e.GET("/prova", func(c echo.Context) error {
		u := struct {
			Name string `json:"name"`
		}{}
		err := c.Bind(&u)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
		user := &User{
			Name: u.Name,
			Age:  20,
		}
		return c.JSON(http.StatusOK, user)
	})

	e.GET("/avatars", cors(func(c echo.Context) error {
		avatars := make(map[string][]string)
		// entries, err := os.ReadDir("./sticker")
		entries, err := os.ReadDir("./faceset")
		if err != nil {
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, err)
		}
		for _, e := range entries {
			avatars["avatars"] = append(avatars["avatars"], e.Name())
		}
		return c.JSON(http.StatusOK, avatars)
	}))

	e.GET("/chat/:id", cors(chatHandler.GetRoom))

	e.GET("/ws", func(c echo.Context) error {
		sess, err := session.Get("user", c)
		if err != nil {
			fmt.Println(err)
		}

		if sess.Values["username"] == nil {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"error": "you're not logged",
			})
		}

		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			log.Fatal(err)
		}

		// create a client
		client := internal.NewClient(conn, hub, sess.Values["username"].(string))
		hub.RegisterClient(client)
		client.Run()

		return nil
	})

	e.Logger.Fatal(e.Start("localhost:8000"))
}
