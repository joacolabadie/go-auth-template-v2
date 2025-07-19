package auth

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func SetAuthCookies(c echo.Context, environment string, accessTokenString, refreshTokenString string, accessTokenTTL, refreshTokenTTL time.Duration) {
	isProd := environment == "production"

	if accessTokenString != "" {
		accessTokenCookie := &http.Cookie{
			Name:     "access_token",
			Value:    accessTokenString,
			Path:     "/",
			Expires:  time.Now().Add(accessTokenTTL),
			MaxAge:   int(accessTokenTTL.Seconds()),
			Secure:   isProd,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		}
		c.SetCookie(accessTokenCookie)
	}

	if refreshTokenString != "" {
		refreshTokenCookie := &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshTokenString,
			Path:     "/",
			Expires:  time.Now().Add(refreshTokenTTL),
			MaxAge:   int(refreshTokenTTL.Seconds()),
			Secure:   isProd,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		}
		c.SetCookie(refreshTokenCookie)
	}
}

func ClearAuthCookies(c echo.Context, environment string) {
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
