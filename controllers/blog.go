// blog_controller.go
package controllers

import (
	"fmt"
	"net/http"

	"anshumanbiswas.com/blog/models"
	"github.com/go-chi/chi/v5"
)

type Blog struct {
	Templates struct {
		Post Template
	}
	BlogService *models.BlogService
}

func (b *Blog) GetBlogPost(w http.ResponseWriter, r *http.Request) {

	var data struct {
		LoggedIn bool
		Post     *models.Post
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

	data.LoggedIn = false
	data.Post = post

	if post.ID == 0 {
		// Handle case where post is not found
		fmt.Println("Post not found")
		http.Redirect(w, r, "/404", http.StatusFound)
		return
	}

	// Render the blog post template with the retrieved data
	// Example: b.Templates.BlogPost.Execute(w, r, post)
	b.Templates.Post.Execute(w, r, data)
}
