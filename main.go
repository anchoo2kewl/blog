// Path: main.go
package main

import (
	"flag"
	"log"
	"net/http"

	"anshumanbiswas.com/blog/controllers"
	"anshumanbiswas.com/blog/templates"
	"anshumanbiswas.com/blog/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	sugar := logger.Sugar()
	defer logger.Sync()

	listenAddr := flag.String("listen-addr", ":3000", "server listen address")
	flag.Parse()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", controllers.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "home.gohtml", "tailwind.gohtml"))))

	sugar.Infof("server listening on %s", *listenAddr)
	http.ListenAndServe(*listenAddr, r)

}
