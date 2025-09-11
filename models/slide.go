package models

import (
	"database/sql"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type SlidesList struct {
	Slides []Slide
}

type Slide struct {
	ID              int
	UserID          int
	Title           string
	Slug            string
	ContentFilePath string
	ContentHTML     template.HTML
	IsPublished     bool
	CreatedAt       string
	UpdatedAt       string
	RelativeTime    string              // For displaying "10 months ago"
	Categories      []Category `json:"categories,omitempty"`
}

type SlideService struct {
	DB *sql.DB
}

// Create creates a new slide with the given parameters
func (ss *SlideService) Create(userID int, title, slug, content string, isPublished bool, categoryIDs []int) (*Slide, error) {
	// Generate slug if empty
	if slug == "" {
		slug = generateSlug(title)
	} else {
		slug = sanitizeSlug(slug)
	}

	// Create slide directory
	slideDir := filepath.Join("static", "slides", slug)
	if err := os.MkdirAll(slideDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create slide directory: %v", err)
	}

	// Create content file
	contentPath := filepath.Join(slideDir, "content.html")
	if err := os.WriteFile(contentPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write content file: %v", err)
	}

	// Insert slide into database
	query := `INSERT INTO Slides (user_id, title, slug, content_file_path, is_published, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) 
			  RETURNING slide_id, created_at, updated_at`
	
	var slide Slide
	err := ss.DB.QueryRow(query, userID, title, slug, contentPath, isPublished).Scan(
		&slide.ID, &slide.CreatedAt, &slide.UpdatedAt)
	if err != nil {
		// Clean up file if database insert fails
		os.RemoveAll(slideDir)
		return nil, fmt.Errorf("failed to create slide: %v", err)
	}

	slide.UserID = userID
	slide.Title = title
	slide.Slug = slug
	slide.ContentFilePath = contentPath
	slide.IsPublished = isPublished

	// Add categories if provided
	if len(categoryIDs) > 0 {
		if err := ss.AddCategories(slide.ID, categoryIDs); err != nil {
			return nil, fmt.Errorf("failed to add categories: %v", err)
		}
	}

	return &slide, nil
}

// GetBySlug retrieves a slide by its slug
func (ss *SlideService) GetBySlug(slug string) (*Slide, error) {
	query := `SELECT slide_id, user_id, title, slug, content_file_path, is_published, created_at, updated_at 
			  FROM Slides WHERE slug = $1`
	
	var slide Slide
	err := ss.DB.QueryRow(query, slug).Scan(
		&slide.ID, &slide.UserID, &slide.Title, &slide.Slug, &slide.ContentFilePath,
		&slide.IsPublished, &slide.CreatedAt, &slide.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("slide not found")
		}
		return nil, fmt.Errorf("failed to get slide: %v", err)
	}

	// Load content from file
	if err := ss.loadSlideContent(&slide); err != nil {
		return nil, err
	}

	// Load categories
	if err := ss.loadSlideCategories(&slide); err != nil {
		return nil, err
	}

	return &slide, nil
}

// GetByID retrieves a slide by its ID
func (ss *SlideService) GetByID(id int) (*Slide, error) {
	query := `SELECT slide_id, user_id, title, slug, content_file_path, is_published, created_at, updated_at 
			  FROM Slides WHERE slide_id = $1`
	
	var slide Slide
	err := ss.DB.QueryRow(query, id).Scan(
		&slide.ID, &slide.UserID, &slide.Title, &slide.Slug, &slide.ContentFilePath,
		&slide.IsPublished, &slide.CreatedAt, &slide.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("slide not found")
		}
		return nil, fmt.Errorf("failed to get slide: %v", err)
	}

	// Load content from file
	if err := ss.loadSlideContent(&slide); err != nil {
		return nil, err
	}

	// Load categories
	if err := ss.loadSlideCategories(&slide); err != nil {
		return nil, err
	}

	return &slide, nil
}

// GetPublishedSlides retrieves all published slides
func (ss *SlideService) GetPublishedSlides() (*SlidesList, error) {
	list := SlidesList{}
	
	query := `SELECT slide_id, user_id, title, slug, content_file_path, is_published, created_at, updated_at 
			  FROM Slides WHERE is_published = true ORDER BY created_at DESC`
	
	rows, err := ss.DB.Query(query)
	if err != nil {
		return &list, fmt.Errorf("failed to query slides: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var slide Slide
		err := rows.Scan(&slide.ID, &slide.UserID, &slide.Title, &slide.Slug, &slide.ContentFilePath,
			&slide.IsPublished, &slide.CreatedAt, &slide.UpdatedAt)
		if err != nil {
			return &list, fmt.Errorf("failed to scan slide: %v", err)
		}

		// Load categories for each slide
		if err := ss.loadSlideCategories(&slide); err != nil {
			return &list, err
		}

		list.Slides = append(list.Slides, slide)
	}

	return &list, nil
}

// GetAllSlides retrieves all slides (for admin)
func (ss *SlideService) GetAllSlides() (*SlidesList, error) {
	list := SlidesList{}
	
	query := `SELECT slide_id, user_id, title, slug, content_file_path, is_published, created_at, updated_at 
			  FROM Slides ORDER BY created_at DESC`
	
	rows, err := ss.DB.Query(query)
	if err != nil {
		return &list, fmt.Errorf("failed to query slides: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var slide Slide
		err := rows.Scan(&slide.ID, &slide.UserID, &slide.Title, &slide.Slug, &slide.ContentFilePath,
			&slide.IsPublished, &slide.CreatedAt, &slide.UpdatedAt)
		if err != nil {
			return &list, fmt.Errorf("failed to scan slide: %v", err)
		}

		// Load categories for each slide
		if err := ss.loadSlideCategories(&slide); err != nil {
			return &list, err
		}

		list.Slides = append(list.Slides, slide)
	}

	return &list, nil
}

// Update updates an existing slide
func (ss *SlideService) Update(slideID int, title, slug, content string, isPublished bool, categoryIDs []int) error {
	// Get current slide to access file path
	currentSlide, err := ss.GetByID(slideID)
	if err != nil {
		return err
	}

	// Sanitize slug
	if slug == "" {
		slug = generateSlug(title)
	} else {
		slug = sanitizeSlug(slug)
	}

	// Update content file
	if err := os.WriteFile(currentSlide.ContentFilePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to update content file: %v", err)
	}

	// Update database record
	query := `UPDATE Slides SET title = $1, slug = $2, is_published = $3, updated_at = CURRENT_TIMESTAMP 
			  WHERE slide_id = $4`
	
	_, err = ss.DB.Exec(query, title, slug, isPublished, slideID)
	if err != nil {
		return fmt.Errorf("failed to update slide: %v", err)
	}

	// Update categories
	if err := ss.UpdateCategories(slideID, categoryIDs); err != nil {
		return fmt.Errorf("failed to update categories: %v", err)
	}

	return nil
}

// Delete deletes a slide and its associated files
func (ss *SlideService) Delete(slideID int) error {
	// Get slide to access file path
	slide, err := ss.GetByID(slideID)
	if err != nil {
		return err
	}

	// Delete from database first
	query := `DELETE FROM Slides WHERE slide_id = $1`
	_, err = ss.DB.Exec(query, slideID)
	if err != nil {
		return fmt.Errorf("failed to delete slide from database: %v", err)
	}

	// Remove slide directory
	slideDir := filepath.Dir(slide.ContentFilePath)
	if err := os.RemoveAll(slideDir); err != nil {
		return fmt.Errorf("failed to remove slide directory: %v", err)
	}

	return nil
}

// AddCategories adds categories to a slide
func (ss *SlideService) AddCategories(slideID int, categoryIDs []int) error {
	for _, categoryID := range categoryIDs {
		query := `INSERT INTO Slide_Categories (slide_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
		_, err := ss.DB.Exec(query, slideID, categoryID)
		if err != nil {
			return fmt.Errorf("failed to add category %d: %v", categoryID, err)
		}
	}
	return nil
}

// UpdateCategories updates the categories associated with a slide
func (ss *SlideService) UpdateCategories(slideID int, categoryIDs []int) error {
	// Remove existing categories
	deleteQuery := `DELETE FROM Slide_Categories WHERE slide_id = $1`
	_, err := ss.DB.Exec(deleteQuery, slideID)
	if err != nil {
		return fmt.Errorf("failed to remove existing categories: %v", err)
	}

	// Add new categories
	if len(categoryIDs) > 0 {
		return ss.AddCategories(slideID, categoryIDs)
	}

	return nil
}

// loadSlideContent loads the HTML content from file
func (ss *SlideService) loadSlideContent(slide *Slide) error {
	content, err := os.ReadFile(slide.ContentFilePath)
	if err != nil {
		return fmt.Errorf("failed to read content file: %v", err)
	}
	slide.ContentHTML = template.HTML(content)
	return nil
}

// loadSlideCategories loads the categories for a slide
func (ss *SlideService) loadSlideCategories(slide *Slide) error {
	query := `SELECT c.category_id, c.category_name, c.created_at 
			  FROM Categories c 
			  JOIN Slide_Categories sc ON c.category_id = sc.category_id 
			  WHERE sc.slide_id = $1`
	
	rows, err := ss.DB.Query(query, slide.ID)
	if err != nil {
		return fmt.Errorf("failed to load categories: %v", err)
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		err := rows.Scan(&category.ID, &category.Name, &category.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan category: %v", err)
		}
		categories = append(categories, category)
	}

	slide.Categories = categories
	return nil
}

// Helper functions
func generateSlug(title string) string {
	// Convert to lowercase and replace spaces with hyphens
	slug := strings.ToLower(title)
	slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	
	// Ensure slug is not empty
	if slug == "" {
		slug = fmt.Sprintf("slide-%d", time.Now().Unix())
	}
	
	return slug
}

func sanitizeSlug(slug string) string {
	// Convert to lowercase and sanitize
	slug = strings.ToLower(slug)
	slug = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	
	// Ensure slug is not empty
	if slug == "" {
		slug = fmt.Sprintf("slide-%d", time.Now().Unix())
	}
	
	return slug
}