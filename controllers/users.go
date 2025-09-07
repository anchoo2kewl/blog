package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"anshumanbiswas.com/blog/models"
	"anshumanbiswas.com/blog/utils"
)

func (u Users) New(w http.ResponseWriter, r *http.Request) {
    var data struct {
        Email            string
        Username         string
        LoggedIn         bool
        IsSignupDisabled bool
        SignupDisabled   bool
        IsAdmin          bool
        Description      string
        CurrentPage      string
        UserPermissions  models.UserPermissions
    }

	data.Email = r.FormValue("email")
	data.LoggedIn = false
	data.IsSignupDisabled = false
	data.SignupDisabled = false
	data.IsAdmin = false
	data.Description = "Sign up for Anshuman Biswas Blog"
	data.CurrentPage = "signup"
    data.UserPermissions = models.GetPermissions(models.RoleCommenter)
    u.Templates.New.Execute(w, r, data)
}

func (u Users) Disabled(w http.ResponseWriter, r *http.Request) {
    var data struct {
        Email            string
        LoggedIn         bool
        IsSignupDisabled bool
        SignupDisabled   bool
        IsAdmin          bool
        Description      string
        CurrentPage      string
        Username         string
        UserPermissions  models.UserPermissions
    }
	data.Email = r.FormValue("email")
	data.LoggedIn = false
	data.IsSignupDisabled = true
	data.SignupDisabled = true
	data.IsAdmin = false
	data.Description = "Sign up disabled - Anshuman Biswas Blog"
	data.CurrentPage = "signup"
	data.Username = ""
    data.UserPermissions = models.GetPermissions(models.RoleCommenter)
    u.Templates.New.Execute(w, r, data)
}

type Users struct {
	Templates struct {
		New       Template
		SignIn    Template
		Home      Template
		LoggedIn  Template
		Profile   Template
		AdminPosts Template
		UserPosts Template
		APIAccess Template
	}
	UserService      *models.UserService
	SessionService   *models.SessionService
	PostService      *models.PostService
	APITokenService  *models.APITokenService
}

func (u Users) GetTopPosts() (*models.PostsList, error) {
	return u.PostService.GetTopPosts()
}

func (u Users) Home(w http.ResponseWriter, r *http.Request) {

    var data struct {
        Email           string
        LoggedIn        bool
        Posts           *models.PostsList
        Username        string
        IsAdmin         bool
        SignupDisabled  bool
        Description     string
        CurrentPage     string
        UserPermissions models.UserPermissions
    }

	posts, _ := u.GetTopPosts()

	// Get signup disabled setting from environment
	isSignupDisabled, _ := strconv.ParseBool(os.Getenv("APP_DISABLE_SIGNUP"))

	user, err := u.isUserLoggedIn(r)
	if err != nil {
		fmt.Printf("DEBUG HOME: User not logged in, error: %v\n", err)
		fmt.Printf("DEBUG HOME: Session token from cookie: %v\n", func() string {
			if cookie, err := r.Cookie("session"); err == nil {
				return cookie.Value
			}
			return "NO_COOKIE"
		}())
		data.LoggedIn = false
		data.Posts = posts
		data.SignupDisabled = isSignupDisabled
		data.Description = "Engineering Insights - Anshuman Biswas Blog"
		data.CurrentPage = "home"
        data.Username = ""
        data.IsAdmin = false
        data.Email = ""
        data.UserPermissions = models.GetPermissions(models.RoleCommenter)
        u.Templates.Home.Execute(w, r, data)
        return
    }

	data.Email = user.Email
	data.Username = user.Username
	data.LoggedIn = true
	data.Posts = posts
	data.IsAdmin = models.IsAdmin(user.Role)
	fmt.Printf("DEBUG HOME: User logged in: %s, Email: %s, Role: %d, IsAdmin: %v\n", user.Username, user.Email, user.Role, data.IsAdmin)
	data.SignupDisabled = isSignupDisabled
	data.Description = "Engineering Insights - Anshuman Biswas Blog"
	data.CurrentPage = "home"
    data.UserPermissions = models.GetPermissions(user.Role)
    u.Templates.Home.Execute(w, r, data)
}

func (u Users) SignIn(w http.ResponseWriter, r *http.Request) {
    var data struct {
        Email           string
        LoggedIn        bool
        SignupDisabled  bool
        IsAdmin         bool
        Description     string
        CurrentPage     string
        Username        string
        UserPermissions models.UserPermissions
    }
	data.Email = r.FormValue("email")
	data.LoggedIn = false
	data.SignupDisabled, _ = strconv.ParseBool(os.Getenv("APP_DISABLE_SIGNUP"))
	data.IsAdmin = false
	data.Description = "Sign in to Anshuman Biswas Blog"
	data.CurrentPage = "signin"
	data.Username = ""
    data.UserPermissions = models.GetPermissions(models.RoleCommenter)
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

	http.Redirect(w, r, "/", http.StatusFound)
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

	http.Redirect(w, r, "/", http.StatusFound)
}

func (u Users) isUserLoggedIn(r *http.Request) (*models.User, error) {
	return utils.IsUserLoggedIn(r, u.SessionService)
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
        Email           string
        LoggedIn        bool
        Username        string
        IsAdmin         bool
        SignupDisabled  bool
        Description     string
        CurrentPage     string
        Message         string
        UserPermissions models.UserPermissions
    }

	data.Email = user.Email
	data.Username = user.Username
	data.LoggedIn = true
	data.IsAdmin = models.IsAdmin(user.Role)
	data.SignupDisabled, _ = strconv.ParseBool(os.Getenv("APP_DISABLE_SIGNUP"))
	data.Description = "Profile Management - Anshuman Biswas Blog"
	data.CurrentPage = "profile"
    data.Message = r.URL.Query().Get("message")
    data.UserPermissions = models.GetPermissions(user.Role)

	u.Templates.Profile.Execute(w, r, data)
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
	user, err := u.UserService.Create(newUser.Email, newUser.Username, newUser.Password, newUser.Role)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Return created user
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (u Users) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	user, err := u.isUserLoggedIn(r)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	currentPassword := r.FormValue("current_password")
	newPassword := r.FormValue("new_password")
	confirmPassword := r.FormValue("confirm_password")

	if newPassword != confirmPassword {
		http.Redirect(w, r, "/users/me?message=Passwords do not match", http.StatusFound)
		return
	}

	// Verify current password
	_, err = u.UserService.Authenticate(user.Email, currentPassword)
	if err != nil {
		http.Redirect(w, r, "/users/me?message=Current password is incorrect", http.StatusFound)
		return
	}

	// Update password
	err = u.UserService.UpdatePassword(user.UserID, newPassword)
	if err != nil {
		http.Redirect(w, r, "/users/me?message=Failed to update password", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/users/me?message=Password updated successfully", http.StatusFound)
}

func (u Users) UpdateEmail(w http.ResponseWriter, r *http.Request) {
	user, err := u.isUserLoggedIn(r)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	newEmail := r.FormValue("new_email")
	password := r.FormValue("password")

	// Verify password
	_, err = u.UserService.Authenticate(user.Email, password)
	if err != nil {
		http.Redirect(w, r, "/users/me?message=Password is incorrect", http.StatusFound)
		return
	}

	// Update email
	err = u.UserService.UpdateEmail(user.UserID, newEmail)
	if err != nil {
		http.Redirect(w, r, "/users/me?message=Failed to update email", http.StatusFound)
		return
	}

	// Update cookie with new email
	setCookie(w, CookieUserEmail, newEmail)
	
	http.Redirect(w, r, "/users/me?message=Email updated successfully", http.StatusFound)
}

// AdminPosts shows all posts for admin users
func (u Users) AdminPosts(w http.ResponseWriter, r *http.Request) {
	user, err := u.isUserLoggedIn(r)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	// Check if user is admin
	if !models.IsAdmin(user.Role) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	posts, err := u.PostService.GetAllPosts()
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

    var data struct {
        Email           string
        LoggedIn        bool
        Username        string
        IsAdmin         bool
        SignupDisabled  bool
        Description     string
        CurrentPage     string
        Posts           *models.PostsList
        UserPermissions models.UserPermissions
    }

	data.Email = user.Email
	data.Username = user.Username
	data.LoggedIn = true
	data.IsAdmin = true
	data.SignupDisabled, _ = strconv.ParseBool(os.Getenv("APP_DISABLE_SIGNUP"))
	data.Description = "Manage All Posts - Anshuman Biswas Blog"
	data.CurrentPage = "admin-posts"
    data.Posts = posts
    data.UserPermissions = models.GetPermissions(user.Role)

	u.Templates.AdminPosts.Execute(w, r, data)
}

// UserPosts shows posts for the current user
func (u Users) UserPosts(w http.ResponseWriter, r *http.Request) {
	user, err := u.isUserLoggedIn(r)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	posts, err := u.PostService.GetPostsByUser(user.UserID)
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

    var data struct {
        Email           string
        LoggedIn        bool
        Username        string
        IsAdmin         bool
        SignupDisabled  bool
        Description     string
        CurrentPage     string
        Posts           *models.PostsList
        UserPermissions models.UserPermissions
    }

	data.Email = user.Email
	data.Username = user.Username
	data.LoggedIn = true
	data.IsAdmin = (user.Role == 2)
	data.SignupDisabled, _ = strconv.ParseBool(os.Getenv("APP_DISABLE_SIGNUP"))
	data.Description = "My Posts - Anshuman Biswas Blog"
	data.CurrentPage = "my-posts"
    data.Posts = posts
    data.UserPermissions = models.GetPermissions(user.Role)

	u.Templates.UserPosts.Execute(w, r, data)
}

// APIAccess shows the API access management page
func (u Users) APIAccess(w http.ResponseWriter, r *http.Request) {
	user, err := u.isUserLoggedIn(r)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	tokens, err := u.APITokenService.GetByUser(user.UserID)
	if err != nil {
		http.Error(w, "Failed to fetch API tokens", http.StatusInternalServerError)
		return
	}

	var data struct {
		Email           string
		LoggedIn        bool
		Username        string
		IsAdmin         bool
		SignupDisabled  bool
		Description     string
		CurrentPage     string
		Message         string
		Tokens          []*models.APIToken
		UserPermissions models.UserPermissions
	}

	data.Email = user.Email
	data.Username = user.Username
	data.LoggedIn = true
	data.IsAdmin = models.IsAdmin(user.Role)
	data.SignupDisabled, _ = strconv.ParseBool(os.Getenv("APP_DISABLE_SIGNUP"))
	data.Description = "API Access Management - Anshuman Biswas Blog"
	data.CurrentPage = "api-access"
	data.Message = r.URL.Query().Get("message")
	data.Tokens = tokens
	data.UserPermissions = models.GetPermissions(user.Role)

	u.Templates.APIAccess.Execute(w, r, data)
}

// CreateAPIToken creates a new API token for the user
func (u Users) CreateAPIToken(w http.ResponseWriter, r *http.Request) {
	user, err := u.isUserLoggedIn(r)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	tokenName := r.FormValue("name")
	if tokenName == "" {
		http.Redirect(w, r, "/api-access?message=Token name is required", http.StatusFound)
		return
	}

	token, err := u.APITokenService.Create(user.UserID, tokenName, nil)
	if err != nil {
		http.Redirect(w, r, "/api-access?message=Failed to create API token", http.StatusFound)
		return
	}

	// For security, we show the token only once after creation
	http.Redirect(w, r, fmt.Sprintf("/api-access?message=Token created successfully: %s&new_token=%s", tokenName, token.Token), http.StatusFound)
}

// RevokeAPIToken revokes an API token
func (u Users) RevokeAPIToken(w http.ResponseWriter, r *http.Request) {
	user, err := u.isUserLoggedIn(r)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	tokenIDStr := r.FormValue("token_id")
	tokenID, err := strconv.Atoi(tokenIDStr)
	if err != nil {
		http.Redirect(w, r, "/api-access?message=Invalid token ID", http.StatusFound)
		return
	}

	err = u.APITokenService.Revoke(tokenID, user.UserID)
	if err != nil {
		http.Redirect(w, r, "/api-access?message=Failed to revoke token", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/api-access?message=Token revoked successfully", http.StatusFound)
}

// DeleteAPIToken deletes an API token
func (u Users) DeleteAPIToken(w http.ResponseWriter, r *http.Request) {
	user, err := u.isUserLoggedIn(r)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	tokenIDStr := r.FormValue("token_id")
	tokenID, err := strconv.Atoi(tokenIDStr)
	if err != nil {
		http.Redirect(w, r, "/api-access?message=Invalid token ID", http.StatusFound)
		return
	}

	err = u.APITokenService.Delete(tokenID, user.UserID)
	if err != nil {
		http.Redirect(w, r, "/api-access?message=Failed to delete token", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/api-access?message=Token deleted successfully", http.StatusFound)
}
