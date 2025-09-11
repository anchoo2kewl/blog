// Path: main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"anshumanbiswas.com/blog/controllers"
	authmw "anshumanbiswas.com/blog/middleware"
	"anshumanbiswas.com/blog/models"
	"anshumanbiswas.com/blog/templates"
	"anshumanbiswas.com/blog/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func getAppPort() string {
	port := os.Getenv("APP_PORT")
	if port == "" {
		return "3000"
	}
	return port
}

func main() {
	sugar := sugarLog()

	apiToken := os.Getenv("API_TOKEN")

	if apiToken == "" {
		log.Fatal("API token not set in environment variable: API_TOKEN")
	} else {
		sugar.Infof("API Token: %s", apiToken)
	}

	listenAddr := flag.String("listen-addr", ":"+getAppPort(), "server listen address")
	flag.Parse()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	dbUser, dbPassword, dbName, dbHost, dbPort :=
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASSWORD"),
		os.Getenv("PG_DB"),
		os.Getenv("PG_HOST"),
		os.Getenv("PG_PORT")

	database, err := Initialize(dbUser, dbPassword, dbName, dbHost, dbPort)

	if err != nil {
		log.Fatalf("Could not set up database: %v", err)
	}
	defer database.Conn.Close()

	userService := models.UserService{
		DB: DB,
	}

	sessionService := models.SessionService{
		DB: DB,
	}

	apiTokenService := models.APITokenService{
		DB: DB,
	}

	r.Get("/about", controllers.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "about.gohtml", "tailwind.gohtml")), &sessionService))

	// Public docs routes to the formatting guide
	r.Get("/docs/formatting-guide", controllers.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "admin-formatting-guide.gohtml", "tailwind.gohtml")), &sessionService))
	r.Get("/docs/complete-formatting-guide", controllers.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "admin-formatting-guide.gohtml", "tailwind.gohtml")), &sessionService))

	r.Get("/admin/formatting-guide", controllers.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "admin-formatting-guide.gohtml", "tailwind.gohtml")), &sessionService))

	r.Get("/docs/formatting-guide", controllers.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "admin-formatting-guide.gohtml", "tailwind.gohtml")), &sessionService))

	r.Get("/docs/complete-formatting-guide", controllers.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "complete-formatting-guide.gohtml", "tailwind.gohtml")), &sessionService))

	postService := models.PostService{
		DB: DB,
	}

	// Initialize BlogService
	blogService := models.NewBlogService(DB)

	// Initialize CategoryService
	categoryService := models.CategoryService{
		DB: DB,
	}

	// Initialize SlideService
	slideService := models.SlideService{
		DB: DB,
	}

	// Setup our controllers
	usersC := controllers.Users{
		UserService:     &userService,
		SessionService:  &sessionService,
		PostService:     &postService,
		APITokenService: &apiTokenService,
		CategoryService: &categoryService,
	}

	// Initialize Blog controller
	blogC := controllers.Blog{
		BlogService:    blogService,
		SessionService: &sessionService,
	}

	// Initialize Categories controller
	categoriesC := controllers.Categories{
		CategoryService: &categoryService,
		SessionService:  &sessionService,
	}

	// Initialize Slides controller
	slidesC := controllers.Slides{
		SlideService:    &slideService,
		SessionService:  &sessionService,
		CategoryService: &categoryService,
	}

	usersC.Templates.New = views.Must(views.ParseFS(
		templates.FS, "signup.gohtml", "tailwind.gohtml"))

	isSignupDisabled, _ := strconv.ParseBool(os.Getenv("APP_DISABLE_SIGNUP"))

	if isSignupDisabled {
		fmt.Println("Signups Disabled ...")
		r.Get("/signup", usersC.Disabled)
	} else {
		fmt.Println("Signups Enabled ...")
		r.Get("/signup", usersC.New)
		r.Post("/signup", usersC.Create)
	}

	usersC.Templates.SignIn = views.Must(views.ParseFS(
		templates.FS, "signin.gohtml", "tailwind.gohtml"))

	usersC.Templates.LoggedIn = views.Must(views.ParseFS(
		templates.FS, "home.gohtml", "tailwind.gohtml"))

	r.Get("/signin", usersC.SignIn)
	r.Post("/signin", usersC.ProcessSignIn)

	usersC.Templates.Home = views.Must(views.ParseFS(
		templates.FS, "home.gohtml", "tailwind.gohtml"))

	usersC.Templates.Profile = views.Must(views.ParseFS(
		templates.FS, "profile.gohtml", "tailwind.gohtml"))

	usersC.Templates.AdminPosts = views.Must(views.ParseFS(
		templates.FS, "admin-posts.gohtml", "tailwind.gohtml"))

	usersC.Templates.UserPosts = views.Must(views.ParseFS(
		templates.FS, "user-posts.gohtml", "tailwind.gohtml"))

	usersC.Templates.APIAccess = views.Must(views.ParseFS(
		templates.FS, "api-access.gohtml", "tailwind.gohtml"))

	usersC.Templates.PostEditor = views.Must(views.ParseFS(
		templates.FS, "post-editor.gohtml", "tailwind.gohtml"))

	categoriesC.Templates.Manage = views.Must(views.ParseFS(
		templates.FS, "admin-categories.gohtml", "tailwind.gohtml"))

	// Initialize Slides templates
	slidesC.Templates.AdminSlides = views.Must(views.ParseFS(
		templates.FS, "admin-slides.gohtml", "tailwind.gohtml"))
	
	slidesC.Templates.SlideEditor = views.Must(views.ParseFS(
		templates.FS, "slide-editor.gohtml", "tailwind.gohtml"))
	
	slidesC.Templates.SlidesList = views.Must(views.ParseFS(
		templates.FS, "slides-list.gohtml", "tailwind.gohtml"))
	
	slidesC.Templates.SlidePresentation = views.Must(views.ParseFS(
		templates.FS, "slide-presentation.gohtml", "tailwind.gohtml"))

	r.Get("/", usersC.Home)
	r.Get("/admin/posts", usersC.AdminPosts)
	r.Get("/admin/posts/new", usersC.NewPost)
	r.Post("/admin/posts", usersC.CreatePost)
	r.Post("/admin/posts/from-file", usersC.CreatePostFromFile)
	r.Get("/admin/posts/{postID}/edit", usersC.EditPost)
	r.Post("/admin/posts/{postID}", usersC.UpdatePost)
	r.Post("/admin/uploads", usersC.UploadImage)
	r.Post("/admin/uploads/multiple", usersC.UploadMultipleImages)
	r.Get("/admin/uploads/list", usersC.ListUploadedImages)
	r.Delete("/admin/uploads", usersC.DeleteImage)
	r.Post("/admin/preview", usersC.PreviewRender)
	r.Get("/my-posts", usersC.UserPosts)
	r.Get("/api-access", usersC.APIAccess)

	// Category Management Routes
	r.Get("/admin/categories", categoriesC.Manage)
	r.Post("/admin/categories", categoriesC.CreateCategoryForm)
	r.Post("/admin/categories/{id}", categoriesC.UpdateCategoryForm)
	r.Post("/admin/categories/{id}/delete", categoriesC.DeleteCategoryForm)

	// Slides Routes
	r.Get("/slides", slidesC.PublicSlidesList)
	r.Get("/slides/{slug}", slidesC.ViewSlide)
	
	// Admin Slides Routes
	r.Get("/admin/slides", slidesC.AdminSlides)
	r.Get("/admin/slides/new", slidesC.NewSlide)
	r.Post("/admin/slides", slidesC.CreateSlide)
	r.Get("/admin/slides/{slideID}/edit", slidesC.EditSlide)
	r.Post("/admin/slides/{slideID}", slidesC.UpdateSlide)
	r.Post("/admin/slides/{slideID}/delete", slidesC.DeleteSlide)
	r.Post("/admin/slides/preview", slidesC.PreviewSlide)

	r.Get("/users/me", usersC.CurrentUser)
	r.Post("/users/password", usersC.UpdatePassword)
	r.Post("/users/email", usersC.UpdateEmail)
	r.Post("/users/api-tokens", usersC.CreateAPIToken)
	r.Post("/users/api-tokens/revoke", usersC.RevokeAPIToken)
	r.Post("/users/api-tokens/delete", usersC.DeleteAPIToken)

	// JSON API endpoints for AJAX operations
	r.Post("/api/users/api-tokens", usersC.CreateAPITokenJSON)
	r.Post("/api/users/api-tokens/revoke", usersC.RevokeAPITokenJSON)
	r.Delete("/api/users/api-tokens/{token_id}", usersC.DeleteAPITokenJSON)
	r.Get("/api/users/api-tokens", usersC.GetAPITokensJSON)
	r.Get("/users/logout", usersC.Logout)

	// Logout redirect route for convenience
	r.Get("/logout", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/users/logout", http.StatusFound)
	})

	blogC.Templates.Post = views.Must(views.ParseFS(
		templates.FS, "blogpost.gohtml", "tailwind.gohtml"))

	// Define a route for the blog post
	r.Get("/blog/{slug}", blogC.GetBlogPost)

	// Public API for lazy loading posts
	r.Get("/api/posts/load-more", usersC.LoadMorePosts)

	// REST API endpoints for users
	r.Route("/api/users", func(r chi.Router) {
		r.Use(authmw.APIAuthMiddleware(apiToken, &apiTokenService))
		r.Get("/", usersC.ListUsers)
		r.Post("/", usersC.CreateUser)
	})

	r.Route("/api/posts", func(r chi.Router) {
		r.Use(authmw.APIAuthMiddleware(apiToken, &apiTokenService))
		r.Get("/", getAllPosts)
		r.Get("/{postID}", getPostByID)
		r.Post("/", createPost)
		r.Post("/from-file", usersC.CreatePostFromFile)
	})

	r.Route("/api/categories", func(r chi.Router) {
		r.Use(authmw.APIAuthMiddleware(apiToken, &apiTokenService))
		r.Get("/", categoriesC.ListCategories)
		r.Post("/", categoriesC.CreateCategory)
		r.Get("/{id}", categoriesC.GetCategory)
		r.Put("/{id}", categoriesC.UpdateCategory)
		r.Delete("/{id}", categoriesC.DeleteCategory)
	})

	// Define a custom 404 handler
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		// You can render a custom 404 page here
		// For simplicity, let's just return a plain text response
		http.ServeFile(w, r, "templates/NotFoundPage.gohtml")
	})

	sugar.Infof("server listening on %s", *listenAddr)

	// Serve favicon at root level for both GET and HEAD requests
	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		http.ServeFile(w, r, "./static/favicon.svg")
	})

	// Serve static files from ./static/ directory
	staticFileServer := http.FileServer(http.Dir("./static/"))
	r.Handle("/static/*", http.StripPrefix("/static/", staticFileServer))

	// Keep legacy CSS route for backward compatibility
	cssFileServer := http.FileServer(http.Dir("./css/"))
	r.Handle("/css/*", http.StripPrefix("/css/", cssFileServer))

	http.ListenAndServe(*listenAddr, r)
}

// AuthMiddleware is a middleware function to check API token
func AuthMiddleware(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			tokenReceived := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenReceived != token {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getAllPosts(w http.ResponseWriter, r *http.Request) {

	postService := models.PostService{
		DB: DB,
	}

	posts, err := postService.GetTopPosts()
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}
	// Send the posts as JSON response
	jsonResponse(w, posts, http.StatusOK)
}

func getPostByID(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "postID")
	post := postID
	// Fetch post from the database using postID
	// Implement this logic based on your database schema
	// Example: post, err := postService.GetPostByID(postID)
	// Handle errors and send appropriate response
	// Send the post as JSON response
	jsonResponse(w, post, http.StatusOK)
}

func createPost(w http.ResponseWriter, r *http.Request) {

	postService := models.PostService{
		DB: DB,
	}

	newPost := models.Post{}
	// Decode the JSON request to newPost
	err := json.NewDecoder(r.Body).Decode(&newPost)
	if err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	// Create a new post using the postService
	post, _ := postService.Create(newPost.UserID, newPost.CategoryID, newPost.Title, newPost.Content, newPost.IsPublished, newPost.FeaturedImageURL, newPost.Slug)

	// Handle errors and send appropriate response
	// Send the created post as JSON response
	jsonResponse(w, post, http.StatusCreated)
}

// jsonResponse sends a JSON response with the given data and status code.
func jsonResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
