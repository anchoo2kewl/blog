package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"anshumanbiswas.com/blog/models"
	"anshumanbiswas.com/blog/utils"
	"anshumanbiswas.com/blog/views"
	"github.com/go-chi/chi/v5"
)

type Categories struct {
	CategoryService *models.CategoryService
	SessionService  *models.SessionService
	Templates       struct {
		Manage views.Template
	}
}

// Admin Category Management Page
func (c *Categories) Manage(w http.ResponseWriter, r *http.Request) {
	user, err := utils.IsUserLoggedIn(r, c.SessionService)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	// Check if user is admin
	if !models.IsAdmin(user.Role) {
		http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
		return
	}

	// Get all categories
	categories, err := c.CategoryService.GetAll()
	if err != nil {
		log.Printf("Error getting categories: %v", err)
		http.Error(w, "Failed to load categories", http.StatusInternalServerError)
		return
	}

	// Get post counts for each category
	postCounts, err := c.CategoryService.GetPostCountByCategory()
	if err != nil {
		log.Printf("Error getting post counts: %v", err)
		postCounts = make(map[int]int) // Fallback to empty map
	}

	data := struct {
		Email           string
		LoggedIn        bool
		Username        string
		IsAdmin         bool
		SignupDisabled  bool
		Description     string
		CurrentPage     string
		User            *models.User
		Categories      []models.Category
		PostCounts      map[int]int
		Flash           string
		UserPermissions models.UserPermissions
	}{
		Email:           user.Email,
		LoggedIn:        true,
		Username:        user.Username,
		IsAdmin:         models.IsAdmin(user.Role),
		SignupDisabled:  true, // Default for admin pages
		Description:     "Manage Categories - Anshuman Biswas Blog",
		CurrentPage:     "admin-categories",
		User:            user,
		Categories:      categories,
		PostCounts:      postCounts,
		Flash:           "",
		UserPermissions: models.GetPermissions(user.Role),
	}

	// Check for flash messages
	if msg := r.URL.Query().Get("message"); msg != "" {
		data.Flash = msg
	}

	c.Templates.Manage.Execute(w, r, data)
}

// REST API Endpoints

// ListCategories - GET /api/categories
func (c *Categories) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := c.CategoryService.GetAll()
	if err != nil {
		log.Printf("Error getting categories: %v", err)
		http.Error(w, "Failed to get categories", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// GetCategory - GET /api/categories/{id}
func (c *Categories) GetCategory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	category, err := c.CategoryService.GetByID(id)
	if err != nil {
		log.Printf("Error getting category: %v", err)
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

// CreateCategory - POST /api/categories
func (c *Categories) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	category, err := c.CategoryService.Create(req.Name)
	if err != nil {
		log.Printf("Error creating category: %v", err)
		http.Error(w, fmt.Sprintf("Failed to create category: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}

// UpdateCategory - PUT /api/categories/{id}
func (c *Categories) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	category, err := c.CategoryService.Update(id, req.Name)
	if err != nil {
		log.Printf("Error updating category: %v", err)
		if err.Error() == "category not found" {
			http.Error(w, "Category not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to update category: %v", err), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

// DeleteCategory - DELETE /api/categories/{id}
func (c *Categories) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	err = c.CategoryService.Delete(id)
	if err != nil {
		log.Printf("Error deleting category: %v", err)
		if err.Error() == "category not found" {
			http.Error(w, "Category not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to delete category: %v", err), http.StatusBadRequest)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Form-based endpoints for web interface

// CreateCategoryForm - POST /admin/categories
func (c *Categories) CreateCategoryForm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.Form.Get("name")
	if name == "" {
		http.Redirect(w, r, "/admin/categories?message=Category+name+is+required", http.StatusFound)
		return
	}

	_, err := c.CategoryService.Create(name)
	if err != nil {
		log.Printf("Error creating category: %v", err)
		http.Redirect(w, r, "/admin/categories?message=Failed+to+create+category", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/admin/categories?message=Category+created+successfully", http.StatusFound)
}

// UpdateCategoryForm - POST /admin/categories/{id}
func (c *Categories) UpdateCategoryForm(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.Form.Get("name")
	if name == "" {
		http.Redirect(w, r, "/admin/categories?message=Category+name+is+required", http.StatusFound)
		return
	}

	_, err = c.CategoryService.Update(id, name)
	if err != nil {
		log.Printf("Error updating category: %v", err)
		http.Redirect(w, r, "/admin/categories?message=Failed+to+update+category", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/admin/categories?message=Category+updated+successfully", http.StatusFound)
}

// DeleteCategoryForm - POST /admin/categories/{id}/delete
func (c *Categories) DeleteCategoryForm(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	err = c.CategoryService.Delete(id)
	if err != nil {
		log.Printf("Error deleting category: %v", err)
		http.Redirect(w, r, "/admin/categories?message=Failed+to+delete+category", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/admin/categories?message=Category+deleted+successfully", http.StatusFound)
}