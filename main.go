// File: main.go
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	// I nostri package interni
	"todolist-api-v2/internal/http/handler"
	"todolist-api-v2/internal/store"
)

func main() {
	// Inizializza lo store, che caricherà i dati da "todos.json".
	todoStore, err := store.New("todos.json")
	if err != nil {
		log.Fatalf("Errore nell'inizializzare lo store: %v", err)
	}

	// Inizializza l'handler, passandogli lo store.
	todoHandler := handler.NewTodoHandler(todoStore)

	// Inizializza il router Chi.
	r := chi.NewRouter()

	// Aggiunge dei Middleware standard di Chi.
	r.Use(middleware.RequestID)                 // Aggiunge un ID univoco a ogni richiesta.
	r.Use(middleware.RealIP)                    // Usa l'IP reale del client.
	r.Use(middleware.Logger)                    // Logga ogni richiesta in modo strutturato.
	r.Use(middleware.Recoverer)                 // Recupera da panic e risponde con un 500.
	r.Use(middleware.Timeout(60 * time.Second)) // Timeout per le richieste.

	// Definiamo le nostre rotte (le API).
	r.Route("/todos", func(r chi.Router) {
		r.Get("/", todoHandler.GetAll)  // GET /todos
		r.Post("/", todoHandler.Create) // POST /todos

		// Sotto-router per percorsi con un ID.
		r.Route("/{todoID}", func(r chi.Router) {
			r.Get("/", todoHandler.GetByID)   // GET /todos/123
			r.Put("/", todoHandler.Update)    // PUT /todos/123
			r.Delete("/", todoHandler.Delete) // DELETE /todos/123
		})
	})

	log.Println("Server in ascolto su http://localhost:8080")
	http.ListenAndServe(":8080", r)
}
