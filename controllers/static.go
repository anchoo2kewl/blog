package controllers

import (
	"net/http"
	"os"
	"strconv"
	
	"anshumanbiswas.com/blog/models"
	"anshumanbiswas.com/blog/utils"
)

func StaticHandler(tpl Template, sessionService *models.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create data structure with required fields for modern template
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
		
		// Check if user is logged in
		user, err := utils.IsUserLoggedIn(r, sessionService)
		if err != nil {
			// User not logged in
			data.LoggedIn = false
			data.IsAdmin = false
			data.Email = ""
			data.Username = ""
			data.UserPermissions = models.GetPermissions(models.RoleCommenter)
		} else {
			// User is logged in
			data.LoggedIn = true
			data.Email = user.Email
			data.Username = user.Username
			data.IsAdmin = (user.Role == 2) // Role 2 is Administrator
			data.UserPermissions = models.GetPermissions(user.Role)
		}
		
		data.SignupDisabled, _ = strconv.ParseBool(os.Getenv("APP_DISABLE_SIGNUP"))
		
		// Set page-specific values based on the URL
		switch r.URL.Path {
		case "/about":
			data.Description = "About Anshuman Biswas - Software Engineering Leader"
			data.CurrentPage = "about"
		case "/admin/formatting-guide":
			data.Description = "Content Formatting Guide - Anshuman Biswas Blog"
			data.CurrentPage = "admin"
		default:
			data.Description = "Anshuman Biswas Blog - Engineering Insights"
			data.CurrentPage = "static"
		}
		
		tpl.Execute(w, r, data)
	}
}
