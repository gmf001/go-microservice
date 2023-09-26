package app

import (
	"net/http"

	"github.com/gmf001/go-microservice/handlers"
	"github.com/gmf001/go-microservice/libs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)


func (a *App) loadRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	router.Route("/orders", a.loadOrderRoutes)

	a.router = router

	return router
}

func (a *App) loadOrderRoutes(router chi.Router) {
	orderHandler := &handlers.Order{
		Repo: &libs.RedisClient{
			Client: a.rdb,
		},
	}
	router.Post("/", orderHandler.Create)
	router.Get("/", orderHandler.List)
	router.Get("/{id}", orderHandler.GetByID)
	router.Put("/{id}", orderHandler.UpdateByID)
	router.Delete("/{id}", orderHandler.DeleteByID)
}