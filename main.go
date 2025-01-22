package main

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Routes
	r.Get("/todos", getAllTodos)
	r.Get("/todos/{id}", getTodoByID)
	r.Post("/todos", createTodo)
	r.Put("/todos/{id}", updateTodo)
	r.Delete("/todos/{id}", deleteTodo)

	port := ":8080"
	log.Printf("HTTP server is running at %s", port)
	// Start server
	http.ListenAndServe(port, r)
}

type Todo struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

var (
	todoStore = make(map[int]Todo)
	mu        sync.RWMutex
)

// Handler to get all todos
func getAllTodos(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	todos := make([]Todo, 0)
	for _, todo := range todoStore {
		todos = append(todos, todo)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

// Handler to get a single todo by ID
func getTodoByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	todoID := atoi(id)
	mu.RLock()
	defer mu.RUnlock()

	todo, exists := todoStore[todoID]
	if !exists {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

// Handler to create a new todo
func createTodo(w http.ResponseWriter, r *http.Request) {
	var newTodo Todo
	if err := json.NewDecoder(r.Body).Decode(&newTodo); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	newTodo.ID = rand.Intn(1000) // Random ID assignment

	mu.Lock()
	if _, ok := todoStore[newTodo.ID]; ok {
		http.Error(w, "try again, internal error", http.StatusInternalServerError)
		return
	}

	todoStore[newTodo.ID] = newTodo
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTodo)
}

// Handler to update an existing todo
func updateTodo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	todoID := atoi(id)
	var updatedTodo Todo
	if err := json.NewDecoder(r.Body).Decode(&updatedTodo); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	existingTodo, exists := todoStore[todoID]
	if !exists {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	updatedTodo.ID = existingTodo.ID // Keep the same ID
	todoStore[todoID] = updatedTodo

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTodo)
}

// Handler to delete a todo
func deleteTodo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	todoID := atoi(id)

	mu.Lock()
	defer mu.Unlock()

	if _, exists := todoStore[todoID]; !exists {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	delete(todoStore, todoID)
	w.WriteHeader(http.StatusNoContent)
}

// Helper function to convert string to int
func atoi(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return -1
	}
	return val
}
