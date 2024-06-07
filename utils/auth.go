// utils/auth.go
package utils

import (
	"net/http"

	"anshumanbiswas.com/blog/models"
)

func IsUserLoggedIn(r *http.Request, sessionService *models.SessionService) (*models.User, error) {
	token, _ := ReadCookie(r, CookieSession)
	email, err := ReadCookie(r, CookieUserEmail)

	if err != nil {
		return nil, err
	}
	return sessionService.User(token, email)
}
