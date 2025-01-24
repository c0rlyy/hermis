package utils

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// sets authToken and refresh token cookies in http only secure cookie
func SetAuthCookies(c echo.Context, token, rtToken, path string) error {
	cookie := &http.Cookie{
		Name:     "authToken",
		Value:    token,
		Expires:  time.Now().Add(5 * time.Minute), // 5 min
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}
	// sending http only cookie
	c.SetCookie(cookie)

	rtCookie := &http.Cookie{
		Name:     "refreshToken",
		Value:    rtToken,
		Expires:  time.Now().AddDate(0, 0, 7), // 7 Days
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}
	c.SetCookie(rtCookie)
	return nil
}

func UnsetAuthCookies(c echo.Context, path string) error {
	cookie := &http.Cookie{
		Name:     "authToken",
		Value:    "deleted",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}
	c.SetCookie(cookie)

	rtCookie := &http.Cookie{
		Name:     "refreshToken",
		Value:    "deleted",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}

	c.SetCookie(rtCookie)
	return nil
}
