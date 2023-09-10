// Path: main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

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

	r.Get("/", controllers.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "home.gohtml", "tailwind.gohtml"))))
	r.Get("/about", controllers.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "about.gohtml", "tailwind.gohtml"))))

	userService := models.UserService{
		DB: DB,
	}

	sessionService := models.SessionService{
		DB: DB,
	}

	// Setup our controllers
	usersC := controllers.Users{
		UserService:    &userService,
		SessionService: &sessionService,
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

	r.Get("/signin", usersC.SignIn)
	r.Post("/signin", usersC.ProcessSignIn)

	r.Get("/users/me", usersC.CurrentUser)
	r.Get("/users/logout", usersC.Logout)

	sugar.Infof("server listening on %s", *listenAddr)
	http.ListenAndServe(*listenAddr, r)
}
