package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UserID           int
	Username         string
	Email            string
	PasswordHash     string
	RegistrationDate string
	Role             int
}

type UserService struct {
	DB *sql.DB
}

func (us *UserService) Create(email, username, password string) (*User, error) {
	email = strings.ToLower(email)

	hashedBytes, err := bcrypt.GenerateFromPassword(
		[]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	role_id := 1 // This is assumed to be the commenter. Please verify.

	passwordHash := string(hashedBytes)

	row := us.DB.QueryRow(`
		INSERT INTO Users (email, username, password, role_id, registration_date)
		VALUES ($1, $2, $3, $4, $5) RETURNING user_id`, email, username, passwordHash, role_id, time.Now().UTC())

	user := User{
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role_id,
	}

	err = row.Scan(&user.UserID)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &user, nil
}

func (us UserService) Authenticate(email, password string) (*User, error) {
	email = strings.ToLower(email)
	user := User{
		Email: email,
	}

	row := us.DB.QueryRow(`SELECT user_id, password FROM users WHERE email=$1`, email)
	err := row.Scan(&user.UserID, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}
	return &user, nil
}

func (ss *UserService) GenerateHashedToken(token string) (string, error) {
	hashedTokenBytes, err := bcrypt.GenerateFromPassword(
		[]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("create token: %w", err)
	}

	return string(hashedTokenBytes), nil
}
