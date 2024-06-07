package models

import (
	"database/sql"
	"fmt"
	"html/template"
	"strings"
	"time"
)

type PostsList struct {
	Posts []Post
}

type Post struct {
	ID               int
	UserID           int // Added UserID field
	CategoryID       int
	Title            string
	Content          string
	ContentHTML      template.HTML
	Slug             string
	PublicationDate  string
	LastEditDate     string
	IsPublished      bool
	FeaturedImageURL string
	CreatedAt        string
}

type PostService struct {
	DB *sql.DB
}

// Create will create a new session for the user provided. The session token
// will be returned as the Token field on the Session type, but only the hashed
// session token is stored in the database.
// func (pp *PostService) Create() (*Post, error) {

// }

func (pp *PostService) GetTopPosts() (*PostsList, error) {
	list := PostsList{}

	query := `SELECT * FROM posts LIMIT 5`
	rows, err := pp.DB.Query(query)
	if err != nil {
		return &list, nil
	}

	for rows.Next() {

		var post Post
		err := rows.Scan(&post.ID, &post.UserID, &post.CategoryID, &post.Title, &post.Content, &post.Slug, &post.PublicationDate, &post.LastEditDate, &post.IsPublished, &post.FeaturedImageURL, &post.CreatedAt)
		if err != nil {
			panic(err)
		}

		t, err := time.Parse(time.RFC3339, post.CreatedAt)
		if err != nil {
			fmt.Println(err)
		}

		post.CreatedAt = t.Format("January 2, 2006")

		post.Content = trimContent(post.Content)

		list.Posts = append(list.Posts, post)
	}

	if err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	} else {
		fmt.Println("Posts fetched successfully!")
	}

	return &list, nil
}

// Function to trim content up to the <more--> tag
func trimContent(content string) string {
	const moreTag = "<more-->"
	if idx := strings.Index(content, moreTag); idx != -1 {
		return content[:idx]
	}
	return content
}

func (pp *PostService) Create(userID int, categoryID int, title, content string, isPublished bool, featuredImageURL string, slug string) (*Post, error) {
	timefmt := time.Now()
	query := `
		INSERT INTO posts (user_id, category_id, title, content, slug, publication_date, last_edit_date, is_published, featured_image_url, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING post_id
	`
	var postID int
	println(userID, categoryID, title, content, isPublished, featuredImageURL)
	err := pp.DB.QueryRow(query, userID, categoryID, title, content, slug, timefmt,
		timefmt, isPublished, featuredImageURL, timefmt).Scan(&postID)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return nil, fmt.Errorf("create post: %w", err)
	}
	fmt.Println("Post created successfully!")
	fmt.Println(postID)

	return &Post{
		ID:               postID,
		UserID:           userID,
		CategoryID:       categoryID,
		Title:            title,
		Content:          content,
		Slug:             slug,
		PublicationDate:  timefmt.Format("January 2, 2006"),
		LastEditDate:     timefmt.Format("January 2, 2006"),
		IsPublished:      isPublished,
		FeaturedImageURL: featuredImageURL,
		CreatedAt:        timefmt.Format("January 2, 2006"),
	}, nil
}
