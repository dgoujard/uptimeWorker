package router

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"

	"github.com/dgoujard/uptimeWorker/app/app"
	"github.com/dgoujard/uptimeWorker/app/requestlog"
)

func New(a *app.App) *chi.Mux {
	l := a.Logger()

	r := chi.NewRouter()

	rcors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	r.Use(rcors.Handler)
	// Routes for healthz
	r.Get("/healthz", app.HandleHealth)

	// Routes for APIs
	r.Method("GET","/",requestlog.NewHandler(a.HandleIndex,l))

	return r
}
