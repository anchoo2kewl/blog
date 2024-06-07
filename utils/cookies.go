// utils/cookies.go
package utils

import (
	"net/http"
)

var (
	CookieSession   = "session"
	CookieUserEmail = "user_email"
)

func ReadCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
