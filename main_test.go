package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

func TestGetAllTodos(t *testing.T) {
	// Seed data
	mu.Lock()
	t1 := Todo{ID: 1, Title: "Test Todo 1", Description: "Test Description 1", Completed: false}
	t2 := Todo{ID: 2, Title: "Test Todo 2", Description: "Test Description 2", Completed: false}
	todoStore[1] = t1
	todoStore[2] = t2
	mu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/todos", nil)
	rec := httptest.NewRecorder()

	getAllTodos(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	var todos []Todo
	if err := json.NewDecoder(rec.Body).Decode(&todos); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !(slices.Equal(todos, []Todo{t1, t2}) || slices.Equal(todos, []Todo{t2, t1})) {
		t.Errorf("Expected todos are not returned")
	}
}

func TestGetTodoByID(t *testing.T) {
	// Seed data
	mu.Lock()
	todoStore[1] = Todo{ID: 1, Title: "Test Todo", Description: "Test Description", Completed: false}
	mu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/todos/1", nil)
	rec := httptest.NewRecorder()

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", "1")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
	getTodoByID(rec, req.WithContext(ctx))

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	var todo Todo
	if err := json.NewDecoder(rec.Body).Decode(&todo); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if todo.ID != 1 {
		t.Errorf("Expected todo ID 1, got %d", todo.ID)
	}
}

func TestCreateTodo(t *testing.T) {
	todo := Todo{Title: "New Todo", Description: "New Description", Completed: false}
	body, _ := json.Marshal(todo)

	req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	createTodo(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, rec.Code)
	}

	var createdTodo Todo
	if err := json.NewDecoder(rec.Body).Decode(&createdTodo); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if createdTodo.Title != todo.Title {
		t.Errorf("Expected title %q, got %q", todo.Title, createdTodo.Title)
	}
}

func TestUpdateTodo(t *testing.T) {
	// Seed data
	mu.Lock()
	todoStore[1] = Todo{ID: 1, Title: "Test Todo", Description: "Test Description", Completed: false}
	mu.Unlock()

	updatedTodo := Todo{Title: "Updated Todo", Description: "Updated Description", Completed: true}
	body, _ := json.Marshal(updatedTodo)

	req := httptest.NewRequest(http.MethodPut, "/todos/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", "1")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
	updateTodo(rec, req.WithContext(ctx))

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	var todo Todo
	if err := json.NewDecoder(rec.Body).Decode(&todo); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if todo.Title != updatedTodo.Title {
		t.Errorf("Expected title %q, got %q", updatedTodo.Title, todo.Title)
	}
}

func TestDeleteTodo(t *testing.T) {
	// Seed data
	mu.Lock()
	todoStore[1] = Todo{ID: 1, Title: "Test Todo", Description: "Test Description", Completed: false}
	mu.Unlock()

	req := httptest.NewRequest(http.MethodDelete, "/todos/1", nil)
	rec := httptest.NewRecorder()

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", "1")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
	deleteTodo(rec, req.WithContext(ctx))

	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, rec.Code)
	}

	mu.Lock()
	if _, exists := todoStore[1]; exists {
		t.Errorf("Expected todo to be deleted")
	}
	mu.Unlock()
}
