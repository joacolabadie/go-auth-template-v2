package auth

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func ClearCookies(c echo.Context, environment string) {
	isProd := environment == "production"

	accessTokenCookie := &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Secure:   isProd,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(accessTokenCookie)

	refreshTokenCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Secure:   isProd,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(refreshTokenCookie)
}
