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
	Password         string
	PasswordHash     string
	RegistrationDate string
	Role             int
}

type UserService struct {
	DB *sql.DB
}

func (us *UserService) Create(email, username, password string, role_id int) (*User, error) {
	email = strings.ToLower(email)

	hashedBytes, err := bcrypt.GenerateFromPassword(
		[]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	passwordHash := string(hashedBytes)

	fmt.Println("Password: ", password)
	fmt.Println("Password Hash: ", passwordHash)

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

	row := us.DB.QueryRow(`SELECT user_id, username, password, role_id FROM users WHERE email=$1`, email)
	err := row.Scan(&user.UserID, &user.Username, &user.PasswordHash, &user.Role)
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

func (us *UserService) GetAllUsers() ([]*User, error) {
	rows, err := us.DB.Query("SELECT user_id, email, username, registration_date, role_id FROM Users")
	if err != nil {
		return nil, fmt.Errorf("get all users: %w", err)
	}
	defer rows.Close()

	var users []*User

	for rows.Next() {
		var user User
		err := rows.Scan(&user.UserID, &user.Email, &user.Username, &user.RegistrationDate, &user.Role)
		if err != nil {
			return nil, fmt.Errorf("get all users: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get all users: %w", err)
	}

	return users, nil
}

func (us *UserService) UpdatePassword(userID int, newPassword string) error {
	hashedBytes, err := bcrypt.GenerateFromPassword(
		[]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	passwordHash := string(hashedBytes)

	_, err = us.DB.Exec("UPDATE Users SET password = $1 WHERE user_id = $2", passwordHash, userID)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}

func (us *UserService) UpdateEmail(userID int, newEmail string) error {
	newEmail = strings.ToLower(newEmail)
	
	_, err := us.DB.Exec("UPDATE Users SET email = $1 WHERE user_id = $2", newEmail, userID)
	if err != nil {
		return fmt.Errorf("update email: %w", err)
	}

	return nil
}
