// blog_controller.go
package controllers

import (
    "fmt"
    "net/http"
    "strings"
 
    "anshumanbiswas.com/blog/models"
    "anshumanbiswas.com/blog/utils"
    "github.com/go-chi/chi/v5"
)

type Blog struct {
	Templates struct {
		Post Template
	}
	BlogService    *models.BlogService
	SessionService *models.SessionService
}

func (b *Blog) GetBlogPost(w http.ResponseWriter, r *http.Request) {

	var data struct {
		LoggedIn      bool
		Email         string
		Username      string
		IsAdmin       bool
		SignupDisabled bool
		Description   string
		CurrentPage   string
		ReadTime      string
		FullURL       string
		Post          *models.Post
		PrevPost      *models.Post
		NextPost      *models.Post
		UserPermissions models.UserPermissions
	}

	// Extract the slug from the URL
	slug := chi.URLParam(r, "slug")

	fmt.Println("Slug:", slug)
	// Fetch the blog post using the BlogService
    post, err := b.BlogService.GetBlogPostBySlug(slug)
	if err != nil {
		// Handle error (e.g., render a 404 page)
		http.NotFound(w, r)
		return
	}

	fmt.Println("Post:", post)

	// Initialize default data
	data.LoggedIn = false
	data.Post = post
	data.SignupDisabled = true // Default based on environment 
	data.Description = fmt.Sprintf("%s - Anshuman Biswas Blog", post.Title)
	data.CurrentPage = "blog"
	data.FullURL = fmt.Sprintf("http://localhost:22222/blog/%s", slug)
	
	// Set prev/next posts to nil for now (can be implemented later)
	data.PrevPost = nil
	data.NextPost = nil
	
	// Calculate reading time (simple estimation: ~200 words per minute)
	wordCount := len(strings.Fields(post.Content))
	readingMinutes := (wordCount + 199) / 200 // Round up
	if readingMinutes < 1 {
		readingMinutes = 1
	}
	data.ReadTime = fmt.Sprintf("%d", readingMinutes)

	if post.ID == 0 {
		// Handle case where post is not found
		fmt.Println("Post not found")
		http.Redirect(w, r, "/404", http.StatusFound)
		return
	}

    // Fix featured image URL if it's relative
	if post.FeaturedImageURL != "" && !strings.HasPrefix(post.FeaturedImageURL, "http") {
		// Make it a proper static URL
		if post.FeaturedImageURL == "image.jpg" {
			post.FeaturedImageURL = "/static/placeholder-featured.svg"
		} else if !strings.HasPrefix(post.FeaturedImageURL, "/static/") {
			post.FeaturedImageURL = "/static/" + post.FeaturedImageURL
		}
	}
    // ContentHTML is already prepared by BlogService (Markdown -> HTML, list/blockquote tweaks)

	user, _ := utils.IsUserLoggedIn(r, b.SessionService)
	fmt.Print(user)
	if user != nil {
		data.LoggedIn = true
		data.Email = user.Email
		data.Username = user.Username
		data.IsAdmin = (user.Role == 2) // Administrator role
		data.UserPermissions = models.GetPermissions(user.Role)
	}
	// Render the blog post template with the retrieved data
	// Example: b.Templates.BlogPost.Execute(w, r, post)
	b.Templates.Post.Execute(w, r, data)
}
