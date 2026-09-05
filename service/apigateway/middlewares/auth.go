package middlewares

import (
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

var whiteListPaths = []string{
	"/api/auth/login",
	"/api/auth/register",
	"/api/auth/hello",
	"/api/auth/verify-code",
	"/docs/",
	"/docs",
	"/swagger",
	"/metrics",
}

func WebSecurityConfig(e *echo.Echo) {
	config := echojwt.Config{
		SigningKey: []byte(viper.GetString("SECRET_KEY")),
		Skipper:    skipAuth,
		SuccessHandler: func(c echo.Context) {
			user := c.Get("user").(*jwt.Token)

			if claims, ok := user.Claims.(jwt.MapClaims); ok {
				c.Set("user_id", fmt.Sprint(claims["sub"]))
			}
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return echo.ErrUnauthorized
		},
	}
	e.Use(echojwt.WithConfig(config))

}

func skipAuth(c echo.Context) bool {
	path := c.Request().URL.Path

	for _, p := range whiteListPaths {
		if path == p ||
			strings.HasPrefix(path, "/swagger") ||
			strings.HasPrefix(path, "/metrics") {
			return true
		}
	}

	return false
}
