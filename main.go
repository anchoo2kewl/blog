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
	"anshumanbiswas.com/blog/models"
	"anshumanbiswas.com/blog/templates"
	"anshumanbiswas.com/blog/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	APP_PORT = "3000"
)

func main() {
	sugar := sugarLog()

	apiToken := os.Getenv("API_TOKEN")

	if apiToken == "" {
		log.Fatal("API token not set in environment variable: API_TOKEN")
	} else {
		sugar.Infof("API Token: %s", apiToken)
	}

	listenAddr := flag.String("listen-addr", ":"+APP_PORT, "server listen address")
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

	r.Get("/about", controllers.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "about.gohtml", "tailwind.gohtml"))))

	userService := models.UserService{
		DB: DB,
	}

	sessionService := models.SessionService{
		DB: DB,
	}

	postService := models.PostService{
		DB: DB,
	}

	// Setup our controllers
	usersC := controllers.Users{
		UserService:    &userService,
		SessionService: &sessionService,
		PostService:    &postService,
	}

	usersC.Templates.New = views.Must(views.ParseFS(
		templates.FS, "signup.gohtml", "tailwind.gohtml"))

	isSignupDisabled, err := strconv.ParseBool(os.Getenv("APP_DISABLE_SIGNUP"))

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

	r.Get("/", usersC.Home)

	r.Get("/users/me", usersC.CurrentUser)
	r.Get("/users/logout", usersC.Logout)

	// REST API endpoints for users
	r.Route("/api/users", func(r chi.Router) {
		r.Use(AuthMiddleware(apiToken)) // Middleware to check token
		r.Get("/", usersC.ListUsers)
		r.Post("/", usersC.CreateUser)
	})

	r.Route("/api/posts", func(r chi.Router) {
		r.Use(AuthMiddleware(apiToken)) // Middleware to check token
		r.Get("/", getAllPosts)
		r.Get("/{postID}", getPostByID)
		r.Post("/", createPost)
	})

	sugar.Infof("server listening on %s", *listenAddr)

	fileServer := http.FileServer(http.Dir("./css/"))

	r.Handle("/css/*", http.StripPrefix("/css/", fileServer))

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
