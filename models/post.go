package models

import (
	"database/sql"
	"fmt"
	"time"
)

type PostsList struct {
	Posts []Post
}

type Post struct {
	ID               int
	userID           int
	CategoryID       int
	Title            string
	Content          string
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
		err := rows.Scan(&post.ID, &post.userID, &post.CategoryID, &post.Title, &post.Content, &post.PublicationDate, &post.LastEditDate, &post.IsPublished, &post.FeaturedImageURL, &post.CreatedAt)
		if err != nil {
			panic(err)
		}

		t, err := time.Parse(time.RFC3339, post.CreatedAt)
		if err != nil {
			fmt.Println(err)
		}

		post.CreatedAt = t.Format("January 2, 2006")

		list.Posts = append(list.Posts, post)
	}

	if err != nil {
		fmt.Errorf("Posts cannot be fetched! %w", err)
	}
	return &list, nil
}
