package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/zeroyukiy/the-throne-api/database/entity"
	"github.com/zeroyukiy/the-throne-api/database/repository"
	"github.com/zeroyukiy/the-throne-api/internal"
)

type AuthHandler struct {
	Repo repository.UserRepository
}

func NewAuthHandler(conn *sqlx.DB) *AuthHandler {
	repo := repository.NewUserRepository(conn)
	return &AuthHandler{
		Repo: repo,
	}
}

type LoginBinding struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login endpoint /api/auth/login
func (h *AuthHandler) Login(c echo.Context) error {
	login := &LoginBinding{}
	err := c.Bind(login)
	if err != nil {
		return echo.ErrBadRequest
	}

	// validate and clean up the form before send to the database

	// user := h.Repo.Get(login.Username, login.Password)
	// if login.Username != user.Username {
	// 	return echo.ErrNotFound
	// }

	user := entity.User{
		Username: "pippo",
		Avatar:   "http://localhost:8000/assets/avatars/avatar_wizard_elve_man_01.png",
	}

	claims := &internal.JwtCustomClaims{
		Name:   user.Username,
		Admin:  true,
		Avatar: user.Avatar,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "localhost",
			Subject:   "user",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 18)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString([]byte("my_secret_string"))
	if err != nil {
		return err
	}

	claims_2 := &internal.JwtCustomClaims{
		Name:   user.Username,
		Admin:  true,
		Avatar: user.Avatar,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "localhost",
			Subject:   "user",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	refresh_token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims_2)

	r, err := refresh_token.SignedString([]byte("my_secret_string"))
	if err != nil {
		return err
	}

	// // put the refresh_token in the cookie
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    r,
		Path:     "/",
		Domain:   "localhost",
		Secure:   true,
		Expires:  time.Now().Add(time.Hour * 24 * 30),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return c.JSON(http.StatusOK, echo.Map{
		"token": ss,
	})
}

// Refresh endpoint /api/auth/refresh
func (h *AuthHandler) Refresh(c echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "no refresh_token",
		})
	}

	claims := &internal.JwtCustomClaims{}
	token, err := jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte("my_secret_string"), nil
	})
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "error with the token",
		})
	}

	fmt.Println("claims token from cookie: ", claims)

	// if token is expired
	if claims.ExpiresAt.Unix() < time.Now().Unix() {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "refresh_token has expired, you need to login.",
		})
	}

	// TODO clean this shit
	claims = &internal.JwtCustomClaims{
		Name:   claims.Name,
		Admin:  claims.Admin,
		Avatar: claims.Avatar,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "localhost",
			Subject:   "user",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 18)),
		},
	}

	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte("my_secret_string"))
	if err != nil {
		return err
	}

	claims_2 := &internal.JwtCustomClaims{
		Name:   claims.Name,
		Admin:  claims.Admin,
		Avatar: claims.Avatar,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "localhost",
			Subject:   "user",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	refresh_token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims_2)

	r, err := refresh_token.SignedString([]byte("my_secret_string"))
	if err != nil {
		return err
	}

	// put the refresh_token in the cookie
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    r,
		Path:     "/",
		Domain:   "localhost",
		Secure:   false,
		Expires:  time.Now().Add(time.Hour * 24 * 30),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// the garbage end here

	return c.JSON(http.StatusOK, echo.Map{
		"message": "new access token and refresh token",
		"token":   t,
	})
}

// Logout endpoint /api/auth/logout
func (h *AuthHandler) Logout(c echo.Context) error {
	_, err := c.Cookie("refresh_token")
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, nil)
	}

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	return c.JSON(http.StatusOK, nil)
}
