package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"anshumanbiswas.com/blog/models"
)

func (u Users) New(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email            string
		Username         string
		LoggedIn         bool
		IsSignupDisabled bool
	}

	data.Email = r.FormValue("email")
	data.LoggedIn = false
	data.IsSignupDisabled = false
	u.Templates.New.Execute(w, r, data)
}

func (u Users) Disabled(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email            string
		LoggedIn         bool
		IsSignupDisabled bool
	}
	data.Email = r.FormValue("email")
	data.LoggedIn = false
	data.IsSignupDisabled = true
	u.Templates.New.Execute(w, r, data)
}

type Users struct {
	Templates struct {
		New      Template
		SignIn   Template
		Home     Template
		LoggedIn Template
	}
	UserService    *models.UserService
	SessionService *models.SessionService
	PostService    *models.PostService
}

func (u Users) GetTopPosts() (*models.PostsList, error) {
	return u.PostService.GetTopPosts()
}

func (u Users) Home(w http.ResponseWriter, r *http.Request) {

	var data struct {
		Email    string
		LoggedIn bool
		Posts    *models.PostsList
	}

	posts, _ := u.GetTopPosts()

	user, err := u.isUserLoggedIn(r)
	if err != nil {
		data.LoggedIn = false
		data.Posts = posts
		u.Templates.Home.Execute(w, r, data)
		return
	}

	data.Email = user.Email
	data.LoggedIn = true
	data.Posts = posts
	u.Templates.Home.Execute(w, r, data)
}

func (u Users) SignIn(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email    string
		LoggedIn bool
	}
	data.Email = r.FormValue("email")
	data.LoggedIn = false
	u.Templates.SignIn.Execute(w, r, data)
}

func (u Users) ProcessSignIn(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email    string
		Password string
	}
	data.Email = r.FormValue("email")
	data.Password = r.FormValue("password")

	user, err := u.UserService.Authenticate(data.Email, data.Password)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	session, err := u.SessionService.Create(user.UserID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	setCookie(w, CookieSession, session.Token)
	setCookie(w, CookieUserEmail, data.Email)

	http.Redirect(w, r, "/users/me", http.StatusFound)
}

func (u Users) Create(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	username := r.FormValue("username")
	password := r.FormValue("password")
	fmt.Printf("[Creating user: %s/%s]", email, username)
	user, err := u.UserService.Create(email, username, password, 1)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	session, err := u.SessionService.Create(user.UserID)
	if err != nil {
		fmt.Println(err)
		// TODO: Long term, we should show a warning about not being able to sign the user in.
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	setCookie(w, CookieSession, session.Token)
	setCookie(w, CookieUserEmail, email)

	http.Redirect(w, r, "/users/me", http.StatusFound)
}

func (u Users) isUserLoggedIn(r *http.Request) (*models.User, error) {
	token, err := readCookie(r, CookieSession)
	email, err := readCookie(r, CookieUserEmail)

	if err != nil {
		return nil, err
	}
	return u.SessionService.User(token, email)
}

func (u Users) CurrentUser(w http.ResponseWriter, r *http.Request) {

	user, err := u.isUserLoggedIn(r)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	var data struct {
		Email    string
		LoggedIn bool
		Posts    *models.PostsList
	}

	posts, err := u.GetTopPosts()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	data.Email = user.Email
	data.LoggedIn = true
	data.Posts = posts

	u.Templates.LoggedIn.Execute(w, r, data)
}

func (u Users) Logout(w http.ResponseWriter, r *http.Request) {

	email, err := readCookie(r, CookieUserEmail)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	u.SessionService.Logout(email)

	deleteCookie(w, CookieSession, "XXXXXX")
	deleteCookie(w, CookieUserEmail, "XXXXXXX")

	http.Redirect(w, r, "/", http.StatusFound)

}

func (u Users) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := u.UserService.GetAllUsers()
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (u Users) CreateUser(w http.ResponseWriter, r *http.Request) {
	var newUser models.User

	// Parse request body
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create user
	user, err := u.UserService.Create(newUser.Email, newUser.Username, newUser.PasswordHash, newUser.Role)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Return created user
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
