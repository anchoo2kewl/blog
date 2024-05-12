package models

import (
	"database/sql"
	"fmt"
	"time"
)

type BlogService struct {
	DB *sql.DB
}

func (bs *BlogService) GetBlogPostBySlug(slug string) (*Post, error) {

	post := Post{}

	fmt.Println("Fetching blog post with slug:", slug)

	query := `SELECT * FROM posts WHERE slug = $1 LIMIT 1`
	rows, err := bs.DB.Query(query, slug)
	if err != nil {
		return &post, nil
	}

	for rows.Next() {

		err := rows.Scan(&post.ID, &post.UserID, &post.CategoryID, &post.Title, &post.Content, &post.Slug, &post.PublicationDate, &post.LastEditDate, &post.IsPublished, &post.FeaturedImageURL, &post.CreatedAt)
		if err != nil {
			panic(err)
		}

		t, err := time.Parse(time.RFC3339, post.CreatedAt)
		if err != nil {
			fmt.Println(err)
		}

		post.CreatedAt = t.Format("January 2, 2006")

	}

	if err != nil {
		return nil, fmt.Errorf("Post could not be fetched: %w", err)
	} else {
		fmt.Println("Posts fetched successfully!")
	}

	fmt.Println("Blog Post:", post)

	return &post, nil
}
